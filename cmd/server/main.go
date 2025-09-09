package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

var clients = make(map[*Client]bool)
var clientsMutex sync.Mutex

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected: %s", conn.RemoteAddr().String())

	client := &Client{conn: conn, send: make(chan []byte, 256)}
	clientsMutex.Lock()
	clients[client] = true
	clientsMutex.Unlock()

	go client.writePump()
	go client.readPump()

	// Simulate business interaction
	for {
		select {
		case <-time.After(time.Second * 2):
			message := []byte("Business data: " + time.Now().Format("15:04:05"))
			select {
			case client.send <- message:
			default:
				close(client.send)
				clientsMutex.Lock()
				delete(clients, client)
				clientsMutex.Unlock()
				return
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		log.Printf("Client disconnected: %s", c.conn.RemoteAddr().String())
		clientsMutex.Lock()
		delete(clients, c)
		clientsMutex.Unlock()
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func gracefulShutdown() {
	log.Println("Received shutdown signal, starting graceful shutdown...")

	clientsMutex.Lock()
	for client := range clients {
		// Notify client of shutdown
		select {
		case client.send <- []byte("SHUTDOWN"):
		default:
			close(client.send)
		}
	}
	clientsMutex.Unlock()

	// Send random number of business data (1-10)
	rand.Seed(time.Now().UnixNano())
	numData := rand.Intn(10) + 1
	log.Printf("Sending %d additional business data items", numData)

	clientsMutex.Lock()
	for client := range clients {
		for i := 0; i < numData; i++ {
			message := []byte("Final data: " + time.Now().Format("15:04:05") + " - " + string(rune(i)))
			select {
			case client.send <- message:
				time.Sleep(100 * time.Millisecond) // Simulate processing time
			default:
				close(client.send)
			}
		}
		// Send close message
		select {
		case client.send <- []byte("CLOSE"):
		default:
			close(client.send)
		}
	}
	clientsMutex.Unlock()

	time.Sleep(2 * time.Second) // Wait for clients to process
	log.Println("Graceful shutdown complete")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ws", handleConnections)

	go func() {
		log.Printf("Server starting on :%s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	gracefulShutdown()
	os.Exit(0)
}
