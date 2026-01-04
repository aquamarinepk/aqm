package mongo

import (
	"context"
	"os"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestMongo(t *testing.T) (*mongo.Collection, func()) {
	t.Helper()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	db := client.Database("test_auth")
	coll := db.Collection("users")

	cleanup := func() {
		coll.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	return coll, cleanup
}

func TestMongoUserStoreCreate(t *testing.T) {
	coll, cleanup := setupTestMongo(t)
	defer cleanup()

	store := NewUserStore(coll)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "testuser"
	user.Name = "Test User"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup123")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.Status = auth.UserStatusActive
	user.BeforeCreate()

	err := store.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	retrieved, err := store.Get(ctx, user.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Username != user.Username {
		t.Errorf("Username = %v, want %v", retrieved.Username, user.Username)
	}
}

func TestMongoUserStoreGetByUsername(t *testing.T) {
	coll, cleanup := setupTestMongo(t)
	defer cleanup()

	store := NewUserStore(coll)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "findme"
	user.Name = "Find Me"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup_findme")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "existing username",
			username: "findme",
			wantErr:  nil,
		},
		{
			name:     "non-existing username",
			username: "notfound",
			wantErr:  auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByUsername(ctx, tt.username)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByUsername() unexpected error = %v", err)
			}

			if got.Username != tt.username {
				t.Errorf("GetByUsername() username = %v, want %v", got.Username, tt.username)
			}
		})
	}
}

func TestMongoUserStoreUpdate(t *testing.T) {
	coll, cleanup := setupTestMongo(t)
	defer cleanup()

	store := NewUserStore(coll)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "updateme"
	user.Name = "Original Name"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup_updateme")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	user.Name = "Updated Name"
	user.BeforeUpdate()

	err := store.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	retrieved, _ := store.Get(ctx, user.ID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("Update() name = %v, want Updated Name", retrieved.Name)
	}
}
