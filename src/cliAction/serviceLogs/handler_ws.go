package serviceLogs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zerops-io/zcli/src/i18n"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	done      chan interface{}
	interrupt chan os.Signal
	lastMsgId string
)

func (h *Handler) getLogStream(
	ctx context.Context, inputs InputValues, uri, query, containerId, logServiceId, projectId string,
) error {
	url := updateUri(uri, query)

	interrupt = make(chan os.Signal, 1)    // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
	}
	defer conn.Close()

	done = make(chan interface{}) // Channel to indicate that the receiverHandler is done

	go h.receiveHandler(conn, inputs.format, inputs.mode)

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
			err := h.printLogs(ctx, inputs, containerId, logServiceId, projectId)
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
func updateUri(uri, query string) string {
	from := ""
	if lastMsgId != "" {
		from = fmt.Sprintf("&from=%s", lastMsgId)
	}
	return WSS + uri + query + from
}

func (h *Handler) receiveHandler(connection *websocket.Conn, format, mode string) {
	defer close(done)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			// websocket close err (appears on expiration of token)
			closeErr := strings.Contains(err.Error(), "websocket: close")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || closeErr {
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

		if strings.Contains(string(msg), "{\"items\"") && !strings.Contains(string(msg), "{\"items\":[]}") {
			lastMsgId = updateLastMsgId(msg)
			err := parseResponseByFormat(msg, format, "", mode)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

// update last msg ID for ws reconnection, but only if there is a new message coming
func updateLastMsgId(body []byte) string {
	var jsonData Response
	err := json.Unmarshal(body, &jsonData)
	if err != nil || len(jsonData.Items) == 0 {
		return lastMsgId
	}
	return jsonData.Items[len(jsonData.Items)-1].Id
}
