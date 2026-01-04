package mongo

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userStore struct {
	coll *mongo.Collection
}

func NewUserStore(coll *mongo.Collection) auth.UserStore {
	return &userStore{coll: coll}
}

func (s *userStore) Create(ctx context.Context, user *auth.User) error {
	_, err := s.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *userStore) Get(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	filter := bson.M{"_id": id}
	user := &auth.User{}
	err := s.coll.FindOne(ctx, filter).Decode(user)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByEmailLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	filter := bson.M{"email_lookup": lookup}
	user := &auth.User{}
	err := s.coll.FindOne(ctx, filter).Decode(user)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	filter := bson.M{"username": username}
	user := &auth.User{}
	err := s.coll.FindOne(ctx, filter).Decode(user)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByPINLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	filter := bson.M{"pin_lookup": lookup}
	user := &auth.User{}
	err := s.coll.FindOne(ctx, filter).Decode(user)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) Update(ctx context.Context, user *auth.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	result, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

func (s *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "deleted", "updated_at": bson.M{"$currentDate": true}}}
	result, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

func (s *userStore) List(ctx context.Context) ([]*auth.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*auth.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userStore) ListByStatus(ctx context.Context, status auth.UserStatus) ([]*auth.User, error) {
	filter := bson.M{"status": status}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*auth.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

var _ auth.UserStore = (*userStore)(nil)
