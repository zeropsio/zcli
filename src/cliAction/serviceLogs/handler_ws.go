package serviceLogs

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types"
	"os"
	"os/signal"
	"strings"
	"time"
)

var done chan interface{}
var interrupt chan os.Signal

func getLogStream(ctx context.Context, expiration types.DateTime, format, uri, path, mode string) error {
	// TODO move this validation to the beginning
	if format == JSON {
		return fmt.Errorf("%s", i18n.LogFormatStreamMismatch)
	}
	fmt.Printf("expiration: %s\n", expiration)
	// todo compare expiration time with time now
	url := WSS + uri + path
	if err := listenWS(ctx, url, format, mode); err != nil {
		return err
	}

	return nil
}

func listenWS(_ context.Context, url, format, mode string) error {
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
	}
	defer conn.Close()

	go receiveHandler(conn, format, mode)
	// TODO count expiration
	expired := true
	expired = false

	// main loop
	for !expired {
		select {
		// received a SIGINT (Ctrl + C). Terminate gracefully...
		case <-interrupt:
			//log.Println("Received SIGINT interrupt signal. Closing all pending connections...")
			// Close the  websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}

			select {
			case <-done:
				return nil
			case <-time.After(time.Duration(1) * time.Second):
				return nil
			}
		}
	}
	return nil
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
			err := parseResponseByFormat(msg, format, "", mode)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
