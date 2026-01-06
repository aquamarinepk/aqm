package list

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// TodoListStore abstracts persistence for todo list aggregates.
//
// Note: We use separate tables (todo_lists + todo_items) instead of JSONB
// for several reasons:
//   1. Portability: SQLite doesn't have JSONB, only TEXT (requires manual parsing)
//   2. Reference implementation: Demonstrates how aqm should handle aggregates
//   3. Performance: Proper indexes on items table vs full document scan
//
// For this simple case, JSONB would work fine in Postgres, but this pattern
// scales to complex aggregates and works across all SQL databases.
type TodoListStore interface {
	Save(ctx context.Context, list *TodoList) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	Delete(ctx context.Context, listID uuid.UUID) error
}

var ErrNotFound = errors.New("list not found")
