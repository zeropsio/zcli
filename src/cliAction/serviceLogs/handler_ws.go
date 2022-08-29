package serviceLogs

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

var done chan interface{}
var interrupt chan os.Signal

func getLogStream(ctx context.Context, expiration types.DateTime, format, serviceId, uri, path string) error {
	if format == JSON {
		return fmt.Errorf("%s", i18n.LogFormatStreamMismatch)
	}

	fmt.Printf("stream with:\n expiration %s\n for serviceId %s\n in format %s\n url is %s\n", expiration, serviceId, format, uri+path)
	// todo add websocket
	// compare expiration time with time now
	u := url.URL{Scheme: "wss", Host: uri + path, Path: ""}
	log.Printf("connecting to %s", u.String())
	err := listenWS(ctx, u)
	if err != nil {
		return err
	}

	return nil
}

func listenWS(_ context.Context, url url.URL) error {
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT
	urlFixed1 := strings.ReplaceAll(url.String(), "%2F", "/")
	urlFixed := strings.ReplaceAll(urlFixed1, "%3F", "?")
	conn, _, err := websocket.DefaultDialer.Dial(urlFixed, nil)
	if err != nil {
		return fmt.Errorf("error connecting to Websocket Server: %v", err)
	}
	defer conn.Close()
	go func() {
		err := receiveHandler(conn)
		if err != nil {
			fmt.Println("problem kamo ", err)
		}
	}()

	// Our main loop for the client
	// We send our relevant packets here
	for {
		fmt.Println("Hi")
		select {
		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return nil
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return nil
		}
	}
}

func receiveHandler(connection *websocket.Conn) error {
	defer close(done)
	for {
		fmt.Println("hi")
		_, msg, err := connection.ReadMessage()
		if err != nil {
			return fmt.Errorf("Log reading failed. %s\n", err)
		}
		log.Println(msg)
	}
}
