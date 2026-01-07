package htmx

import (
	"net/http"

	"github.com/aquamarinepk/aqm/log"
)

// RespondDelete responds correctly to an htmx delete request.
// Returns 200 OK with empty body on success, or 500 on error.
func RespondDelete(w http.ResponseWriter, err error, log log.Logger) {
	if err != nil {
		log.Errorf("error deleting: %v", err)
		http.Error(w, "Cannot delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}
