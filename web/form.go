package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FormValues provides typed access to form values.
type FormValues struct {
	r *http.Request
}

// ParseForm parses the request form and returns FormValues helper.
func ParseForm(r *http.Request) (*FormValues, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return &FormValues{r: r}, nil
}

// String returns the value of a form field.
func (f *FormValues) String(name string) string {
	return f.r.FormValue(name)
}

// StringOr returns the value of a form field or the default if empty.
func (f *FormValues) StringOr(name, def string) string {
	v := f.r.FormValue(name)
	if v == "" {
		return def
	}
	return v
}

// Bool returns true if the form field value is "true" or "on".
func (f *FormValues) Bool(name string) bool {
	v := f.r.FormValue(name)
	return v == "true" || v == "on"
}

// UUID parses a form field value as UUID.
func (f *FormValues) UUID(name string) (uuid.UUID, error) {
	return uuid.Parse(f.r.FormValue(name))
}

// ParseIDParam extracts and parses a UUID from URL path parameters.
func ParseIDParam(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}
