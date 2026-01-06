package list

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTodoList(t *testing.T) {
	userID := uuid.New()

	list := NewTodoList(userID)

	if list == nil {
		t.Fatal("NewTodoList() returned nil")
	}

	if list.ListID == uuid.Nil {
		t.Error("ListID should not be nil")
	}

	if list.UserID != userID {
		t.Errorf("UserID = %v, want %v", list.UserID, userID)
	}

	if list.Items == nil {
		t.Error("Items should be initialized, not nil")
	}

	if len(list.Items) != 0 {
		t.Errorf("Items length = %d, want 0", len(list.Items))
	}

	if list.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if list.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if !list.CreatedAt.Equal(list.UpdatedAt) {
		t.Error("CreatedAt and UpdatedAt should be equal for new list")
	}
}

func TestTodoListID(t *testing.T) {
	list := NewTodoList(uuid.New())

	if list.ID() != list.ListID {
		t.Errorf("ID() = %v, want %v", list.ID(), list.ListID)
	}
}

func TestTodoListTouch(t *testing.T) {
	list := NewTodoList(uuid.New())
	originalUpdatedAt := list.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	list.Touch()

	if !list.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Touch() should update UpdatedAt timestamp")
	}

	if list.CreatedAt != originalUpdatedAt {
		t.Error("Touch() should not modify CreatedAt")
	}
}

func TestTodoListAddItem(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{
			name:    "valid item",
			text:    "Buy groceries",
			wantErr: nil,
		},
		{
			name:    "item with whitespace trimmed",
			text:    "  Clean house  ",
			wantErr: nil,
		},
		{
			name:    "empty text",
			text:    "",
			wantErr: ErrItemTextEmpty,
		},
		{
			name:    "only whitespace",
			text:    "   ",
			wantErr: ErrItemTextEmpty,
		},
		{
			name:    "text too long",
			text:    strings.Repeat("a", maxItemTextLength+1),
			wantErr: ErrItemTextTooLong,
		},
		{
			name:    "text at max length",
			text:    strings.Repeat("a", maxItemTextLength),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewTodoList(uuid.New())
			originalUpdatedAt := list.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			item, err := list.AddItem(tt.text)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("AddItem() error = %v, want %v", err, tt.wantErr)
				}
				if item != nil {
					t.Error("AddItem() should return nil item on error")
				}
				if list.UpdatedAt != originalUpdatedAt {
					t.Error("AddItem() should not update timestamp on error")
				}
				return
			}

			if err != nil {
				t.Errorf("AddItem() unexpected error: %v", err)
				return
			}

			if item == nil {
				t.Fatal("AddItem() returned nil item")
			}

			if item.ItemID == uuid.Nil {
				t.Error("ItemID should not be nil")
			}

			expectedText := strings.TrimSpace(tt.text)
			if item.Text != expectedText {
				t.Errorf("Item text = %q, want %q", item.Text, expectedText)
			}

			if item.Completed {
				t.Error("New item should not be completed")
			}

			if item.CompletedAt != nil {
				t.Error("New item should have nil CompletedAt")
			}

			if item.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if len(list.Items) != 1 {
				t.Errorf("Items length = %d, want 1", len(list.Items))
			}

			if list.Items[0].ItemID != item.ItemID {
				t.Error("Item should be added to list")
			}

			if !list.UpdatedAt.After(originalUpdatedAt) {
				t.Error("AddItem() should update list timestamp")
			}
		})
	}
}

func TestTodoListUpdateItem(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TodoList) uuid.UUID
		itemID    func(*TodoList) uuid.UUID
		text      *string
		completed *bool
		wantErr   error
	}{
		{
			name: "update text only",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Original text")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      stringPtr("Updated text"),
			completed: nil,
			wantErr:   nil,
		},
		{
			name: "update completed only",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Test item")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      nil,
			completed: boolPtr(true),
			wantErr:   nil,
		},
		{
			name: "update both text and completed",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Test item")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      stringPtr("New text"),
			completed: boolPtr(true),
			wantErr:   nil,
		},
		{
			name: "item not found",
			setup: func(l *TodoList) uuid.UUID {
				return uuid.Nil
			},
			itemID: func(l *TodoList) uuid.UUID {
				return uuid.New()
			},
			text:      stringPtr("Text"),
			completed: nil,
			wantErr:   ErrItemNotFound,
		},
		{
			name: "empty text",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Original")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      stringPtr(""),
			completed: nil,
			wantErr:   ErrItemTextEmpty,
		},
		{
			name: "whitespace only text",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Original")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      stringPtr("   "),
			completed: nil,
			wantErr:   ErrItemTextEmpty,
		},
		{
			name: "text too long",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Original")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      stringPtr(strings.Repeat("a", maxItemTextLength+1)),
			completed: nil,
			wantErr:   ErrItemTextTooLong,
		},
		{
			name: "mark as completed sets timestamp",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Test")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      nil,
			completed: boolPtr(true),
			wantErr:   nil,
		},
		{
			name: "mark as uncompleted clears timestamp",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Test")
				completed := true
				l.UpdateItem(item.ItemID, nil, &completed)
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			text:      nil,
			completed: boolPtr(false),
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewTodoList(uuid.New())
			tt.setup(list)
			itemID := tt.itemID(list)

			originalUpdatedAt := list.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			err := list.UpdateItem(itemID, tt.text, tt.completed)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("UpdateItem() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateItem() unexpected error: %v", err)
				return
			}

			idx := list.findItem(itemID)
			if idx == -1 {
				t.Fatal("Item not found in list after update")
			}

			item := list.Items[idx]

			if tt.text != nil {
				expectedText := strings.TrimSpace(*tt.text)
				if item.Text != expectedText {
					t.Errorf("Item text = %q, want %q", item.Text, expectedText)
				}
			}

			if tt.completed != nil {
				if item.Completed != *tt.completed {
					t.Errorf("Item completed = %v, want %v", item.Completed, *tt.completed)
				}

				if *tt.completed && item.CompletedAt == nil {
					t.Error("CompletedAt should be set when marking as completed")
				}

				if !*tt.completed && item.CompletedAt != nil {
					t.Error("CompletedAt should be nil when marking as uncompleted")
				}
			}

			if !list.UpdatedAt.After(originalUpdatedAt) {
				t.Error("UpdateItem() should update list timestamp")
			}
		})
	}
}

func TestTodoListRemoveItem(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TodoList) uuid.UUID
		itemID  func(*TodoList) uuid.UUID
		wantErr error
	}{
		{
			name: "remove existing item",
			setup: func(l *TodoList) uuid.UUID {
				item, _ := l.AddItem("Test item")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[0].ItemID
			},
			wantErr: nil,
		},
		{
			name: "remove non-existent item",
			setup: func(l *TodoList) uuid.UUID {
				return uuid.Nil
			},
			itemID: func(l *TodoList) uuid.UUID {
				return uuid.New()
			},
			wantErr: ErrItemNotFound,
		},
		{
			name: "remove from multiple items",
			setup: func(l *TodoList) uuid.UUID {
				l.AddItem("Item 1")
				item, _ := l.AddItem("Item 2")
				l.AddItem("Item 3")
				return item.ItemID
			},
			itemID: func(l *TodoList) uuid.UUID {
				return l.Items[1].ItemID
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewTodoList(uuid.New())
			tt.setup(list)
			itemID := tt.itemID(list)

			originalCount := len(list.Items)
			originalUpdatedAt := list.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			err := list.RemoveItem(itemID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("RemoveItem() error = %v, want %v", err, tt.wantErr)
				}
				if len(list.Items) != originalCount {
					t.Error("Items should not be modified on error")
				}
				return
			}

			if err != nil {
				t.Errorf("RemoveItem() unexpected error: %v", err)
				return
			}

			if len(list.Items) != originalCount-1 {
				t.Errorf("Items length = %d, want %d", len(list.Items), originalCount-1)
			}

			if list.findItem(itemID) != -1 {
				t.Error("Item should be removed from list")
			}

			if !list.UpdatedAt.After(originalUpdatedAt) {
				t.Error("RemoveItem() should update list timestamp")
			}
		})
	}
}

func TestTodoListFindItem(t *testing.T) {
	list := NewTodoList(uuid.New())
	item1, _ := list.AddItem("Item 1")
	item2, _ := list.AddItem("Item 2")
	item3, _ := list.AddItem("Item 3")

	tests := []struct {
		name     string
		itemID   uuid.UUID
		wantIdx  int
	}{
		{
			name:    "first item",
			itemID:  item1.ItemID,
			wantIdx: 0,
		},
		{
			name:    "middle item",
			itemID:  item2.ItemID,
			wantIdx: 1,
		},
		{
			name:    "last item",
			itemID:  item3.ItemID,
			wantIdx: 2,
		},
		{
			name:    "non-existent item",
			itemID:  uuid.New(),
			wantIdx: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := list.findItem(tt.itemID)

			if idx != tt.wantIdx {
				t.Errorf("findItem() = %d, want %d", idx, tt.wantIdx)
			}
		})
	}
}

// Helper functions for test data
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
