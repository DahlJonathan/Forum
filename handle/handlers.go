// handlers.go
package handle

import (
	"database/sql"
	"fmt"
	"html/template"
	"lions/database"
	"lions/email"

	//"lions/post"
	"lions/session"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// PageData is used to pass data to templates
type PageData struct {
	Email string
	Error string
}

///////////////SessionMiddleware ////////////////////

// MainPageHandler serves the main page of the application
func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve session data from the context
	username, _ := r.Context().Value(session.Username).(string)
	authenticated, _ := r.Context().Value(session.Authenticated).(bool)

	// Prepare data to be passed to the template
	data := map[string]interface{}{
		"Username":      username,
		"Authenticated": authenticated,
	}

	// Parse and execute the main page template
	tmpl, err := template.ParseFiles("static/html/mainpage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Render the registration form
		tmpl, err := template.ParseFiles("static/html/register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		// Retrieve form values
		name := r.FormValue("username")
		emailAddr := r.FormValue("email")
		password := r.FormValue("password")

		// Hash the user's password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the new user into the database
		_, err = database.DB.Exec(`INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)`, name, emailAddr, hashedPassword)
		if err != nil {
			var errorMessage string
			if sqliteErr, ok := err.(sqlite3.Error); ok {
				if sqliteErr.Code == sqlite3.ErrConstraint {
					if strings.Contains(sqliteErr.Error(), "User.Username") {
						errorMessage = "The username is already taken."
					} else if strings.Contains(sqliteErr.Error(), "User.Email") {
						errorMessage = "The email is already registered."
					}
				}
			}
			renderRegister(w, errorMessage)
			return
		}

		// Send a welcome email to the user
		subject := "Welcome to Literary Lions Forum!"
		body := fmt.Sprintf("Hello %s,\n\nThank you for registering at Literary Lions Forum!\n\nBest regards,\nThe Literary Lions Team", name)
		err = email.SendEmail(emailAddr, subject, body)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

		log.Printf("User registered: username=%s, email=%s", name, emailAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Retrieve form values
		email := r.FormValue("email")
		password := r.FormValue("password")

		log.Printf("Login attempt with email: %s", email)

		var dbPassword, username string
		// Fetch the hashed password and username from the database
		err := database.DB.QueryRow(`SELECT Password, Username FROM User WHERE Email = ?`, email).Scan(&dbPassword, &username)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Email not found: %s", email)
				renderLogin(w, "Invalid email or password")
				return
			}
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("Fetched hashed password for email: %s", email)

		// Compare the provided password with the hashed password
		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
		if err != nil {
			log.Printf("Invalid password for email: %s", email)
			renderLogin(w, "Invalid email or password")
			return
		}

		// Create a new session for the authenticated user
		sessionID := uuid.New().String()
		session.SetSession(sessionID, session.SessionData{
			Username:      username,
			Authenticated: true,
		})

		// Set the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true, // Important for security
			Secure:   true, // Use true in production with HTTPS
		})
		http.Redirect(w, r, "/mainpage", http.StatusSeeOther)
		return
	} else {
		renderLogin(w, "")
	}
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Invalidate the session
		session.SetSession(cookie.Value, session.SessionData{})

		// Remove the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "session_id",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ProfileHandler serves the user's profile page
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve session data from the cookie
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionData, ok := session.GetSession(sessionID.Value)
	if !ok || !sessionData.Authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Fetch user data from the database
	var userInfo struct {
		Username    string
		Email       string
		NumPosts    int
		NumComments int
		NumLikes    int
		NumDislikes int
	}

	// Get user ID based on username
	var userID int
	err = database.DB.QueryRow(`SELECT UserID FROM User WHERE Username = ?`, sessionData.Username).Scan(&userID)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch basic user info
	err = database.DB.QueryRow(`SELECT Username, Email FROM User WHERE Username = ?`, sessionData.Username).
		Scan(&userInfo.Username, &userInfo.Email)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch counts of posts, comments, likes, and dislikes
	err = database.DB.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM Post WHERE UserID = ?) AS NumPosts,
			(SELECT COUNT(*) FROM Comment WHERE UserID = ?) AS NumComments,
			(SELECT COUNT(*) FROM PostLikes WHERE UserID = ? AND IsLike = 1) AS NumLikes,
			(SELECT COUNT(*) FROM PostLikes WHERE UserID = ? AND IsLike = 0) AS NumDislikes
	`, userID, userID, userID, userID).Scan(&userInfo.NumPosts, &userInfo.NumComments, &userInfo.NumLikes, &userInfo.NumDislikes)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render the profile page
	tmpl, err := template.ParseFiles("static/html/profile.html")
	if err != nil {
		log.Println("Template parsing error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username":    userInfo.Username,
		"Email":       userInfo.Email,
		"NumPosts":    userInfo.NumPosts,
		"NumComments": userInfo.NumComments,
		"NumLikes":    userInfo.NumLikes,
		"NumDislikes": userInfo.NumDislikes,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

///////////////SessionMiddleware END////////////////////

// renderRegister renders the registration page with an error message
func renderRegister(w http.ResponseWriter, errorMessage string) {
	tmpl, err := template.ParseFiles("static/html/register.html")
	if err != nil {
		log.Println("Template parsing error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}

	tmpl.Execute(w, data)
}

// renderLogin renders the login page with an error message
func renderLogin(w http.ResponseWriter, errorMessage string) {
	tmpl, err := template.ParseFiles("static/html/login.html")
	if err != nil {
		log.Println("Template parsing error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ErrorMessage": errorMessage,
	}

	tmpl.Execute(w, data)
}

///////////////Email ////////////////////

// ConfirmEmailHandler confirms the user's email address
func ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the email address from the query parameters
	emailAddr := r.URL.Query().Get("email")
	if emailAddr == "" {
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	// Update the user's email confirmation status in the database
	_, err := database.DB.Exec(`UPDATE User SET Confirmed = 1 WHERE Email = ?`, emailAddr)
	if err != nil {
		log.Println("Error confirming email:", err)
		http.Error(w, "Failed to confirm email", http.StatusInternalServerError)
		return
	}

	log.Printf("Email confirmed: %s", emailAddr)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

///////////////Password ////////////////////

// PasswordResetRequestHandler handles password reset requests
func PasswordResetRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		emailAddr := r.FormValue("email")

		log.Printf("Password reset request for email: %s", emailAddr)

		var username string
		// Verify that the email exists in the database
		err := database.DB.QueryRow(`SELECT Username FROM User WHERE Email = ?`, emailAddr).Scan(&username)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Email not found: %s", emailAddr)
				renderPasswordReset(w, emailAddr, "Email not found")
				return
			}
			log.Printf("Database error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Generate a password reset token and store it in the database
		token := uuid.New().String()
		expiry := time.Now().Add(1 * time.Hour)

		_, err = database.DB.Exec(`INSERT INTO PasswordReset (Email, Token, Expiry) VALUES (?, ?, ?)`, emailAddr, token, expiry)
		if err != nil {
			log.Printf("Failed to insert reset token for email %s: %v", emailAddr, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send password reset email with the token
		resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
		subject := "Password Reset Request"
		body := fmt.Sprintf("Hello %s,\n\nTo reset your password, click the following link: %s\n\nIf you did not request this, please ignore this email.\n\nBest regards,\nThe Literary Lions Team", username, resetURL)
		err = email.SendEmail(emailAddr, subject, body)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", emailAddr, err)
			renderPasswordReset(w, emailAddr, "Failed to send email")
			return
		}

		log.Printf("Password reset email sent to %s", emailAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	} else {
		renderPasswordReset(w, "", "")
	}
}

// renderPasswordReset renders the password reset request page
func renderPasswordReset(w http.ResponseWriter, email string, errorMsg string) {
	tmpl, err := template.ParseFiles("static/html/password-reset-request.html")
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]string{
		"Email": email,
		"Error": errorMsg,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ResetPasswordHandler handles password resets using a token
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		token := r.FormValue("token")
		newPassword := r.FormValue("password")

		var emailAddr string
		var expiry time.Time

		// Verify the token and fetch its associated email and expiry time
		err := database.DB.QueryRow(`SELECT Email, Expiry FROM PasswordReset WHERE Token = ?`, token).Scan(&emailAddr, &expiry)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid or expired token", http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if time.Now().After(expiry) {
			http.Error(w, "Token has expired", http.StatusBadRequest)
			return
		}

		// Hash the new password and update it in the database
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec(`UPDATE User SET Password = ? WHERE Email = ?`, hashedPassword, emailAddr)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Remove the reset token from the database
		_, err = database.DB.Exec(`DELETE FROM PasswordReset WHERE Token = ?`, token)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		// Render the password reset page with the token
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token is required", http.StatusBadRequest)
			return
		}

		tmpl, err := template.ParseFiles("static/html/reset-password.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Token": token,
		}

		tmpl.Execute(w, data)
	}
}

///////////////Delete Account ////////////////////

// DeleteAccountHandler handles account deletion
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the user's authentication
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionData, ok := session.GetSession(sessionID.Value)
	if !ok || !sessionData.Authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Delete the user's account from the database
	_, err = database.DB.Exec(`DELETE FROM User WHERE Username = ?`, sessionData.Username)
	if err != nil {
		log.Println("Error deleting user:", err)
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	// Invalidate the user's session
	session.SetSession(sessionID.Value, session.SessionData{})

	// Remove the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	// Redirect to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

/*/ PostsHandler handles displaying and creating posts
func PostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		post.ListPosts(w, r)  // List existing posts
	case http.MethodPost:
		post.CreatePost(w, r)  // Create a new post
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}*/
