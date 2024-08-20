package serviceLogs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) getLogStream(
	ctx context.Context,
	inputs InputValues,
	projectId uuid.ProjectId,
	serviceId uuid.ServiceStackId,
	containerId uuid.ContainerId,
	uri, query string,
) error {
	url := h.updateUri(uri, query)

	interrupt := make(chan os.Signal, 1)   // Channel to listen for interrupt signal to terminate gracefully
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	conn, reps, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return errors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
	}
	defer reps.Body.Close()
	defer conn.Close()

	done := make(chan interface{}) // Channel to indicate that the receiverHandler is done

	go h.receiveHandler(conn, inputs.format, inputs.formatTemplate, done)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-done:
			// if interrupted by user
			if ctx.Err() != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return nil
				}
				return err
			}
			// otherwise try to reconnect the websocket
			err := h.printLogs(ctx, inputs, projectId, serviceId, containerId)
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
	if h.lastMsgId != "" {
		from = fmt.Sprintf("&from=%s", h.lastMsgId)
	}
	return WSS + uri + query + from
}

func (h *Handler) receiveHandler(connection *websocket.Conn, format, formatTemplate string, done chan interface{}) {
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
				errMsg := errors.Errorf("%s %s\n", i18n.T(i18n.LogReadingFailed), err.Error())
				fmt.Println(errMsg)
			}
			return
		}

		h.printStreamLog(msg, format, formatTemplate)
	}
}

func (h *Handler) printStreamLog(data []byte, format, formatTemplate string) {
	jsonData, _ := parseResponse(data)
	// only if there is a new message coming
	if len(jsonData.Items) > 0 {
		// update last msg ID for ws reconnection
		h.lastMsgId = jsonData.Items[len(jsonData.Items)-1].Id
		err := parseResponseByFormat(jsonData, format, formatTemplate, STREAM)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
