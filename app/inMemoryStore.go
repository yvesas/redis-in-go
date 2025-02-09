package main

import (
	"fmt"
	"sync"
)

type InMemoryStore struct {
	data map[string]string
	mu   sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

func (store *InMemoryStore) Set(key string, value string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Set:", r)
		}
	}()

	if key == "" || value == "" {
		return "", fmt.Errorf("⚠️ Invalid key or value. Both must be non-empty.")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if existingValue, exists := store.data[key]; exists && existingValue == value {
		return "⚠️ Key already exists with the same value. No changes made.", nil
	}

	store.data[key] = value
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
		return "", fmt.Errorf("⚠️ Invalid key. Must be non-empty.")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.data[key]; !exists {
		return "", fmt.Errorf("⚠️ Key not found.")
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
		return nil, fmt.Errorf("⚠️ No keys found. Store is empty.")
	}

	return keys, nil
}
