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
			break
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
	allUpdates        []string
}

func NewSession() *Session {
	return &Session{
		members:           make(map[*Client]bool),
		pendingLeaves:     make(chan *Client),
		pendingJoins:      make(chan *Client),
		pendingBroadcasts: make(chan string),
		allUpdates:        make([]string, 0),
	}
}

func (s *Session) Start() {
	for {
    log.Println(s)
		select {
		case client := <-s.pendingLeaves:
			delete(s.members, client)
			close(client.OutgoingQueue)
			log.Printf("Client left: %v\n", client)

		case client := <-s.pendingJoins:
			s.members[client] = true
			for _, message := range s.allUpdates {
				client.OutgoingQueue <- message
			}
			log.Printf("Client joined: %v\n", client)

		case message := <-s.pendingBroadcasts:
			log.Printf("Broadcasting message: %v\n", message)
			s.allUpdates = append(s.allUpdates, message)
			for client := range s.members {
				client.OutgoingQueue <- message
			}
		}
	}
}

func (s *Session) Join(client *Client) {
	s.pendingJoins <- client
}

func (s *Session) Leave(client *Client) {
	s.pendingLeaves <- client
}

func (s *Session) Share(message string) {
	s.pendingBroadcasts <- message
}

func main() {
	// TODO(aryann): Add support for multiple sessions.
	session := NewSession()
	go session.Start()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/websocket/", makeWebSocketUpgrader(session))

	http.ListenAndServe(":8080", nil)
}
