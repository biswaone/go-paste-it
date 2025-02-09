package main

import (
	"log"
	"net/http"
	"time"
)

func (s *server) HandleView(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/view")+1:]
	snippet, err := s.store.GetSnippet(r.Context(), id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		templates["404.html"].Execute(w, nil)
		return
	}

	if snippet.Expiration.Before(time.Now()) {
		// Delete expired snippet from DB
		_ = s.store.DeleteSnippet(r.Context(), id)
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
		if !checkPasswordHash(password, snippet.Password) {
			templates["password.html"].Execute(w, map[string]string{
				"ID":           id,
				"ErrorMessage": "Invalid password. Please try again.",
			})
			return
		}
	}

	if snippet.ViewCount+1 > 1 && snippet.BurnAfterRead {
		err := s.store.DeleteSnippet(r.Context(), id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		templates["view.html"].Execute(w, map[string]string{
			"Created": snippet.CreatedAt.Local().String(),
			"Content": snippet.Content,
		})
		return
	} else {
		snippet.ViewCount += 1
		err := s.store.UpdateSnippet(r.Context(), id, snippet)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
	templates["view.html"].Execute(w, map[string]string{
		"Created": snippet.CreatedAt.Local().String(),
		"Content": snippet.Content,
	})

}
