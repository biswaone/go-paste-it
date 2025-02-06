package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

type Snippet struct {
	ID             string    `bson:"id"`
	Title          string    `bson:"title"`
	Expiration     time.Time `bson:"expiration"`
	BurnAfterRead  bool      `bson:"burn_after_read"`
	EnablePassword bool      `bson:"enable_password"`
	Password       *string   `bson:"password,omitempty"`
	Content        string    `bson:"content"`
	ViewCount      uint64    `bson:"view_count"`
	CreatedAt      time.Time `bson:"created_at"`
}

// NewSnippet creates and initializes a new Snippet instance.
func NewSnippet(title, content, expiration string, burnAfterRead, enablePassword bool, password *string) *Snippet {
	return &Snippet{
		ID:             generateID([]byte(content)),
		Title:          title,
		Expiration:     getExpirationTime(expiration),
		BurnAfterRead:  burnAfterRead,
		EnablePassword: enablePassword,
		Password:       password,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}

// generateID generates a unique ID for the snippet based on its content.
func generateID(content []byte) string {
	salt := time.Now().String()
	h := sha256.New()
	io.WriteString(h, salt)
	h.Write(content)
	sum := h.Sum(nil)
	encoded := base64.URLEncoding.EncodeToString(sum)

	hashLen := 11
	for hashLen <= len(encoded) && encoded[hashLen-1] == '_' {
		hashLen++
	}
	return encoded[:hashLen]
}

// getExpirationTime calculates the expiration time based on the provided key.
func getExpirationTime(expiration string) time.Time {
	duration, exists := expirationDurations[expiration]
	if !exists {
		duration = expirationDurations["never"]
	}
	return time.Now().Add(duration)
}

// expirationDurations maps expiration keys to their respective durations.
var expirationDurations = map[string]time.Duration{
	"never": 100 * 365 * 24 * time.Hour, // 100 years
	"10m":   10 * time.Minute,
	"1h":    time.Hour,
	"1d":    24 * time.Hour,
	"1w":    7 * 24 * time.Hour,
}

// Validate checks if the snippet content is valid.
func (s *Snippet) Validate() error {
	if s.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if len(s.Content) > maxSnippetSize {
		return fmt.Errorf("content is too large")
	}
	return nil
}
