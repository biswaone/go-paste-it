package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxSnippetSize = 64 * 1024

type Snippet struct {
	ID             string    `bson:"id"`
	Title          string    `bson:"title"`
	Expiration     time.Time `bson:"expiration"`
	BurnAfterRead  bool      `bson:"burn_after_read"`
	EnablePassword bool      `bson:"enable_password"`
	Password       *string   `bson:"password,omitempty"`
	Content        string    `bson:"content"`
}

// ExpirationDurations maps expiration keys to their respective durations
var ExpirationDurations = map[string]time.Duration{
	"never": 100 * 365 * 24 * time.Hour, // 100 years
	"10m":   10 * time.Minute,
	"1h":    time.Hour,
	"1d":    24 * time.Hour,
	"1w":    7 * 24 * time.Hour,
}

func getExpirationTime(expiration string) time.Time {
	if duration, exists := ExpirationDurations[expiration]; exists {
		return time.Now().Add(duration)
	}
	return time.Now().Add(ExpirationDurations["never"])
}

func generateID(body []byte) string {
	salt := time.Now().String()
	h := sha256.New()
	_, _ = io.WriteString(h, salt)
	h.Write(body)
	sum := h.Sum(nil)
	encoded := base64.URLEncoding.EncodeToString(sum)

	hashLen := 11
	for hashLen <= len(encoded) && encoded[hashLen-1] == '_' {
		hashLen++
	}
	return encoded[:hashLen]
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
	if len(content) > maxSnippetSize {
		http.Error(w, "Content is too large", http.StatusRequestEntityTooLarge)
		return
	}

	title := r.FormValue("title")
	expiration := r.FormValue("expiration")
	burnAfterRead := r.FormValue("burn_after_read") == "on"
	enablePassword := r.FormValue("enable_password") == "on"
	password := r.FormValue("password")

	var hashedPassword *string
	if enablePassword {
		if password == "" {
			http.Error(w, "Password cannot be empty", http.StatusBadRequest)
			return
		}
		hashed, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		hashedPassword = &hashed
	}

	var body bytes.Buffer
	_, err := io.Copy(&body, io.LimitReader(r.Body, maxSnippetSize+1))
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if body.Len() > maxSnippetSize {
		http.Error(w, "Snippet is too large", http.StatusRequestEntityTooLarge)
		return
	}

	id := generateID(body.Bytes())

	snippet := &Snippet{
		ID:             id,
		Title:          title,
		Expiration:     getExpirationTime(expiration),
		BurnAfterRead:  burnAfterRead,
		EnablePassword: enablePassword,
		Content:        content,
		Password:       hashedPassword,
	}
	fmt.Println(snippet)

	http.Redirect(w, r, "/view/"+id, http.StatusSeeOther)
}
