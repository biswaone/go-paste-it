package main

import (
	"fmt"
	"net/http"
)

type server struct {
	router *http.ServeMux
	store  store
}

func newServer(store store) (*server, error) {
	s := &server{store: store}
	s.init()
	return s, nil
}

func (s *server) init() {
	s.router = http.NewServeMux()
	s.router.HandleFunc("/_health", s.handleHealthCheck)
	s.router.HandleFunc("/", s.handleHomePage)
	s.router.HandleFunc("/paste", s.HandlePaste)
	s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func (s *server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Application is healthy")))
}

func (s *server) handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templates["index.html"].Execute(w, nil)
}

func (s *server) HandleView(w http.ResponseWriter, r *http.Request) {
	// id := r.URL.Path[len("/view/"):]
	snippet := "Hello World"
	templates["view.html"].Execute(w, snippet)
}
