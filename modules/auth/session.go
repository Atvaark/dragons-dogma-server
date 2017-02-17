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
	"time"
)

const sessionIDSize = 16

type Session struct {
	ID         string
	User       *User
	CreatedAt  time.Time
	Expiration time.Time
}

func NewSession(duration time.Duration, user *User) (Session, error) {
	ID, err := genSessionID()
	if err != nil {
		return Session{}, err
	}

	var session Session
	session.ID = ID
	session.User = user
	session.CreatedAt = time.Now().UTC()
	session.Expiration = session.CreatedAt.Add(duration)
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

type Database interface {
	GetSession(ID string) (*Session, error)
	PutSession(*Session) error
	DeleteSession(ID string) error
}

const sessionCookieName = "session"

type SessionHandler struct {
	database        Database
	sessionDuration time.Duration
	hashKey         []byte
	cipherKey       []byte
}

func NewSessionHandler(database Database) *SessionHandler {
	return &SessionHandler{
		database:        database,
		sessionDuration: 5 * time.Minute, // TODO: Configure duration and keys
		hashKey:         make([]byte, 32),
		cipherKey:       make([]byte, 32),
	}
}

func (h *SessionHandler) GetSessionCookie(w http.ResponseWriter, r *http.Request) (*User, bool) {
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

		session, err := h.database.GetSession(value)
		if err != nil {
			return nil, false
		}

		if session.Expiration.Before(time.Now().UTC()) {
			return nil, false
		}

		if session.User != nil {
			return session.User, true
		}
	}

	return nil, false
}

func (h *SessionHandler) SetSessionCookie(w http.ResponseWriter, user *User) error {
	session, err := NewSession(h.sessionDuration, user)
	if err != nil {
		return errors.New("could not create session")
	}

	err = h.database.PutSession(&session)
	if err != nil {
		return errors.New("could not save session")
	}

	value, err := h.encrypt(session.ID)
	if err != nil {
		return errors.New("could not encrypt session id")
	}

	sessionCookie := &http.Cookie{Name: sessionCookieName, Value: value, Path: "/", Expires: session.Expiration, HttpOnly: true}
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
