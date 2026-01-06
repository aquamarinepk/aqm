package list

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// testService is a simple fake service for testing handlers.
type testService struct {
	getOrCreateListFunc func(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	getListFunc         func(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	addItemFunc         func(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error)
	updateItemFunc      func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error)
	removeItemFunc      func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*TodoList, error)
}

func (s *testService) GetOrCreateList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	if s.getOrCreateListFunc != nil {
		return s.getOrCreateListFunc(ctx, userID)
	}
	return NewTodoList(userID), nil
}

func (s *testService) GetList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	if s.getListFunc != nil {
		return s.getListFunc(ctx, userID)
	}
	return nil, ErrNotFound
}

func (s *testService) AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error) {
	if s.addItemFunc != nil {
		return s.addItemFunc(ctx, userID, text)
	}
	return nil, nil
}

func (s *testService) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error) {
	if s.updateItemFunc != nil {
		return s.updateItemFunc(ctx, userID, itemID, text, completed)
	}
	return nil, nil
}

func (s *testService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*TodoList, error) {
	if s.removeItemFunc != nil {
		return s.removeItemFunc(ctx, userID, itemID)
	}
	return nil, nil
}

func TestNewHandler(t *testing.T) {
	svc := &testService{}

	h := NewHandler(svc, nil, nil)

	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if h.service == nil {
		t.Error("service should not be nil")
	}

	if h.log == nil {
		t.Error("logger should not be nil")
	}
}

func TestHandlerGetList(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		service    *testService
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success",
			userID: userID.String(),
			service: &testService{
				getListFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					list.AddItem("Test item")
					return list, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "list not found",
			userID: userID.String(),
			service: &testService{
				getListFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "LIST_NOT_FOUND",
		},
		{
			name:       "invalid user ID",
			userID:     "invalid-uuid",
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:   "service error",
			userID: userID.String(),
			service: &testService{
				getListFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, errors.New("database error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service, nil, nil)
			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/list", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handleGetList() status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp errorResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Code != tt.wantCode {
					t.Errorf("handleGetList() code = %s, want %s", resp.Code, tt.wantCode)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var list TodoList
				if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestHandlerAddItem(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		payload    interface{}
		service    *testService
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success",
			userID: userID.String(),
			payload: map[string]string{
				"text": "New item",
			},
			service: &testService{
				addItemFunc: func(ctx context.Context, uid uuid.UUID, text string) (*TodoList, error) {
					list := NewTodoList(userID)
					list.AddItem(text)
					return list, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "empty text",
			userID: userID.String(),
			payload: map[string]string{
				"text": "",
			},
			service: &testService{
				addItemFunc: func(ctx context.Context, uid uuid.UUID, text string) (*TodoList, error) {
					return nil, ErrItemTextEmpty
				},
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "ITEM_TEXT_EMPTY",
		},
		{
			name:   "text too long",
			userID: userID.String(),
			payload: map[string]string{
				"text": "x",
			},
			service: &testService{
				addItemFunc: func(ctx context.Context, uid uuid.UUID, text string) (*TodoList, error) {
					return nil, ErrItemTextTooLong
				},
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "ITEM_TEXT_TOO_LONG",
		},
		{
			name:       "invalid user ID",
			userID:     "invalid-uuid",
			payload:    map[string]string{"text": "Item"},
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "invalid payload",
			userID:     userID.String(),
			payload:    "invalid json",
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PAYLOAD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service, nil, nil)
			r := chi.NewRouter()
			h.RegisterRoutes(r)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.payload)

			req := httptest.NewRequest(http.MethodPost, "/users/"+tt.userID+"/list/items", &body)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handleAddItem() status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp errorResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Code != tt.wantCode {
					t.Errorf("handleAddItem() code = %s, want %s", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandlerUpdateItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		itemID     string
		payload    interface{}
		service    *testService
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success update text",
			userID: userID.String(),
			itemID: itemID.String(),
			payload: map[string]interface{}{
				"text": "Updated text",
			},
			service: &testService{
				updateItemFunc: func(ctx context.Context, uid, id uuid.UUID, text *string, completed *bool) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: *text}
					list.Items = []TodoItem{item}
					return list, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "success update completed",
			userID: userID.String(),
			itemID: itemID.String(),
			payload: map[string]interface{}{
				"completed": true,
			},
			service: &testService{
				updateItemFunc: func(ctx context.Context, uid, id uuid.UUID, text *string, completed *bool) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Completed: *completed}
					list.Items = []TodoItem{item}
					return list, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "item not found",
			userID: userID.String(),
			itemID: itemID.String(),
			payload: map[string]interface{}{
				"text": "Text",
			},
			service: &testService{
				updateItemFunc: func(ctx context.Context, uid, id uuid.UUID, text *string, completed *bool) (*TodoList, error) {
					return nil, ErrItemNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "ITEM_NOT_FOUND",
		},
		{
			name:       "invalid user ID",
			userID:     "invalid-uuid",
			itemID:     itemID.String(),
			payload:    map[string]string{"text": "Text"},
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "invalid item ID",
			userID:     userID.String(),
			itemID:     "invalid-uuid",
			payload:    map[string]string{"text": "Text"},
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ITEM_ID",
		},
		{
			name:       "invalid payload",
			userID:     userID.String(),
			itemID:     itemID.String(),
			payload:    "invalid json",
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PAYLOAD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service, nil, nil)
			r := chi.NewRouter()
			h.RegisterRoutes(r)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.payload)

			req := httptest.NewRequest(http.MethodPatch, "/users/"+tt.userID+"/list/items/"+tt.itemID, &body)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handleUpdateItem() status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp errorResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Code != tt.wantCode {
					t.Errorf("handleUpdateItem() code = %s, want %s", resp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandlerRemoveItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		itemID     string
		service    *testService
		wantStatus int
		wantCode   string
	}{
		{
			name:   "success",
			userID: userID.String(),
			itemID: itemID.String(),
			service: &testService{
				removeItemFunc: func(ctx context.Context, uid, id uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "item not found",
			userID: userID.String(),
			itemID: itemID.String(),
			service: &testService{
				removeItemFunc: func(ctx context.Context, uid, id uuid.UUID) (*TodoList, error) {
					return nil, ErrItemNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "ITEM_NOT_FOUND",
		},
		{
			name:   "list not found",
			userID: userID.String(),
			itemID: itemID.String(),
			service: &testService{
				removeItemFunc: func(ctx context.Context, uid, id uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "LIST_NOT_FOUND",
		},
		{
			name:       "invalid user ID",
			userID:     "invalid-uuid",
			itemID:     itemID.String(),
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "invalid item ID",
			userID:     userID.String(),
			itemID:     "invalid-uuid",
			service:    &testService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ITEM_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service, nil, nil)
			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID+"/list/items/"+tt.itemID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handleRemoveItem() status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp errorResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Code != tt.wantCode {
					t.Errorf("handleRemoveItem() code = %s, want %s", resp.Code, tt.wantCode)
				}
			}
		})
	}
}
