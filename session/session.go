package session

import (
	"context"
	"net/http"
	"sync"
)

// sessionStore holds all active sessions in memory.
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
	// UserID is the key used to store and retrieve the user ID from the context.
	UserID = contextKey("UserID")
)

// SessionData holds information about a user's session.
type SessionData struct {
	Username      string // Username of the user
	Authenticated bool   // Whether the user is authenticated
	UserID        int    // User ID associated with the session
}

// GetSession retrieves session data based on the session ID.
func GetSession(sessionID string) (SessionData, bool) {
	sessionStore.RLock()
	defer sessionStore.RUnlock()
	data, exists := sessionStore.m[sessionID]
	return data, exists
}

// SetSession stores session data associated with a session ID.
func SetSession(sessionID string, data SessionData) {
	sessionStore.Lock()
	defer sessionStore.Unlock()
	sessionStore.m[sessionID] = data
}

// SessionMiddleware is an HTTP middleware that handles session management.
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
		ctx = context.WithValue(ctx, UserID, sessionData.UserID) // Add UserID to context
		r = r.WithContext(ctx)

		// Pass control to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
