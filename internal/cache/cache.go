package cache

import (
	"sync"
	"time"
	"wb-order-service/internal/model"
)

type entry struct {
	val *model.Order
	exp time.Time
}

type Cache interface {
	Get(key string) (*model.Order, bool)
	Set(key string, val *model.Order, ttl time.Duration)
	Stop()
}

type TTLCache struct {
	mu sync.RWMutex
	m  map[string]entry
	stop chan struct{}
}

func NewTTL() *TTLCache {
	c := &TTLCache{m: make(map[string]entry), stop: make(chan struct{})}
	return c
}

func (c *TTLCache) StartJanitor(interval time.Duration) {
	if interval <= 0 { interval = time.Minute }
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				c.cleanup()
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *TTLCache) Stop() { close(c.stop) }

func (c *TTLCache) Get(key string) (*model.Order, bool) {
	c.mu.RLock(); defer c.mu.RUnlock()
	e, ok := c.m[key]
	if !ok || time.Now().After(e.exp) { return nil, false }
	return e.val, true
}

func (c *TTLCache) Set(key string, val *model.Order, ttl time.Duration) {
	if ttl <= 0 { ttl = 5 * time.Minute }
	c.mu.Lock(); defer c.mu.Unlock()
	c.m[key] = entry{val: val, exp: time.Now().Add(ttl)}
}

func (c *TTLCache) cleanup() {
	now := time.Now()
	c.mu.Lock(); defer c.mu.Unlock()
	for k, v := range c.m {
		if now.After(v.exp) { delete(c.m, k) }
	}
}
