package fake

import (
	"context"
	"errors"
	"testing"

	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list"
	"github.com/google/uuid"
)

func TestRepoSave(t *testing.T) {
	tests := []struct {
		name     string
		saveFunc func(ctx context.Context, l *list.TodoList) error
		wantErr  bool
	}{
		{
			name:     "nil function returns nil",
			saveFunc: nil,
			wantErr:  false,
		},
		{
			name: "function called successfully",
			saveFunc: func(ctx context.Context, l *list.TodoList) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "function returns error",
			saveFunc: func(ctx context.Context, l *list.TodoList) error {
				return errors.New("save error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Store{SaveFunc: tt.saveFunc}
			todoList := list.NewTodoList(uuid.New())

			err := repo.Save(context.Background(), todoList)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepoFindByUserID(t *testing.T) {
	userID := uuid.New()
	todoList := list.NewTodoList(userID)

	tests := []struct {
		name             string
		findByUserIDFunc func(ctx context.Context, userID uuid.UUID) (*list.TodoList, error)
		wantList         *list.TodoList
		wantErr          error
	}{
		{
			name:             "nil function returns ErrNotFound",
			findByUserIDFunc: nil,
			wantList:         nil,
			wantErr:          list.ErrNotFound,
		},
		{
			name: "function returns list",
			findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*list.TodoList, error) {
				return todoList, nil
			},
			wantList: todoList,
			wantErr:  nil,
		},
		{
			name: "function returns error",
			findByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*list.TodoList, error) {
				return nil, list.ErrNotFound
			},
			wantList: nil,
			wantErr:  list.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Store{FindByUserIDFunc: tt.findByUserIDFunc}

			got, err := repo.FindByUserID(context.Background(), userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.wantList {
				t.Errorf("FindByUserID() got = %v, want %v", got, tt.wantList)
			}
		})
	}
}

func TestRepoDelete(t *testing.T) {
	listID := uuid.New()

	tests := []struct {
		name       string
		deleteFunc func(ctx context.Context, listID uuid.UUID) error
		wantErr    bool
	}{
		{
			name:       "nil function returns nil",
			deleteFunc: nil,
			wantErr:    false,
		},
		{
			name: "function called successfully",
			deleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "function returns error",
			deleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return list.ErrNotFound
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Store{DeleteFunc: tt.deleteFunc}

			err := repo.Delete(context.Background(), listID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
