package list

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrItemNotFound   = errors.New("item not found")
	ErrItemTextEmpty  = errors.New("item text cannot be empty")
	ErrItemTextTooLong = errors.New("item text exceeds maximum length")
)

const maxItemTextLength = 500

// TodoList represents a user's todo list aggregate.
type TodoList struct {
	ListID    uuid.UUID  `json:"id" bson:"_id"`
	UserID    uuid.UUID  `json:"user_id" bson:"user_id"`
	Items     []TodoItem `json:"items" bson:"items"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
}

// TodoItem represents an item within a todo list.
type TodoItem struct {
	ItemID      uuid.UUID  `json:"id" bson:"id"`
	Text        string     `json:"text" bson:"text"`
	Completed   bool       `json:"completed" bson:"completed"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

// ID satisfies the Identifiable interface.
func (l *TodoList) ID() uuid.UUID {
	return l.ListID
}

// Touch updates the UpdatedAt timestamp.
func (l *TodoList) Touch() {
	l.UpdatedAt = time.Now().UTC()
}

// SortByCreatedAt sorts items by creation date, newest first.
func (l *TodoList) SortByCreatedAt() {
	sort.Slice(l.Items, func(i, j int) bool {
		return l.Items[i].CreatedAt.After(l.Items[j].CreatedAt)
	})
}

// NewTodoList creates a new todo list for a user.
func NewTodoList(userID uuid.UUID) *TodoList {
	now := time.Now().UTC()
	return &TodoList{
		ListID:    uuid.New(),
		UserID:    userID,
		Items:     []TodoItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddItem adds a new item to the list.
func (l *TodoList) AddItem(text string) (*TodoItem, error) {
	trimmed := strings.TrimSpace(text)

	if trimmed == "" {
		return nil, ErrItemTextEmpty
	}

	if len(trimmed) > maxItemTextLength {
		return nil, ErrItemTextTooLong
	}

	now := time.Now().UTC()
	item := TodoItem{
		ItemID:    uuid.New(),
		Text:      trimmed,
		Completed: false,
		CreatedAt: now,
	}

	l.Items = append(l.Items, item)
	l.Touch()

	return &item, nil
}

// UpdateItem updates an existing item in the list.
func (l *TodoList) UpdateItem(itemID uuid.UUID, text *string, completed *bool) error {
	idx := l.findItem(itemID)
	if idx == -1 {
		return ErrItemNotFound
	}

	if text != nil {
		trimmed := strings.TrimSpace(*text)
		if trimmed == "" {
			return ErrItemTextEmpty
		}
		if len(trimmed) > maxItemTextLength {
			return ErrItemTextTooLong
		}
		l.Items[idx].Text = trimmed
	}

	if completed != nil {
		l.Items[idx].Completed = *completed
		if *completed && l.Items[idx].CompletedAt == nil {
			now := time.Now().UTC()
			l.Items[idx].CompletedAt = &now
		} else if !*completed {
			l.Items[idx].CompletedAt = nil
		}
	}

	l.Touch()
	return nil
}

// RemoveItem removes an item from the list.
func (l *TodoList) RemoveItem(itemID uuid.UUID) error {
	idx := l.findItem(itemID)
	if idx == -1 {
		return ErrItemNotFound
	}

	l.Items = append(l.Items[:idx], l.Items[idx+1:]...)
	l.Touch()
	return nil
}

// findItem finds the index of an item by ID.
func (l *TodoList) findItem(itemID uuid.UUID) int {
	for i, item := range l.Items {
		if item.ItemID == itemID {
			return i
		}
	}
	return -1
}
