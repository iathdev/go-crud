package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	user := NewUser(446416, "test@example.com", "Test User")

	if user.PrepUserID != 446416 {
		t.Errorf("expected prep user id 446416, got %d", user.PrepUserID)
	}
	if user.ID == uuid.Nil {
		t.Error("expected non-nil UUID for ID")
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", user.Name)
	}
}

func TestNewPrepUser(t *testing.T) {
	prepUser := NewPrepUser(446416, "test@example.com", "Test User")

	if prepUser.PrepUserID != 446416 {
		t.Errorf("expected prep user id 446416, got %d", prepUser.PrepUserID)
	}
	if prepUser.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", prepUser.Email)
	}
	if prepUser.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", prepUser.Name)
	}
}
