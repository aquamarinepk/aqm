package web

import (
	"html/template"
	"net/http"

	"github.com/aquamarinepk/aqm/log"
)

// RenderTemplate executes a template with standard error handling.
func RenderTemplate(w http.ResponseWriter, tmpl *template.Template, name string, data interface{}, log log.Logger) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Errorf("Template render error (%s): %v", name, err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

// RenderPartial executes a partial template (typically for htmx responses).
func RenderPartial(w http.ResponseWriter, tmpl *template.Template, name string, data interface{}, log log.Logger) {
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Errorf("Partial render error (%s): %v", name, err)
		http.Error(w, "Partial rendering error", http.StatusInternalServerError)
	}
}

// RedirectOrHXRedirect redirects for normal requests or sets HX-Redirect header for htmx requests.
func RedirectOrHXRedirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}
