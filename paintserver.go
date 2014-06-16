package main

import (
  "net/http"
  "html/template"
  "log"

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

func echoHandler(w http.ResponseWriter, r *http.Request) {
  _, err := websocketUpgrader.Upgrade(w, r, nil)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  log.Printf("HTTP connection upgraded to WebSocket connection: %v\n", r)
}

func main() {
  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/echo/", echoHandler)
  http.ListenAndServe(":8080", nil)
}
