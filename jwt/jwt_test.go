package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestNewManagerRejectsEmptySecret(t *testing.T) {
	if _, err := NewManager(nil, time.Minute); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestGenerateAndVerify(t *testing.T) {
	m, err := NewManager([]byte("secret"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := m.GenerateToken("alice", "admin")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Username != "alice" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	a, _ := NewManager([]byte("a"), time.Minute)
	b, _ := NewManager([]byte("b"), time.Minute)
	token, _ := a.GenerateToken("alice", "user")
	if _, err := b.VerifyToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	m, _ := NewManager([]byte("a"), time.Nanosecond)
	token, _ := m.GenerateToken("alice", "user")
	time.Sleep(2 * time.Millisecond)
	if _, err := m.VerifyToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	m, _ := NewManager([]byte("a"), time.Minute)
	if _, err := m.VerifyToken("not.a.token"); err == nil {
		t.Fatal("expected error")
	}
}
