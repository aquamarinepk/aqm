package list

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler wires HTTP routes for todo lists.
type Handler struct {
	service ServiceInterface
	log     log.Logger
	cfg     *config.Config
}

// NewHandler creates a new handler instance.
func NewHandler(service ServiceInterface, cfg *config.Config, log log.Logger) *Handler {
	if log == nil {
		log = &noopLogger{}
	}
	return &Handler{
		service: service,
		log:     log,
		cfg:     cfg,
	}
}

// RegisterRoutes registers all list routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/users/{userID}/list", func(r chi.Router) {
		r.Get("/", h.handleGetList)
		r.Post("/items", h.handleAddItem)
		r.Route("/items/{itemID}", func(r chi.Router) {
			r.Patch("/", h.handleUpdateItem)
			r.Delete("/", h.handleRemoveItem)
		})
	})
}

func (h *Handler) handleGetList(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	list, err := h.service.GetList(r.Context(), userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) handleAddItem(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	var payload struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Malformed JSON payload")
		return
	}

	list, err := h.service.AddItem(r.Context(), userID, payload.Text)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, list)
}

func (h *Handler) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", err.Error())
		return
	}

	var payload struct {
		Text      *string `json:"text"`
		Completed *bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Malformed JSON payload")
		return
	}

	list, err := h.service.UpdateItem(r.Context(), userID, itemID, payload.Text, payload.Completed)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) handleRemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ITEM_ID", err.Error())
		return
	}

	list, err := h.service.RemoveItem(r.Context(), userID, itemID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "LIST_NOT_FOUND", "List not found")
	case errors.Is(err, ErrItemNotFound):
		writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found")
	case errors.Is(err, ErrItemTextEmpty):
		writeError(w, http.StatusBadRequest, "ITEM_TEXT_EMPTY", err.Error())
	case errors.Is(err, ErrItemTextTooLong):
		writeError(w, http.StatusBadRequest, "ITEM_TEXT_TOO_LONG", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

// Helper functions

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func parseUserID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "userID"))
}

func parseItemID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "itemID"))
}
