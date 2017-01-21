package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"
)

const sessionIDSize = 16

type Session struct {
	ID         string
	Value      interface{}
	createdAt  time.Time
	expiration time.Time
}

func NewSession(duration time.Duration) (Session, error) {
	ID, err := genSessionID()
	if err != nil {
		return Session{}, err
	}

	var session Session
	session.ID = ID
	session.createdAt = time.Now().UTC()
	session.expiration = session.createdAt.Add(duration)
	return session, nil
}

func genSessionID() (string, error) {
	var ID [sessionIDSize]byte

	_, err := rand.Read(ID[:])
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ID[:]), nil
}

type SessionStore struct {
	store map[string]Session
	mutex sync.RWMutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		store: make(map[string]Session),
	}
}

func (s *SessionStore) Fetch(ID string) (Session, bool) {
	s.mutex.RLock()
	session, ok := s.store[ID]
	s.mutex.RUnlock()

	if !ok {
		return session, false
	}

	if session.expiration.Before(time.Now().UTC()) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		delete(s.store, ID)
		return session, false
	}

	return session, ok
}

func (s *SessionStore) Save(session Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[session.ID] = session
}

const sessionCookieName = "session"

type SessionHandler struct {
	sessionStore    *SessionStore
	sessionDuration time.Duration
}

func NewSessionHandler(sessionStore *SessionStore) *SessionHandler {
	return &SessionHandler{
		sessionStore:    sessionStore,
		sessionDuration: 5 * time.Minute, // TODO: Configure
	}
}

func (h *SessionHandler) Handle(w http.ResponseWriter, r *http.Request) (*steamUser, bool) {
	// TODO: Decrypt/Validate the cookie
	var sessionCookie *http.Cookie
	for _, c := range r.Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie != nil {
		session, ok := h.sessionStore.Fetch(sessionCookie.Value)
		if ok {
			user, ok := session.Value.(*steamUser)
			if ok {
				return user, true
			}
		}
	}

	return nil, false
}

func (h *SessionHandler) SetSessionCookie(w http.ResponseWriter, profile *steamUser) error {
	// TODO: Encrypt/Sign the cookie
	session, err := NewSession(h.sessionDuration)
	if err != nil {
		return errors.New("could not create session")
	}

	session.Value = profile
	h.sessionStore.Save(session)
	sessionCookie := &http.Cookie{Name: sessionCookieName, Value: session.ID, Path: "/", Expires: session.expiration, HttpOnly: true}
	http.SetCookie(w, sessionCookie)

	return nil
}
