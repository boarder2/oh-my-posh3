package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

type cacheEntry struct {
	Value      string
	Expiration int64
}

// Returns true if the item has expired.
func (ce cacheEntry) Expired() bool {
	if ce.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > ce.Expiration
}

type simplecache struct {
	defaultExpiration time.Duration
	items             map[string]cacheEntry
	mu                sync.RWMutex
}

func (c *simplecache) get(key string) (string, bool) {
	item, found := c.items[key]
	if !found {
		return "", false
	}
	// "Inlining" of Expired
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return "", false
		}
	}
	return item.Value, true
}

func (c *simplecache) put(key, value string, expiration time.Duration) {
	var e int64
	if expiration == DefaultExpiration {
		expiration = c.defaultExpiration
	}
	if expiration > 0 {
		e = time.Now().Add(expiration).UnixNano()
	}
	c.mu.Lock()
	c.items[key] = cacheEntry{
		Value:      value,
		Expiration: e,
	}
	c.mu.Unlock()
}

func newCache(de time.Duration, m map[string]cacheEntry) *simplecache {
	if de == 0 {
		de = -1
	}
	c := &simplecache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func New(defaultExpiration time.Duration) *simplecache {
	items := make(map[string]cacheEntry)
	return newCache(defaultExpiration, items)
}

func (c *simplecache) loadFromFile(filePath string) error {
	fp, err := os.Open(filePath)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(fp)
	items := make(map[string]cacheEntry)
	err = dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, v := range items {
			ov, found := c.items[k]
			if !found || ov.Expired() {
				c.items[k] = v
			}
		}
	}
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func (c *simplecache) saveToFile(filePath string) error {
	fp, err := os.Create(filePath)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(fp)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		gob.Register(v.Value)
	}
	err = enc.Encode(&c.items)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}
