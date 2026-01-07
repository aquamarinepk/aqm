package web

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

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

		// Normalize path to always start from "assets/"
		// e.g., "testdata/assets/templates/test/page.html" → "assets/templates/test/page.html"
		normalizedPath := m.normalizePath(path)

		_, err = tmpl.New(normalizedPath).Parse(string(content))
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
// Namespace and template are combined to form the path: "assets/templates/{namespace}/{template}.html"
func (m *TemplateManager) Render(w http.ResponseWriter, namespace, template string, data interface{}) {
	path := m.buildPath(namespace, template)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := m.templates.ExecuteTemplate(w, path, data); err != nil {
		m.log.Errorf("error rendering template %s/%s: %v", namespace, template, err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

// RenderPartial renders a template fragment without setting Content-Type.
// Used for htmx partial updates.
// Namespace and template are combined to form the path: "assets/templates/{namespace}/{template}.html"
func (m *TemplateManager) RenderPartial(w http.ResponseWriter, namespace, template string, data interface{}) {
	path := m.buildPath(namespace, template)
	if err := m.templates.ExecuteTemplate(w, path, data); err != nil {
		m.log.Errorf("error rendering partial %s/%s: %v", namespace, template, err)
		http.Error(w, "Partial rendering error", http.StatusInternalServerError)
	}
}

// buildPath constructs the full template path from namespace and template name.
func (m *TemplateManager) buildPath(namespace, template string) string {
	return filepath.Join("assets", "templates", namespace, template+".html")
}

// normalizePath extracts the path starting from "assets/" to ensure consistent template names.
func (m *TemplateManager) normalizePath(path string) string {
	// Find "assets/" in the path and return everything from there
	if idx := strings.Index(path, "assets"); idx >= 0 {
		return path[idx:]
	}
	return path
}
