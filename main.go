// main.go
package main

import (
	"lions/database"
	"lions/handle"
	"lions/like"
	"lions/post"
	"lions/session"
	"log"
	"net/http"
)

func main() {
	// Initialize the database connection.
	// This sets up the connection to your database and ensures it's ready to use.
	database.Init()

	// Serve static files
	// This handles requests for static files (like CSS, JavaScript, or images) by serving them from the "static" directory.
	// The StripPrefix removes the "/static/" prefix from the URL path when accessing static files.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Apply session middleware to all routes
	// This applies the session middleware to various routes to manage user sessions.
	// The session middleware handles authentication and session management.
	http.Handle("/", session.SessionMiddleware(http.HandlerFunc(handle.MainPageHandler)))
	http.Handle("/register", session.SessionMiddleware(http.HandlerFunc(handle.RegisterHandler)))
	http.Handle("/login", session.SessionMiddleware(http.HandlerFunc(handle.LoginHandler)))
	http.Handle("/logout", session.SessionMiddleware(http.HandlerFunc(handle.LogoutHandler)))
	http.Handle("/profile", session.SessionMiddleware(http.HandlerFunc(handle.ProfileHandler)))

	http.Handle("/post/create", session.SessionMiddleware(http.HandlerFunc(post.CreatePost)))
	http.Handle("/post/view", session.SessionMiddleware(http.HandlerFunc(post.ViewPost)))
	http.Handle("/post", session.SessionMiddleware(http.HandlerFunc(post.ListPosts)))
	http.Handle("/post/reply", session.SessionMiddleware(http.HandlerFunc(post.AddReply)))
	http.Handle("/like", session.SessionMiddleware(http.HandlerFunc(like.LikeHandler)))

	// Define routes that do not use session middleware
	// These routes handle actions that do not require session management, like password reset or email confirmation.
	http.HandleFunc("/confirm", handle.ConfirmEmailHandler)
	http.HandleFunc("/password-reset-request", handle.PasswordResetRequestHandler)
	http.HandleFunc("/reset-password", handle.ResetPasswordHandler)
	http.HandleFunc("/delete-account", handle.DeleteAccountHandler)

	http.HandleFunc("/filter", post.FilterPostHandler)

	// Start the HTTP server
	// This listens for incoming HTTP requests on port 8080 and serves them using the routes defined above.
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
