package serviceLogs

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gorilla/websocket"
	pkgerrors "github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) getLogStream(
	ctx context.Context,
	inputs InputValues,
	uri, query string,
) error {
	url := h.updateUri(uri, query)

	conn, reps, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return pkgerrors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
	}
	defer reps.Body.Close()
	defer conn.Close()

	done := make(chan error) // Channel to indicate that the receiverHandler is done and send error
	go h.receiveHandler(conn, inputs.format, inputs.formatTemplate, done)

	select {
	case <-ctx.Done():
		// Send a normal-close frame so the server tears down cleanly, then
		// return the context error. Without the return the previous
		// `for { select{} }` loop would re-fire ctx.Done() forever.
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// check last message id, add it to `from` and update the ws url for reconnect
func (h *Handler) updateUri(uri, query string) string {
	from := ""
	if h.lastMsgId != "" {
		from = fmt.Sprintf("&from=%s", h.lastMsgId)
	}
	return uri + query + from
}

func (h *Handler) receiveHandler(connection *websocket.Conn, format, formatTemplate string, done chan error) {
	defer close(done)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			// token-expiration / remote close: let the outer caller decide
			// whether to reconnect; just exit the read loop cleanly.
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || websocket.IsUnexpectedCloseError(err) {
				return
			}
			// The connection was closed locally (e.g. ctx cancel triggered
			// our deferred conn.Close in getLogStream). Don't surface that
			// as a reading-failed error.
			if errors.Is(err, net.ErrClosed) {
				return
			}
			done <- pkgerrors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
			return
		}

		if err := h.printStreamLog(msg, format, formatTemplate); err != nil {
			done <- err
			return
		}
	}
}

func (h *Handler) printStreamLog(data []byte, format, formatTemplate string) error {
	jsonData, err := parseResponse(data)
	if err != nil {
		return err
	}
	if len(jsonData.Items) == 0 {
		return nil
	}
	// update last msg ID for ws reconnection
	h.lastMsgId = jsonData.Items[len(jsonData.Items)-1].Id
	return h.parseResponseByFormat(jsonData, format, formatTemplate, STREAM)
}
