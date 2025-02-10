package main

import (
	"fmt"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	store := NewInMemoryStore()

	// Basic insert and retrieve test
	store.Set("foo", "bar", 0)
	value, err := store.Get("foo")
	if err != nil || value != "bar" {
		t.Fatalf("Expected value 'bar', but got: %v, error: %v", value, err)
	}
}

func TestSetWithTTL0(t *testing.T) {
	store := NewInMemoryStore()

	// Set key with no TTL (should persist indefinitely)
	store.Set("persist", "forever", 0)
	time.Sleep(200 * time.Millisecond) // Wait to ensure it does not expire
	value, err := store.Get("persist")
	if err != nil || value != "forever" {
		t.Fatalf("Expected 'forever' to persist, but got: %v, error: %v", value, err)
	}
}

func TestSetOverwrite(t *testing.T) {
	store := NewInMemoryStore()

	// Set a value and then overwrite it
	store.Set("overwrite", "first", 0)
	store.Set("overwrite", "second", 0)

	value, err := store.Get("overwrite")
	if err != nil || value != "second" {
		t.Fatalf("Expected 'second' after overwrite, but got: %v, error: %v", value, err)
	}
}

func TestSetInvalidKeyValue(t *testing.T) {
	store := NewInMemoryStore()

	// Test with empty key
	_, err := store.Set("", "value", 0)
	if err == nil {
		t.Fatalf("Expected an error when setting an empty key, but got none")
	}

	// Test with empty value
	_, err = store.Set("key", "", 0)
	if err == nil {
		t.Fatalf("Expected an error when setting an empty value, but got none")
	}
}

func TestGetNonExistentKey(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatalf("Expected an error when retrieving a non-existent key, but got none")
	}
}

func TestExpiration(t *testing.T) {
	store := NewInMemoryStore()

	// Set a key with a short TTL
	store.Set("temp", "expired", 100)
	time.Sleep(200 * time.Millisecond)

	// It should be expired
	_, err := store.Get("temp")
	if err == nil {
		t.Fatalf("Expected an error for expired key, but key still exists")
	}
}

func TestCleanupExpiredKeys(t *testing.T) {
	store := NewInMemoryStore()

	store.Set("shortLived", "tempValue", 100)
	time.Sleep(501 * time.Millisecond)

	store.mu.Lock()
	_, exists := store.data["shortLived"]
	store.mu.Unlock()

	fmt.Println("Key exists in store?", exists)

	if exists {
		t.Fatalf("Key 'shortLived' should have been removed by cleanupExpiredKeys, but it still exists")
	}
}

func TestDelete(t *testing.T) {
	store := NewInMemoryStore()

	store.Set("deleteMe", "exists", 0)
	store.Delete("deleteMe")

	_, err := store.Get("deleteMe")
	if err == nil {
		t.Fatalf("Expected an error for deleted key, but key still exists")
	}
}

func TestListKeys(t *testing.T) {
	store := NewInMemoryStore()

	store.Set("key1", "value1", 0)
	store.Set("key2", "value2", 0)

	keys, err := store.ListKeys()
	if err != nil {
		t.Fatalf("Unexpected error when listing keys: %v", err)
	}

	expectedKeys := map[string]bool{"key1": true, "key2": true}
	for _, key := range keys {
		if !expectedKeys[key] {
			t.Fatalf("Unexpected key found: %s", key)
		}
	}
}
