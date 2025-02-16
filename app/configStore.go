package main

import (
	"fmt"
	"sync"
)

type ConfigStore struct {
	data map[string]string
	mu   sync.Mutex
}

func NewConfigStore() *ConfigStore {
	store := &ConfigStore{
		data: make(map[string]string),
	}
	return store
}

func (store *ConfigStore) Init(dir string, dbfilename string) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Init:", r)
		}
	}()

	if dir == "" || dbfilename == "" {
		return false, fmt.Errorf("-ERR ⚠️ Invalid dir or dbfilename. Both must be non-empty")
	}
	_, errDir := store.Set("dir", dir)
	if errDir != nil {
		return false, fmt.Errorf("⚠️ Failed to set dir: %s", errDir.Error())
	}

	_, errFilename := store.Set("dbfilename", dbfilename)
	if errFilename != nil {
		return false, fmt.Errorf("⚠️ Failed to set dbfilename: %s", errFilename.Error())
	}

	return true, nil
}

func (store *ConfigStore) Set(key string, value string) (string, error) {
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

	fmt.Printf("✅ Added new config: [%s] -> %s\n", key, value)
	return fmt.Sprintf("✅ Added new config: [%s] -> %s", key, value), nil
}

func (store *ConfigStore) Get(key string) (string, error) {
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

func (store *ConfigStore) Delete(keys []string) (int, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in Delete:", r)
		}
	}()

	if len(keys) < 1 {
		fmt.Println("⚠️ Invalid key. Must be non-empty")
		return 0, fmt.Errorf("-ERR Invalid key. Must be non-empty")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	count := 0
	for _, key := range keys {
		if _, exists := store.data[key]; !exists {
			fmt.Println("⚠️ Key not found: ", key)
			continue
		}
		delete(store.data, key)
		count++
		fmt.Println("✅ Deleted key: ", key)
	}

	return count, nil
}

func (store *ConfigStore) ListConfig() ([]string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ Panic recovered in ListConfig:", r)
		}
	}()

	store.mu.Lock()
	defer store.mu.Unlock()

	keys := make([]string, 0, len(store.data))
	for key, value := range store.data {
		keys = append(keys, fmt.Sprintf("[%s] -> %s", key, value))
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("-ERR ⚠️ No keys found. Config is empty")
	}

	return keys, nil
}
