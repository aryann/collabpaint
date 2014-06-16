package main

import (
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var templates = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeWebSocketUpgrader(session *Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("HTTP connection upgraded to WebSocket connection: %v\n", r)

		client := NewClient(conn)
		session.Join(client)
		go client.ReadLoop(session)
		go client.WriteLoop()
	}
}

type Client struct {
	conn          *websocket.Conn
	OutgoingQueue chan string
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:          conn,
		OutgoingQueue: make(chan string),
	}
}

func (c *Client) ReadLoop(session *Session) {
	for {
		_, message, err := c.conn.ReadMessage()
		if err == io.EOF {
			session.Leave(c)
		}
		session.Share(string(message))
	}
}

func (c *Client) WriteLoop() {
	for message := range c.OutgoingQueue {
		c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}

type Session struct {
	members           map[*Client]bool
	pendingLeaves     chan *Client
	pendingJoins      chan *Client
	pendingBroadcasts chan string
}

func NewSession() *Session {
	return &Session{
		members:           make(map[*Client]bool),
		pendingLeaves:     make(chan *Client),
		pendingJoins:      make(chan *Client),
		pendingBroadcasts: make(chan string),
	}
}

func (h *Session) Start() {
	for {
		select {
		case client := <-h.pendingLeaves:
			delete(h.members, client)
			close(client.OutgoingQueue)
			log.Printf("Client left: %v\n", client)

		case client := <-h.pendingJoins:
			h.members[client] = true
			log.Printf("Client joined: %v\n", client)

		case message := <-h.pendingBroadcasts:
			log.Printf("Broadcasting message: %v\n", message)
			for client := range h.members {
				select {
				case client.OutgoingQueue <- message:
				default:
					delete(h.members, client)
					close(client.OutgoingQueue)
				}
			}
		}
	}
}

func (h *Session) Join(client *Client) {
	h.pendingJoins <- client
}

func (h *Session) Leave(client *Client) {
	h.pendingLeaves <- client
}

func (h *Session) Share(message string) {
	h.pendingBroadcasts <- message
}

func main() {
	// TODO(aryann): Add support for multiple sessions.
	session := NewSession()
	go session.Start()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/echo/", makeWebSocketUpgrader(session))

	http.ListenAndServe(":8080", nil)
}
