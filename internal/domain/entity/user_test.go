package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	user, err := NewUser("test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", user.Name)
	}
	if user.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if user.Password == "password123" {
		t.Error("password should be hashed, not plaintext")
	}
}

func TestCheckPassword(t *testing.T) {
	user, err := NewUser("test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !user.CheckPassword("password123") {
		t.Error("expected correct password to match")
	}
	if user.CheckPassword("wrongpassword") {
		t.Error("expected wrong password to not match")
	}
}
