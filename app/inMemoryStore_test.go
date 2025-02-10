package main

import (
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	store := NewInMemoryStore()

	// Teste básico de inserção e recuperação
	store.Set("foo", "bar", 0)
	value, err := store.Get("foo")
	if err != nil || value != "bar" {
		t.Fatalf("Erro ao obter valor esperado. Obtido: %v, erro: %v", value, err)
	}
}

func TestExpiration(t *testing.T) {
	store := NewInMemoryStore()

	// Definir chave com TTL curto
	store.Set("temp", "expired", 100)
	time.Sleep(200 * time.Millisecond)

	// Deve estar expirado
	_, err := store.Get("temp")
	if err == nil {
		t.Fatalf("Esperava erro de chave expirada, mas chave ainda existe")
	}
}

func TestDelete(t *testing.T) {
	store := NewInMemoryStore()

	store.Set("deleteMe", "exists", 0)
	store.Delete("deleteMe")

	_, err := store.Get("deleteMe")
	if err == nil {
		t.Fatalf("Esperava erro de chave inexistente, mas chave ainda existe")
	}
}
