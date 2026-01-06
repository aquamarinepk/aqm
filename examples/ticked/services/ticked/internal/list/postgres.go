package list

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list/sqlcgen"
	"github.com/google/uuid"
)

// postgresStore implements TodoListStore using PostgreSQL with sqlc-generated queries.
//
// DESIGN DECISION: Relational tables vs JSONB
// ============================================
// This implementation uses separate tables (todo_lists + todo_items) instead
// of storing items as JSONB in a single table. While JSONB would be simpler
// for this use case, we chose the relational approach because:
//
//   1. PORTABILITY: SQLite has no JSONB support (only TEXT with manual parsing)
//   2. REFERENCE: Demonstrates proper aggregate handling pattern for aqm
//   3. SCALABILITY: Works for complex aggregates with multiple child entities
//   4. PERFORMANCE: Enables indexes on item fields vs full document scans
//   5. TYPE SAFETY: sqlc generates type-safe Go code from SQL queries
//
// This pattern is portable across PostgreSQL, MySQL, and SQLite, making it
// the recommended approach for aqm-based applications.
//
// Schema:
//   todo_lists: id, user_id, created_at, updated_at
//   todo_items: id, list_id (FK), text, completed, created_at, completed_at
type postgresStore struct {
	db      *sql.DB
	queries *sqlcgen.Queries
}

// NewPostgresStore creates a PostgreSQL-backed store for TodoList aggregates.
func NewPostgresStore(db *sql.DB) TodoListStore {
	return &postgresStore{
		db:      db,
		queries: sqlcgen.New(db),
	}
}

// Save persists a TodoList aggregate using a transactional approach with sqlc.
//
// This demonstrates the aggregate pattern:
//   1. Start transaction
//   2. Upsert root entity (todo_lists) using sqlc
//   3. Sync child entities (todo_items) - diff algorithm with sqlc queries
//   4. Commit transaction
//
// The diff algorithm compares current vs persisted state to determine
// which items to INSERT, UPDATE, or DELETE, ensuring data consistency.
func (s *postgresStore) Save(ctx context.Context, list *TodoList) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create queries with transaction
	qtx := s.queries.WithTx(tx)

	// Step 1: Upsert root entity using sqlc
	if err := qtx.UpsertTodoList(ctx, sqlcgen.UpsertTodoListParams{
		ID:        list.ListID,
		UserID:    list.UserID,
		CreatedAt: list.CreatedAt,
		UpdatedAt: list.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("upsert root: %w", err)
	}

	// Step 2: Sync child entities (diff algorithm with sqlc)
	if err := s.syncItems(ctx, qtx, list.ListID, list.Items); err != nil {
		return fmt.Errorf("sync items: %w", err)
	}

	// Step 3: Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// FindByUserID loads a complete TodoList aggregate by user ID using sqlc.
//
// This demonstrates aggregate reconstruction:
//   1. Load root entity from todo_lists (sqlc query)
//   2. Load child entities from todo_items (sqlc query)
//   3. Reconstruct complete aggregate in memory
func (s *postgresStore) FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	// Step 1: Load root using sqlc
	dbList, err := s.queries.GetTodoListByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Step 2: Load children using sqlc
	dbItems, err := s.queries.GetTodoItemsByListID(ctx, dbList.ID)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}

	// Step 3: Reconstruct aggregate from sqlc models
	list := &TodoList{
		ListID:    dbList.ID,
		UserID:    dbList.UserID,
		Items:     make([]TodoItem, 0, len(dbItems)),
		CreatedAt: dbList.CreatedAt,
		UpdatedAt: dbList.UpdatedAt,
	}

	for _, dbItem := range dbItems {
		var completedAt *time.Time
		if dbItem.CompletedAt.Valid {
			completedAt = &dbItem.CompletedAt.Time
		}

		list.Items = append(list.Items, TodoItem{
			ItemID:      dbItem.ID,
			Text:        dbItem.Text,
			Completed:   dbItem.Completed,
			CreatedAt:   dbItem.CreatedAt,
			CompletedAt: completedAt,
		})
	}

	return list, nil
}

// Delete removes a TodoList aggregate using sqlc.
//
// NOTE: With proper foreign key constraints (ON DELETE CASCADE),
// deleting the root automatically deletes all children.
func (s *postgresStore) Delete(ctx context.Context, listID uuid.UUID) error {
	if err := s.queries.DeleteTodoList(ctx, listID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// syncItems performs a diff between current items and persisted items using sqlc,
// then executes the necessary INSERT/UPDATE/DELETE operations.
//
// Diff Algorithm:
//   1. Load existing items from DB (sqlc query)
//   2. Build maps for fast lookup (existingByID)
//   3. Process current items: INSERT new, UPDATE existing (sqlc)
//   4. DELETE items not in current set (sqlc)
//
// TODO: This could be extracted to aqm as a reusable primitive:
//       aggregate.SyncChildren[T](ctx, tx, parentID, current, load, insert, update, delete)
func (s *postgresStore) syncItems(ctx context.Context, qtx *sqlcgen.Queries, listID uuid.UUID, items []TodoItem) error {
	// Load existing items using sqlc
	existing, err := qtx.GetTodoItemsByListID(ctx, listID)
	if err != nil {
		return err
	}

	// Build lookup map
	existingByID := make(map[uuid.UUID]sqlcgen.TodoItem)
	for _, item := range existing {
		existingByID[item.ID] = item
	}

	// Track which items we've seen (to identify deletes)
	seen := make(map[uuid.UUID]bool)

	// Process current items: INSERT or UPDATE using sqlc
	for _, item := range items {
		seen[item.ItemID] = true

		if _, exists := existingByID[item.ItemID]; exists {
			// UPDATE existing item using sqlc
			completedAt := sql.NullTime{}
			if item.CompletedAt != nil {
				completedAt = sql.NullTime{Time: *item.CompletedAt, Valid: true}
			}

			if err := qtx.UpdateTodoItem(ctx, sqlcgen.UpdateTodoItemParams{
				ID:          item.ItemID,
				Text:        item.Text,
				Completed:   item.Completed,
				CompletedAt: completedAt,
			}); err != nil {
				return err
			}
		} else {
			// INSERT new item using sqlc
			completedAt := sql.NullTime{}
			if item.CompletedAt != nil {
				completedAt = sql.NullTime{Time: *item.CompletedAt, Valid: true}
			}

			if err := qtx.InsertTodoItem(ctx, sqlcgen.InsertTodoItemParams{
				ID:          item.ItemID,
				ListID:      listID,
				Text:        item.Text,
				Completed:   item.Completed,
				CreatedAt:   item.CreatedAt,
				CompletedAt: completedAt,
			}); err != nil {
				return err
			}
		}
	}

	// DELETE items that are no longer present using sqlc
	for id := range existingByID {
		if !seen[id] {
			if err := qtx.DeleteTodoItem(ctx, id); err != nil {
				return err
			}
		}
	}

	return nil
}

var _ TodoListStore = (*postgresStore)(nil)
