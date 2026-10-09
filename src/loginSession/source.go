// Package loginSession keeps the access token of a browser login fresh.
package loginSession

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
)

const (
	refreshMargin  = 5 * time.Minute
	refreshTimeout = 20 * time.Second
)

var errNoRefreshToken = errors.New("auth response has no refresh token")

type RefreshApi interface {
	PostAuthRefresh(ctx context.Context, in body.AuthRefresh) (sdk.PostAuthRefreshResponse, error)
}

// Source is a zeropsRestApiClient.TokenSource for the LoginSession in cliStorage.
// Zerops deletes the old token pair on refresh, so every refresh runs under the cross-process storage lock and first checks whether another process already refreshed.
type Source struct {
	storage *cliStorage.Handler
	api     RefreshApi
	now     func() time.Time

	lock    sync.Mutex
	session cliStorage.LoginSession
	expired bool
}

func New(storage *cliStorage.Handler, api RefreshApi) *Source {
	return &Source{
		storage: storage,
		api:     api,
		now:     time.Now,
		session: storage.Data().LoginSession,
	}
}

// FromAuth converts the tokens of a login or a refresh into a LoginSession.
func FromAuth(auth output.AuthFull) (cliStorage.LoginSession, error) {
	refreshToken, ok := auth.RefreshToken.Get()
	if !ok || refreshToken.Native() == "" {
		return cliStorage.LoginSession{}, errNoRefreshToken
	}
	return cliStorage.LoginSession{
		AccessToken:  auth.AccessToken.Native(),
		RefreshToken: refreshToken.Native(),
		ExpiresAt:    auth.ExpiresAt.Native(),
	}, nil
}

// Expired reports that Zerops rejected the refresh token and the stored session was cleared.
func (s *Source) Expired() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.expired
}

// Token refreshes ahead of expiry.
func (s *Source) Token(ctx context.Context) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.refreshAhead(ctx, refreshMargin)
	return s.session.AccessToken, nil
}

// RefreshAhead refreshes the access token if it expires within margin.
func (s *Source) RefreshAhead(ctx context.Context, margin time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.refreshAhead(ctx, margin)
}

func (s *Source) refreshAhead(ctx context.Context, margin time.Duration) {
	if !s.expired && s.expiresWithin(s.session, margin) {
		// If the refresh fails, the request goes out with the current token and a 401 leads to Renew.
		_ = s.refresh(ctx, s.session.AccessToken, margin)
	}
}

func (s *Source) Renew(ctx context.Context, refused string) (string, bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.expired {
		return "", false, nil
	}
	if err := s.refresh(ctx, refused, refreshMargin); err != nil {
		return "", false, err
	}
	if s.expired {
		return "", false, nil
	}
	return s.session.AccessToken, true, nil
}

func (s *Source) expiresWithin(session cliStorage.LoginSession, margin time.Duration) bool {
	return session.ExpiresAt.Sub(s.now()) < margin
}

// refresh replaces the access token that the caller holds as current.
// A stored token that differs from current comes from another process, and it is used as is unless it expires within margin too.
func (s *Source) refresh(ctx context.Context, current string, margin time.Duration) error {
	data, err := s.storage.Modify(func(data cliStorage.Data) (cliStorage.Data, error) {
		stored := data.LoginSession
		if !stored.IsSet() {
			return data, nil
		}
		if stored.AccessToken != current && !s.expiresWithin(stored, margin) {
			return data, nil
		}

		// Without cancellation, because a canceled call may still rotate the tokens on the server and the new pair would be lost.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), refreshTimeout)
		defer cancel()
		resp, err := s.api.PostAuthRefresh(ctx, body.AuthRefresh{RefreshTokenId: types.NewString(stored.RefreshToken)})
		if err != nil {
			return data, errors.Wrap(err, "refresh login session")
		}
		auth, err := resp.Output()
		if apiError.HasErrorCode(err, errorCode.RefreshTokenInvalid) || apiError.HasErrorCode(err, errorCode.UserNotFound) {
			data.LoginSession = cliStorage.LoginSession{}
			return data, nil
		}
		if err != nil {
			return data, err
		}
		data.LoginSession, err = FromAuth(auth)
		return data, err
	})
	if err != nil {
		return err
	}
	s.session = data.LoginSession
	s.expired = !s.session.IsSet()
	return nil
}
