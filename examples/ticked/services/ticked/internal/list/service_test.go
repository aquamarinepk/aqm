package list

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// testStore is a simple fake repository for testing.
type testStore struct {
	saveFunc         func(ctx context.Context, list *TodoList) error
	findByUserIDFunc func(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	deleteFunc       func(ctx context.Context, listID uuid.UUID) error
}

func (r *testStore) Save(ctx context.Context, list *TodoList) error {
	if r.saveFunc != nil {
		return r.saveFunc(ctx, list)
	}
	return nil
}

func (r *testStore) FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	if r.findByUserIDFunc != nil {
		return r.findByUserIDFunc(ctx, userID)
	}
	return nil, ErrNotFound
}

func (r *testStore) Delete(ctx context.Context, listID uuid.UUID) error {
	if r.deleteFunc != nil {
		return r.deleteFunc(ctx, listID)
	}
	return nil
}

func TestNewService(t *testing.T) {
	repo := &testStore{}

	svc := NewService(repo, nil, nil)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}

	if svc.store != repo {
		t.Error("repo not set correctly")
	}

	if svc.log == nil {
		t.Error("logger should not be nil")
	}

	// Verify noop logger implements all methods without panicking
	svc.log.Debug("test")
	svc.log.Debugf("test %s", "format")
	svc.log.Info("test")
	svc.log.Infof("test %s", "format")
	svc.log.Error("test")
	svc.log.Errorf("test %s", "format")
	newLogger := svc.log.With("key", "value")
	if newLogger == nil {
		t.Error("With should return non-nil logger")
	}
}

func TestServiceGetOrCreateList(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		repo     *testStore
		wantErr  bool
		validate func(*testing.T, *TodoList, error)
	}{
		{
			name: "list exists",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					list.AddItem("Existing item")
					return list, nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if list == nil {
					t.Fatal("list is nil")
				}
				if len(list.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(list.Items))
				}
			},
		},
		{
			name: "list not found creates new",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if list == nil {
					t.Fatal("list is nil")
				}
				if list.UserID != userID {
					t.Errorf("userID = %v, want %v", list.UserID, userID)
				}
				if len(list.Items) != 0 {
					t.Errorf("expected 0 items, got %d", len(list.Items))
				}
			},
		},
		{
			name: "save error on create",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return errors.New("save error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
				if list != nil {
					t.Error("list should be nil on error")
				}
			},
		},
		{
			name: "find error",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, errors.New("database error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
				if list != nil {
					t.Error("list should be nil on error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil, nil)

			list, err := svc.GetOrCreateList(context.Background(), userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrCreateList() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, list, err)
		})
	}
}

func TestServiceGetList(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *testStore
		wantErr error
	}{
		{
			name: "list found",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
			},
			wantErr: nil,
		},
		{
			name: "list not found",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil, nil)

			list, err := svc.GetList(context.Background(), userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetList() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && list == nil {
				t.Error("expected list, got nil")
			}

			if tt.wantErr != nil && list != nil {
				t.Error("expected nil list on error")
			}
		})
	}
}

func TestServiceAddItem(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		text     string
		repo     *testStore
		wantErr  bool
		validate func(*testing.T, *TodoList, error)
	}{
		{
			name: "add to existing list",
			text: "New task",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(list.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(list.Items))
				}
				if list.Items[0].Text != "New task" {
					t.Errorf("item text = %q, want %q", list.Items[0].Text, "New task")
				}
			},
		},
		{
			name: "add to new list",
			text: "First task",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(list.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(list.Items))
				}
			},
		},
		{
			name: "validation error empty text",
			text: "",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if !errors.Is(err, ErrItemTextEmpty) {
					t.Errorf("expected ErrItemTextEmpty, got %v", err)
				}
			},
		},
		{
			name: "save error",
			text: "Task",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return errors.New("save error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
			},
		},
		{
			name: "find error",
			text: "Task",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, errors.New("database error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil, nil)

			list, err := svc.AddItem(context.Background(), userID, tt.text)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, list, err)
		})
	}
}

func TestServiceUpdateItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name      string
		text      *string
		completed *bool
		repo      *testStore
		wantErr   bool
		validate  func(*testing.T, *TodoList, error)
	}{
		{
			name: "update text",
			text: stringPtr("Updated text"),
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: "Original"}
					list.Items = []TodoItem{item}
					return list, nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if list.Items[0].Text != "Updated text" {
					t.Errorf("text = %q, want %q", list.Items[0].Text, "Updated text")
				}
			},
		},
		{
			name:      "update completed",
			completed: boolPtr(true),
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: "Task", Completed: false}
					list.Items = []TodoItem{item}
					return list, nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !list.Items[0].Completed {
					t.Error("item should be completed")
				}
			},
		},
		{
			name: "item not found",
			text: stringPtr("Text"),
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if !errors.Is(err, ErrItemNotFound) {
					t.Errorf("expected ErrItemNotFound, got %v", err)
				}
			},
		},
		{
			name: "list not found",
			text: stringPtr("Text"),
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "save error",
			text: stringPtr("Text"),
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: "Task"}
					list.Items = []TodoItem{item}
					return list, nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return errors.New("save error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil, nil)

			list, err := svc.UpdateItem(context.Background(), userID, itemID, tt.text, tt.completed)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, list, err)
		})
	}
}

func TestServiceRemoveItem(t *testing.T) {
	userID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name     string
		repo     *testStore
		wantErr  bool
		validate func(*testing.T, *TodoList, error)
	}{
		{
			name: "remove existing item",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: "Task"}
					list.Items = []TodoItem{item}
					return list, nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(list.Items) != 0 {
					t.Errorf("expected 0 items, got %d", len(list.Items))
				}
			},
		},
		{
			name: "item not found",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return NewTodoList(userID), nil
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if !errors.Is(err, ErrItemNotFound) {
					t.Errorf("expected ErrItemNotFound, got %v", err)
				}
			},
		},
		{
			name: "list not found",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					return nil, ErrNotFound
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "save error",
			repo: &testStore{
				findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*TodoList, error) {
					list := NewTodoList(userID)
					item := TodoItem{ItemID: itemID, Text: "Task"}
					list.Items = []TodoItem{item}
					return list, nil
				},
				saveFunc: func(ctx context.Context, l *TodoList) error {
					return errors.New("save error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, list *TodoList, err error) {
				if err == nil {
					t.Error("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, nil, nil)

			list, err := svc.RemoveItem(context.Background(), userID, itemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, list, err)
		})
	}
}


func TestNoopLogger(t *testing.T) {
	l := &noopLogger{}
	l.Debug("test")
	l.Debugf("test %s", "arg")
	l.Info("test")
	l.Infof("test %s", "arg")
	l.Error("test")
	l.Errorf("test %s", "arg")

	withLogger := l.With("key", "value")
	if withLogger == nil {
		t.Error("With() should return logger")
	}
}
