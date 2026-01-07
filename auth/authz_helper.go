package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
)

type AuthzHelper struct {
	client   *httpclient.Client
	cache    *permissionCache
	cacheTTL time.Duration
	log      log.Logger
}

func NewAuthzHelper(authzURL string, cacheTTL time.Duration, log log.Logger) *AuthzHelper {
	return &AuthzHelper{
		client:   httpclient.New(authzURL, log),
		cache:    newPermissionCache(),
		cacheTTL: cacheTTL,
		log:      log,
	}
}

func (h *AuthzHelper) CheckPermission(ctx context.Context, userID, permission, resource string) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s", userID, permission, resource)

	if cached, ok := h.cache.get(cacheKey); ok {
		return cached, nil
	}

	resp, err := h.client.Post(ctx, "/authz/check", map[string]string{
		"user_id":    userID,
		"permission": permission,
		"resource":   resource,
	})
	if err != nil {
		return false, fmt.Errorf("authz check request: %w", err)
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := resp.JSON(&result); err != nil {
		return false, fmt.Errorf("parse authz response: %w", err)
	}

	h.cache.set(cacheKey, result.Allowed, h.cacheTTL)
	return result.Allowed, nil
}

type permissionCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	allowed   bool
	expiresAt time.Time
}

func newPermissionCache() *permissionCache {
	return &permissionCache{
		items: make(map[string]*cacheItem),
	}
}

func (c *permissionCache) get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return false, false
	}

	return item.allowed, true
}

func (c *permissionCache) set(key string, allowed bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		allowed:   allowed,
		expiresAt: time.Now().Add(ttl),
	}
}
