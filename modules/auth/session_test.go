package auth

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	session, err := NewSession(1 * time.Minute)
	if err != nil {
		t.Error(err)
		return
	}

	if len(session.ID) == 0 {
		t.Error("could not generate a valid session ID")
	}
}

func TestSessionStore_SaveFetch(t *testing.T) {
	store := NewSessionStore()
	s1, _ := NewSession(1 * time.Minute)
	s1.Value = "testvalue"

	store.Save(s1)

	s2, ok := store.Fetch(s1.ID)
	if !ok {
		t.Error("session not found")
		return
	}

	if s1.ID != s2.ID {
		t.Error("ID mismatch")
	}

	if s1.createdAt != s2.createdAt {
		t.Error("createdAt mismatch")
	}

	if s1.expiration != s2.expiration {
		t.Error("expiration mismatch")
	}

	if s1.Value != s2.Value {
		t.Error("value mismatch")
	}
}
