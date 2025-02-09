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

func run() error {
	if err := loadTemplates(); err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}

	// Get MongoDB URI from environment variable with fallback
	// set MONGODB_URI = mongodb://localhost:27017 for local running
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongo:27017" // default for Docker environment
	}

	// Initialize MongoDB
	client, err := initMongoDB(context.Background(), mongoURI)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	mongoStore := newMongoStore(client)
	err = mongoStore.createIndexes(context.Background())
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

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
	return nil

}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

}
