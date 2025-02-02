package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	templates = make(map[string]*template.Template)
)

func loadTemplates() error {
	// Load all templates from the templates folder
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
		log.Fatalf("error loading templates: %v", err)
	}
	s, err := newServer()
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", s.router)
}
