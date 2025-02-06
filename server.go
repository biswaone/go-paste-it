package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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
	s.router.HandleFunc("/view/{id}", s.HandleView)
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
	id := r.URL.Path[len("/view")+1:]
	log.Println(id)
	snippet, err := s.store.GetSnippet(r.Context(), id)
	log.Println(snippet)
	if err != nil {
		log.Println(err)
		http.Error(w, "Paste not found or expired", http.StatusNotFound)
		return
	}
	if snippet.Expiration.Before(time.Now()) {
		// Delete expired snippet from DB
		s.store.DeleteSnippet(r.Context(), id)
		http.Error(w, "Paste has expired", http.StatusGone)
		return
	}
	// If password protection is enabled
	if snippet.EnablePassword {
		password := r.FormValue("password")
		if password == "" {
			templates["password.html"].Execute(w, map[string]string{"ID": id})
			return
		}
		if !checkPasswordHash(password, *snippet.Password) {
			templates["password.html"].Execute(w, map[string]string{
				"ID":           id,
				"ErrorMessage": "Invalid password. Please try again.",
			})
			return
		}
	}
	templates["view.html"].Execute(w, map[string]string{
		"Created": snippet.CreatedAt.Local().String(),
		"Content": snippet.Content,
	})

	if snippet.ViewCount+1 > 1 && snippet.BurnAfterRead {
		s.store.DeleteSnippet(r.Context(), id)
	} else {
		snippet.ViewCount += 1
		s.store.UpdateSnippet(r.Context(), id, snippet)
	}
}
