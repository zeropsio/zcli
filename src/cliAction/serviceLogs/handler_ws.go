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

var done chan interface{}
var interrupt chan os.Signal
var lastMsgId string

func (h *Handler) getLogStream(ctx context.Context, format, uri, query, mode string) error {
	url := updateUri(uri, query)
	interrupt = make(chan os.Signal, 1)    // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
	}
	defer conn.Close()

	done = make(chan interface{}) // Channel to indicate that the receiverHandler is done

	go receiveHandler(conn, format, mode)

	for {
		select {
		case <-done:
			fmt.Println("done")
			return nil
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
		case <-ctx.Done():
			fmt.Println("ctx done")
			return ctx.Err() // dunno what to do with this
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

func receiveHandler(connection *websocket.Conn, format, mode string) {
	defer close(done)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			finishedByUser := strings.Contains(err.Error(), "use of closed network connection")
			if !finishedByUser {
				errMsg := fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
				fmt.Println(errMsg)
			}
			return
		}

		if !strings.Contains(string(msg), "{\"items\":[]}") {
			lastMsgId = getLastMsgId(msg) // update last message id for reconnection
			err := parseResponseByFormat(msg, format, "", mode)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func getLastMsgId(body []byte) string {
	var jsonData Response
	err := json.Unmarshal(body, &jsonData)
	if err != nil {
		return ""
	}
	return jsonData.Items[len(jsonData.Items)-1].Id
}
