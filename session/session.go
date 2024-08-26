// session.go
package session

import (
	"context"
	"net/http"
	"sync"
)

// sessionStore holds all active sessions in memory.
// It uses a read-write mutex to ensure safe concurrent access.
var sessionStore = struct {
	sync.RWMutex
	m map[string]SessionData
}{m: make(map[string]SessionData)}

// contextKey is a custom type for context keys to avoid key collisions.
type contextKey string

const (
	// Username is the key used to store and retrieve the username from the context.
	Username = contextKey("Username")
	// Authenticated is the key used to store and retrieve the authentication status from the context.
	Authenticated = contextKey("Authenticated")
)

// SessionData holds information about a user's session.
type SessionData struct {
	Username      string // Username of the user
	Authenticated bool   // Whether the user is authenticated
	UserID        int    // User ID associated with the session
}

// GetSession retrieves session data based on the session ID.
// Returns the session data and a boolean indicating if the session exists.
func GetSession(sessionID string) (SessionData, bool) {
	sessionStore.RLock() // Acquire a read lock
	defer sessionStore.RUnlock() // Ensure the lock is released when the function exits
	data, exists := sessionStore.m[sessionID]
	return data, exists
}

// SetSession stores session data associated with a session ID.
// It updates or creates a new session entry in the session store.
func SetSession(sessionID string, data SessionData) {
	sessionStore.Lock() // Acquire a write lock
	defer sessionStore.Unlock() // Ensure the lock is released when the function exits
	sessionStore.m[sessionID] = data
}

// SessionMiddleware is an HTTP middleware that handles session management.
// It retrieves session data from the session store and adds it to the request context.
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authenticated bool
		var sessionData SessionData

		// Try to retrieve the session cookie from the request
		sessionID, err := r.Cookie("session_id")
		if err == nil {
			// If the session cookie is present, retrieve the session data
			sessionData, authenticated = GetSession(sessionID.Value)
		} else {
			// If there is an error retrieving the cookie, consider the user as not authenticated
			authenticated = false
		}

		// Add the session data and authentication status to the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, Username, sessionData.Username)
		ctx = context.WithValue(ctx, Authenticated, authenticated)
		r = r.WithContext(ctx)

		// Pass control to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
