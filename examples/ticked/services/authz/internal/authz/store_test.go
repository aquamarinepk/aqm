package authz

import (
	"testing"
)

func TestNewFakeStores(t *testing.T) {
	roleStore, grantStore := NewFakeStores()

	if roleStore == nil {
		t.Error("roleStore is nil")
	}

	if grantStore == nil {
		t.Error("grantStore is nil")
	}
}

func TestNewPostgresStores(t *testing.T) {
	tests := []struct {
		name       string
		connStr    string
		wantErr    bool
		skipReason string
	}{
		{
			name:       "invalid connection string",
			connStr:    "invalid",
			wantErr:    true,
			skipReason: "",
		},
		{
			name:       "valid connection but unreachable",
			connStr:    "host=nonexistent port=5432 user=test password=test dbname=test sslmode=disable",
			wantErr:    true,
			skipReason: "postgres not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore, grantStore, db, err := NewPostgresStores(tt.connStr)

			if tt.skipReason != "" && err != nil {
				t.Skipf("skipping: %s", tt.skipReason)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("NewPostgresStores() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPostgresStores() unexpected error: %v", err)
				return
			}

			if roleStore == nil {
				t.Error("roleStore is nil")
			}

			if grantStore == nil {
				t.Error("grantStore is nil")
			}

			if db == nil {
				t.Error("db is nil")
			}

			if db != nil {
				db.Close()
			}
		})
	}
}
