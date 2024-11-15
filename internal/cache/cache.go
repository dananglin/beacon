// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package cache

import (
	"maps"
	"sync"
	"time"
)

type Cache struct {
	mu              *sync.Mutex
	entries         map[string]Entry
	cleanupInterval time.Duration
}

func NewCache(cleanupInterval time.Duration) *Cache {
	cache := Cache{
		mu:              &sync.Mutex{},
		entries:         make(map[string]Entry),
		cleanupInterval: cleanupInterval,
	}

	go cache.readLoop()

	return &cache
}

func (c *Cache) Add(key string, val []byte, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := Entry{
		expiresAt: expiresAt,
		val:       val,
	}

	c.entries[key] = entry
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

func (c *Cache) Get(key string) (Entry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, exist := c.entries[key]

	return value, exist
}

func (c *Cache) readLoop() {
	ticker := time.Tick(c.cleanupInterval)

	for range ticker {
		c.cleanupEntries()
	}
}

func (c *Cache) cleanupEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range maps.All(c.entries) {
		if c.entries[key].Expired() {
			delete(c.entries, key)
		}
	}
}

type Entry struct {
	expiresAt time.Time
	val       []byte
}

func (e Entry) Value() []byte {
	return e.val
}

func (e Entry) Expired() bool {
	return time.Now().After(e.expiresAt)
}
