package auth

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	session, err := NewSession(1*time.Minute, &User{})
	if err != nil {
		t.Error(err)
		return
	}

	if len(session.ID) == 0 {
		t.Error("could not generate a valid session ID")
	}
}
