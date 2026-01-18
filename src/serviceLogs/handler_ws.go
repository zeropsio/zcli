package serviceLogs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

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
		return errors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
	}
	defer reps.Body.Close()
	defer conn.Close()

	done := make(chan error) // Channel to indicate that the receiverHandler is done and send error
	go h.receiveHandler(conn, inputs.format, inputs.formatTemplate, done)

	for {
		select {
		case <-ctx.Done():
			// Close the websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}
		case err := <-done:
			return err
		}
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
			// on token expiration
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || websocket.IsUnexpectedCloseError(err) {
				time.Sleep(time.Second * 5)
				return
			}
			finishedByUser := strings.Contains(err.Error(), "use of closed network connection")
			if !finishedByUser {
				done <- errors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
			}
			return
		}

		if err := h.printStreamLog(msg, format, formatTemplate); err != nil {
			done <- err
			return
		}
	}
}

func (h *Handler) printStreamLog(data []byte, format, formatTemplate string) error {
	jsonData, _ := parseResponse(data)
	// only if there is a new message coming
	if len(jsonData.Items) > 0 {
		// update last msg ID for ws reconnection
		h.lastMsgId = jsonData.Items[len(jsonData.Items)-1].Id
		if err := h.parseResponseByFormat(jsonData, format, formatTemplate, STREAM); err != nil {
			return err
		}
	}
	return nil
}
