package web

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/aquamarinepk/aqm/log"
)

// TemplateManager handles template loading and rendering with embedded FS support.
type TemplateManager struct {
	fs        embed.FS
	templates *template.Template
	log       log.Logger
}

// NewTemplateManager creates a new template manager.
// Parsing is deferred to Start() to support async initialization.
func NewTemplateManager(fs embed.FS, log log.Logger) *TemplateManager {
	return &TemplateManager{
		fs:  fs,
		log: log,
	}
}

// Start parses templates from embedded FS.
// This can be called asynchronously by the aqm framework.
func (m *TemplateManager) Start(ctx context.Context) error {
	tmpl := template.New("")

	// Walk the embedded FS and parse all .html files
	err := fs.WalkDir(m.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		content, err := m.fs.ReadFile(path)
		if err != nil {
			return err
		}

		// Use base name as template name (e.g., "list.html")
		name := filepath.Base(path)
		_, err = tmpl.New(name).Parse(string(content))
		return err
	})

	if err != nil {
		m.log.Errorf("error parsing templates: %v", err)
		return err
	}

	m.templates = tmpl
	m.log.Info("templates loaded successfully")
	return nil
}

// Stop implements the lifecycle interface.
func (m *TemplateManager) Stop(ctx context.Context) error {
	return nil
}

// Render renders a full template with error handling.
func (m *TemplateManager) Render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := m.templates.ExecuteTemplate(w, name, data); err != nil {
		m.log.Errorf("error rendering template %s: %v", name, err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

// RenderPartial renders a template fragment without setting Content-Type.
// Used for htmx partial updates.
func (m *TemplateManager) RenderPartial(w http.ResponseWriter, name string, data interface{}) {
	if err := m.templates.ExecuteTemplate(w, name, data); err != nil {
		m.log.Errorf("error rendering partial %s: %v", name, err)
		http.Error(w, "Partial rendering error", http.StatusInternalServerError)
	}
}
