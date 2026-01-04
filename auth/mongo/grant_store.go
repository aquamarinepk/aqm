package mongo

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type grantStore struct {
	grantsColl *mongo.Collection
	rolesColl  *mongo.Collection
}

func NewGrantStore(grantsColl, rolesColl *mongo.Collection) auth.GrantStore {
	return &grantStore{
		grantsColl: grantsColl,
		rolesColl:  rolesColl,
	}
}

func (s *grantStore) Create(ctx context.Context, grant *auth.Grant) error {
	_, err := s.grantsColl.InsertOne(ctx, grant)
	if err != nil {
		return err
	}
	return nil
}

func (s *grantStore) Delete(ctx context.Context, userID, roleID uuid.UUID) error {
	filter := bson.M{"user_id": userID, "role_id": roleID}
	result, err := s.grantsColl.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return auth.ErrGrantNotFound
	}
	return nil
}

func (s *grantStore) GetUserGrants(ctx context.Context, userID uuid.UUID) ([]*auth.Grant, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "assigned_at", Value: -1}})
	cursor, err := s.grantsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var grants []*auth.Grant
	if err := cursor.All(ctx, &grants); err != nil {
		return nil, err
	}
	return grants, nil
}

func (s *grantStore) GetRoleGrants(ctx context.Context, roleID uuid.UUID) ([]*auth.Grant, error) {
	filter := bson.M{"role_id": roleID}
	opts := options.Find().SetSort(bson.D{{Key: "assigned_at", Value: -1}})
	cursor, err := s.grantsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var grants []*auth.Grant
	if err := cursor.All(ctx, &grants); err != nil {
		return nil, err
	}
	return grants, nil
}

func (s *grantStore) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*auth.Role, error) {
	// First get all grants for this user
	filter := bson.M{"user_id": userID}
	cursor, err := s.grantsColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var grants []*auth.Grant
	if err := cursor.All(ctx, &grants); err != nil {
		return nil, err
	}

	if len(grants) == 0 {
		return []*auth.Role{}, nil
	}

	// Extract role IDs
	roleIDs := make([]uuid.UUID, len(grants))
	for i, g := range grants {
		roleIDs[i] = g.RoleID
	}

	// Fetch roles
	roleFilter := bson.M{"_id": bson.M{"$in": roleIDs}}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	roleCursor, err := s.rolesColl.Find(ctx, roleFilter, opts)
	if err != nil {
		return nil, err
	}
	defer roleCursor.Close(ctx)

	var roles []*auth.Role
	if err := roleCursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *grantStore) HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	// First find the role by name
	roleFilter := bson.M{"name": roleName}
	role := &auth.Role{}
	err := s.rolesColl.FindOne(ctx, roleFilter).Decode(role)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check if grant exists
	grantFilter := bson.M{"user_id": userID, "role_id": role.ID}
	count, err := s.grantsColl.CountDocuments(ctx, grantFilter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

var _ auth.GrantStore = (*grantStore)(nil)
