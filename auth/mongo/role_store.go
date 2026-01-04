package mongo

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type roleStore struct {
	coll *mongo.Collection
}

func NewRoleStore(coll *mongo.Collection) auth.RoleStore {
	return &roleStore{coll: coll}
}

func (s *roleStore) Create(ctx context.Context, role *auth.Role) error {
	_, err := s.coll.InsertOne(ctx, role)
	if err != nil {
		return err
	}
	return nil
}

func (s *roleStore) Get(ctx context.Context, id uuid.UUID) (*auth.Role, error) {
	filter := bson.M{"_id": id}
	role := &auth.Role{}
	err := s.coll.FindOne(ctx, filter).Decode(role)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleStore) GetByName(ctx context.Context, name string) (*auth.Role, error) {
	filter := bson.M{"name": name}
	role := &auth.Role{}
	err := s.coll.FindOne(ctx, filter).Decode(role)
	if err == mongo.ErrNoDocuments {
		return nil, auth.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleStore) Update(ctx context.Context, role *auth.Role) error {
	filter := bson.M{"_id": role.ID}
	update := bson.M{"$set": role}
	result, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return auth.ErrRoleNotFound
	}
	return nil
}

func (s *roleStore) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "inactive", "updated_at": bson.M{"$currentDate": true}}}
	result, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return auth.ErrRoleNotFound
	}
	return nil
}

func (s *roleStore) List(ctx context.Context) ([]*auth.Role, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*auth.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *roleStore) ListByStatus(ctx context.Context, status auth.RoleStatus) ([]*auth.Role, error) {
	filter := bson.M{"status": status}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*auth.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

var _ auth.RoleStore = (*roleStore)(nil)
