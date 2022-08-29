package serviceLogs

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types"
	"log"
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

	// main loop
	for {
		select {
		case <-interrupt:
			// received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections...")

			// Close the  websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return nil
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Receiver Channel Closed! Exiting....")
			}
			return nil
		}
	}
}

func receiveHandler(connection *websocket.Conn, format, mode string) {
	defer close(done)
	defer close(interrupt)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			if !strings.Contains(string(msg), "use of closed network connection") {
				errMsg := fmt.Errorf("%s %s\n", i18n.LogReadingFailed, err.Error())
				fmt.Println(errMsg)
			}
		}
		if !strings.Contains(string(msg), "{\"items\":[]}") {
			err := parseResponseByFormat(msg, format, "", mode)
			if err != nil {
				fmt.Println(err.Error())
			}

		}
	}
}
