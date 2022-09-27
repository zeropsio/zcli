package serviceLogs

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeropsio/zcli/src/i18n"
	"os"
	"os/signal"
	"strings"
	"time"
)

func (h *Handler) getLogStream(
	ctx context.Context, config RunConfig, inputs InputValues, uri, query, containerId, logServiceId, projectId string,
) error {
	url := h.updateUri(uri, query)

	interrupt := make(chan os.Signal, 1)   // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
	}
	defer conn.Close()

	done := make(chan interface{}) // Channel to indicate that the receiverHandler is done

	go h.receiveHandler(conn, inputs.format, done)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-done:
			// if interrupted by user
			if ctx.Err() != nil {
				return nil
			}
			// otherwise try to reconnect the websocket
			err := h.printLogs(ctx, config, inputs, containerId, logServiceId, projectId)
			if err != nil {
				return err
			}
		// received a SIGINT (Ctrl + C). Terminate gracefully...
		case <-interrupt:
			// Close the websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}

			select {
			case <-done:
			case <-time.After(time.Duration(1) * time.Second):
				return nil
			}
		}
	}
}

// check last message id, add it to `from` and update the ws url for reconnect
func (h *Handler) updateUri(uri, query string) string {
	from := ""
	if h.LastMsgId != "" {
		from = fmt.Sprintf("&from=%s", h.LastMsgId)
	}
	return WSS + uri + query + from
}

func (h *Handler) receiveHandler(connection *websocket.Conn, format string, done chan interface{}) {
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
				errMsg := fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
				fmt.Println(errMsg)
			}
			return
		}

		h.printStreamLog(msg, format)
	}
}

func (h *Handler) printStreamLog(data []byte, format string) {
	jsonData, _ := parseResponse(data)
	// only if there is a new message coming
	if len(jsonData.Items) > 0 {
		//update last msg ID for ws reconnection
		h.LastMsgId = jsonData.Items[len(jsonData.Items)-1].Id
		err := parseResponseByFormat(jsonData, format, "", STREAM)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
