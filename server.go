package main

import (
	"fmt"
	"net/http"
)

type server struct {
	router *http.ServeMux
}

func newServer() (*server, error) {
	s := &server{}
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

func (s *server) HandlePaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}
	id := "123"
	http.Redirect(w, r, "/view/"+id, http.StatusSeeOther)
}

func (s *server) HandleView(w http.ResponseWriter, r *http.Request) {
	// id := r.URL.Path[len("/view/"):]
	snippet := "Hello World"
	templates["view.html"].Execute(w, snippet)
}
