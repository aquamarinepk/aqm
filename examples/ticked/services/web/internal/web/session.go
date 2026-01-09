package web

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with authentication information.
type Session struct {
	ID        string
	UserID    uuid.UUID
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore manages in-memory sessions with TTL and automatic cleanup.
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewSessionStore creates a new session store with the given TTL.
// A background goroutine is started to periodically clean up expired sessions.
func NewSessionStore(ttl time.Duration) *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}

	go store.cleanup()

	return store
}

// Save stores a session in the store.
func (s *SessionStore) Save(session *Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

// Get retrieves a session by ID.
// Returns an error if the session is not found or has expired.
func (s *SessionStore) Get(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return session, nil
}

// Delete removes a session from the store.
func (s *SessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// cleanup runs periodically to remove expired sessions.
// This prevents memory leaks from abandoned sessions.
func (s *SessionStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}
