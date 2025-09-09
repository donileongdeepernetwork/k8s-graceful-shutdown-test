package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

func connectAndRun() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "ws://localhost:8080/ws"
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		log.Println("Parse URL error:", err)
		return
	}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			log.Printf("Received: %s", message)

			if string(message) == "SHUTDOWN" {
				log.Println("Server is shutting down, processing remaining data...")
			} else if string(message) == "CLOSE" {
				log.Println("Server sent close signal")
				return
			}
		}
	}()

	// Simulate sending messages
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			log.Println("Connection closed, will reconnect...")
			return
		case <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte("Client message: "+time.Now().Format("15:04:05")))
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		connectAndRun()

		select {
		case <-interrupt:
			log.Println("Client interrupted")
			return
		default:
			log.Println("Reconnecting in 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}
}
