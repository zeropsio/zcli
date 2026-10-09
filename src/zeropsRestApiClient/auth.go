package zeropsRestApiClient

import (
	"context"
	"net/http"
)

type TokenSource interface {
	// Token returns the token for the next request.
	Token(ctx context.Context) (string, error)
	// Renew is called after the API refused the token with 401.
	// It returns false when there is no other token to retry with, and the client then returns the 401 response.
	Renew(ctx context.Context, refused string) (string, bool, error)
}

type StaticToken string

func (t StaticToken) Token(context.Context) (string, error) {
	return string(t), nil
}

func (t StaticToken) Renew(context.Context, string) (string, bool, error) {
	return "", false, nil
}

type authTransport struct {
	base   http.RoundTripper
	source TokenSource
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	token, err := t.source.Token(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := t.base.RoundTrip(withToken(req, token))
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}
	// A request with a body that cannot be read again cannot be retried.
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return resp, nil
	}

	renewed, ok, err := t.source.Renew(ctx, token)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	if !ok {
		return resp, nil
	}
	retry := withToken(req, renewed)
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		retry.Body = body
	}
	resp.Body.Close()
	return t.base.RoundTrip(retry)
}

// withToken returns a copy of req, because a RoundTripper must not modify the request it was given.
func withToken(req *http.Request, token string) *http.Request {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+token)
	return clone
}
