package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var templates = make(map[string]*template.Template)

func loadTemplates() error {
	templateDir := "templates"
	pattern := filepath.Join(templateDir, "*.html")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error finding templates: %v", err)
	}
	for _, file := range files {
		name := filepath.Base(file)
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %v", name, err)
		}
		templates[name] = tmpl
	}
	return nil
}

func main() {
	if err := loadTemplates(); err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}

	// Initialize MongoDB
	client, err := initMongoDB("mongodb://localhost:27017", context.Background())
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	mongoStore := &mongoStore{client: client}

	s, err := newServer(mongoStore)
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: s.router,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("Server started at :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("Error closing MongoDB connection: %v", err)
	}

	log.Println("Server stopped gracefully")
}
