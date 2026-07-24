package main

import (
	"context"
	"sync"
	"time"
)

// modelCache memoizes the (potentially slow) model-discovery API calls so that
// repeated get_configuration / generate_image calls don't re-hit Google's API.
// This keeps the agent-facing tools fast and cheap.
type modelCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]cacheEntry
}

type cacheEntry struct {
	models  []string
	expires time.Time
}

func newModelCache(ttl time.Duration) *modelCache {
	return &modelCache{
		ttl:   ttl,
		items: make(map[string]cacheEntry),
	}
}

// Models returns the cached model list for a provider, refreshing via fetch()
// when the entry is missing or expired.
func (c *modelCache) Models(ctx context.Context, providerID string, fetch func() []string) []string {
	c.mu.Lock()
	if e, ok := c.items[providerID]; ok && time.Now().Before(e.expires) {
		c.mu.Unlock()
		return e.models
	}
	c.mu.Unlock()

	models := fetch()

	c.mu.Lock()
	c.items[providerID] = cacheEntry{models: models, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return models
}
