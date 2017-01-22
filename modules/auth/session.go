package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
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
	hashKey         []byte
	cipherKey       []byte
}

func NewSessionHandler(sessionStore *SessionStore) *SessionHandler {
	return &SessionHandler{
		sessionStore:    sessionStore,
		sessionDuration: 5 * time.Minute, // TODO: Configure duration and keys
		hashKey:         make([]byte, 32),
		cipherKey:       make([]byte, 32),
	}
}

func (h *SessionHandler) GetSessionCookie(w http.ResponseWriter, r *http.Request) (*steamUser, bool) {
	// TODO: Decrypt/Validate the cookie
	var sessionCookie *http.Cookie
	for _, c := range r.Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie != nil {
		value, err := h.decrypt(sessionCookie.Value)
		if err != nil {
			return nil, false
		}

		session, ok := h.sessionStore.Fetch(value)
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
	session, err := NewSession(h.sessionDuration)
	if err != nil {
		return errors.New("could not create session")
	}

	session.Value = profile
	h.sessionStore.Save(session)

	value, err := h.encrypt(session.ID)
	if err != nil {
		return errors.New("could not encrypt session id")
	}

	sessionCookie := &http.Cookie{Name: sessionCookieName, Value: value, Path: "/", Expires: session.expiration, HttpOnly: true}
	http.SetCookie(w, sessionCookie)

	return nil
}

func (h *SessionHandler) encrypt(value string) (string, error) {
	hash := hmac.New(sha256.New, h.hashKey)
	c, err := aes.NewCipher(h.cipherKey)
	if err != nil {
		return "", err
	}

	// serialize value
	v := []byte(value)

	// generate IV
	iv := make([]byte, c.BlockSize())
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}

	// encrypt value
	stream := cipher.NewCTR(c, iv)
	stream.XORKeyStream(v, v)

	// merge IV and value
	v = append(iv, v...)

	// create HMAC
	hash.Write(v)
	mac := hash.Sum(nil)

	// merge HMAC and value
	v = append(mac, v...)

	// encode
	v64 := base64.URLEncoding.EncodeToString(v)

	return v64, nil
}

func (h *SessionHandler) decrypt(value string) (string, error) {
	hash := hmac.New(sha256.New, h.hashKey)
	c, err := aes.NewCipher(h.cipherKey)
	if err != nil {
		return "", err
	}

	// decode
	v, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	// split HMAC and value
	if len(v) < hash.Size() {
		return "", errors.New("missing HMAC")
	}

	expectedMac := v[:hash.Size()]
	v = v[hash.Size():]

	// validate HMAC
	hash.Write(v)
	mac := hash.Sum(nil)

	if subtle.ConstantTimeCompare(expectedMac, mac) != 1 {
		return "", errors.New("HMAC mismatch")
	}

	// split IV and value
	ivLen := c.BlockSize()
	if len(v) < ivLen {
		return "", errors.New("missing IV")
	}

	iv := v[:ivLen]
	v = v[ivLen:]

	// decrypt value
	s := cipher.NewCTR(c, iv)
	s.XORKeyStream(v, v)

	// deserialize value
	val := string(v)

	return val, nil
}
