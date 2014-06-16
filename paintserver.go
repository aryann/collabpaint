package main

import (
  "net/http"
  "html/template"
)

var templates = template.Must(template.ParseFiles("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
  err := templates.ExecuteTemplate(w, "index.html", nil)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func main() {
  http.HandleFunc("/", indexHandler)
  http.ListenAndServe(":8080", nil)
}
