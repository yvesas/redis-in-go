package main

import (
	"fmt"
	"sync"
	"time"
)

type InMemoryStore struct {
	data    map[string]string
	expires map[string]int64
	mu      sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		data:    make(map[string]string),
		expires: make(map[string]int64),
	}
	go store.cleanupExpiredKeys()
	return store
}

func (store *InMemoryStore) cleanupExpiredKeys() {
	for {
		time.Sleep(500 * time.Millisecond)
		store.mu.Lock()
		now := time.Now().UnixMilli()
		for key, exp := range store.expires {
			if now > exp {
				delete(store.data, key)
				delete(store.expires, key)
			}
		}
		store.mu.Unlock()
	}
}

func (store *InMemoryStore) Set(key string, value string, ttlMs int64) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Set:", r)
		}
	}()

	if key == "" || value == "" {
		return "", fmt.Errorf("-ERR ⚠️ Invalid key or value. Both must be non-empty")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if existingValue, exists := store.data[key]; exists && existingValue == value {
		fmt.Println("⚠️ Key already exists with the same value. No changes made.")
		return "⚠️ Key already exists with the same value. No changes made.", nil
	}

	store.data[key] = value
	if ttlMs > 0 {
		store.expires[key] = time.Now().UnixMilli() + ttlMs
	} else {
		delete(store.expires, key)
	}

	fmt.Printf("✅ Added: [%s] -> %s", key, value)
	return fmt.Sprintf("✅ Added: [%s] -> %s", key, value), nil
}

func (store *InMemoryStore) Get(key string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Get:", r)
		}
	}()

	if key == "" {
		return "", fmt.Errorf("⚠️ Invalid key. Must be non-empty")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if exp, exists := store.expires[key]; exists && time.Now().UnixMilli() > exp {
		delete(store.data, key)
		delete(store.expires, key)
		return "", fmt.Errorf("⚠️ Key expired")
	}

	value, exists := store.data[key]
	if !exists {
		return "", fmt.Errorf("⚠️ Key not found")
	}

	return value, nil
}

func (store *InMemoryStore) Delete(key string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Delete:", r)
		}
	}()

	if key == "" {
		return "", fmt.Errorf("-ERR ⚠️ Invalid key. Must be non-empty")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.data[key]; !exists {
		return "", fmt.Errorf("-ERR ⚠️ Key not found")
	}

	delete(store.data, key)
	return fmt.Sprintf("✅ Deleted key: %s", key), nil
}

func (store *InMemoryStore) ListKeys() ([]string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in ListKeys:", r)
		}
	}()

	store.mu.Lock()
	defer store.mu.Unlock()

	keys := make([]string, 0, len(store.data))
	for key := range store.data {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("-ERR ⚠️ No keys found. Store is empty")
	}

	return keys, nil
}
