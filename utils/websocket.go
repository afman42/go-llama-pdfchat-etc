package utils

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nxadm/tail"
)

// https://medium.com/better-programming/streaming-log-files-in-real-time-with-golang-and-websockets-a-tail-f-simulation-89e080bebfe
type Client struct {
	socket *websocket.Conn
	send   chan []byte
}

type Broadcaster struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (b *Broadcaster) Run() {
	for {
		select {
		case client := <-b.register:
			b.clients[client] = true
		case client := <-b.unregister:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
		case message := <-b.broadcast:
			for client := range b.clients {
				client.send <- []byte(message)
			}
		}
	}
}

func readLastNLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 20 {
			lines = lines[1:]
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return lines, nil
}

func (b *Broadcaster) initialRead(client *Client, filePath string, logger *log.Logger) {
	// Send last n lines from file to the client
	lines, err := readLastNLines(filePath)
	if err != nil {
		logger.Println(err)
		return
	}
	client.send <- []byte(strings.Join(lines, "\n"))
}

func HandleWebSocketConnection(b *Broadcaster, filePath string, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Println(err)
			return
		}
		client := &Client{socket: ws, send: make(chan []byte)}
		b.register <- client

		go b.initialRead(client, filePath, logger)

		go func() {
			defer func() {
				b.unregister <- client
				ws.Close()
			}()

			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					b.unregister <- client
					ws.Close()
					break
				}
			}
		}()

		go func() {
			defer ws.Close()
			for {
				message, ok := <-client.send
				if !ok {
					ws.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				ws.WriteMessage(websocket.TextMessage, message)
			}
		}()
	}
}

func (b *Broadcaster) TailFile(filepath string, logger *log.Logger) {
	//Need Refactor
	t, err := tail.TailFile(
		filepath,
		tail.Config{Location: &tail.SeekInfo{Offset: 0, Whence: 2}, Follow: true},
	)
	if err != nil {
		logger.Fatalf("tail file err: %v", err)
	}

	for line := range t.Lines {
		line.Text = line.Text + "\n"
		if line.Text != "" {
			b.broadcast <- line.Text
		}
	}
}
