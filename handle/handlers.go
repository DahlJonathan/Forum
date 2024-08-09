package handle

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"lions/database"
	"lions/password"
	"lions/email"
	"lions/post"
	"log"
	"net/http"
	"time"
    "crypto/rand"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

var (
	DB *sql.DB
)

type contextKey string

const (
    UsernameKey      = contextKey("Username")
    AuthenticatedKey = contextKey("Authenticated")
)

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")

		username, usernameOk := session.Values["username"].(string)
		authenticated, authOk := session.Values["authenticated"].(bool)

		if !usernameOk {
			username = ""
		}
		if !authOk {
			authenticated = false
		}

		ctx := context.WithValue(r.Context(), "Username", username)
		ctx = context.WithValue(ctx, "Authenticated", authenticated)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	username, _ := r.Context().Value("Username").(string)
	authenticated, _ := r.Context().Value("Authenticated").(bool)

	data := map[string]interface{}{
		"Username":      username,
		"Authenticated": authenticated,
	}

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

func ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	emailAddr := r.URL.Query().Get("email")
	if emailAddr == "" {
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	// Update the user's status to confirmed in the database
	_, err := database.DB.Exec(`UPDATE User SET Confirmed = 1 WHERE Email = ?`, emailAddr)
	if err != nil {
		log.Println("Error confirming email:", err)
		http.Error(w, "Failed to confirm email", http.StatusInternalServerError)
		return
	}

	log.Printf("Email confirmed: %s", emailAddr)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("static/html/register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		name := r.FormValue("username")
		emailAddr := r.FormValue("email")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec(`INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)`, name, emailAddr, hashedPassword)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send welcome email
		subject := "Welcome to Our Service"
		body := fmt.Sprintf("Hello %s,\n\nWelcome to our literary-lions task!\n\nThank you for registering.\n\nBest regards\nLaura and Jonathan", name)
		err = email.SendEmail(emailAddr, subject, body)
		if err != nil {
			log.Println("Error sending welcome email:", err)
			http.Error(w, "Failed to send welcome email", http.StatusInternalServerError)
			return
		}

		log.Printf("User registered: username=%s, email=%s", name, emailAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post.CreatePost(w, r)
	case http.MethodGet:
		post.ListPosts(w, r)
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		log.Printf("Login attempt with email: %s", email)

		var dbPassword, username string
		err := database.DB.QueryRow(`SELECT Password, Username FROM User WHERE Email = ?`, email).Scan(&dbPassword, &username)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Email not found: %s", email)
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
				return
			}
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("Fetched hashed password for email: %s", email)

		// Compare the hashed password with the plain text password
		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
		if err != nil {
			log.Printf("Invalid password for email: %s", email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["authenticated"] = true
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, "/mainpage", http.StatusSeeOther)
		return
	} else {
		tmpl, err := template.ParseFiles("static/html/login.html")
		if err != nil {
			log.Println("Template parsing error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["authenticated"] = false
	session.Values["username"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, _ := session.Values["username"].(string)

	var email string
	var userID int

	// Fetch user details
	err := database.DB.QueryRow(`
        SELECT UserID, Email 
        FROM User 
        WHERE Username = ?`, username).Scan(&userID, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get number of posts and comments
	numPosts, numComments, err := database.GetUserStats(userID)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username":    username,
		"Email":       email,
		"NumPosts":    numPosts,
		"NumComments": numComments,
	}

	tmpl, err := template.ParseFiles("static/html/profile.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


func GenerateResetToken(userID int) (string, error) {
    token := make([]byte, 32) // Generate a 32-byte token
    _, err := rand.Read(token)
    if err != nil {
        return "", err
    }
    tokenStr := fmt.Sprintf("%x", token)

    expiration := time.Now().Add(1 * time.Hour) // Token valid for 1 hour

    _, err = DB.Exec("INSERT INTO PasswordResetToken (UserID, Token, Expiration) VALUES (?, ?, ?)", userID, tokenStr, expiration)
    if err != nil {
        return "", fmt.Errorf("failed to store reset token: %w", err)
    }

    return tokenStr, nil
}


func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show the reset password form
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token not provided", http.StatusBadRequest)
			return
		}

		// Render the form with the token
		data := map[string]string{"Token": token}
		tmpl, err := template.ParseFiles("static/html/password_reset_form.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Handle the form submission
		token := r.FormValue("token")
		newPassword := r.FormValue("password")

		if token == "" || newPassword == "" {
			http.Error(w, "Token or password not provided", http.StatusBadRequest)
			return
		}

		// Fetch the userID associated with the token
		var userID int
		err := DB.QueryRow("SELECT UserID FROM PasswordResetTokens WHERE Token = ?", token).Scan(&userID)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusBadRequest)
			return
		}

		// Hash the new password
		hash, err := password.HashPassword(newPassword)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Update the password
		_, err = DB.Exec("UPDATE User SET Password = ? WHERE UserID = ?", hash, userID)
		if err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		// Delete the token
		_, err = DB.Exec("DELETE FROM PasswordResetTokens WHERE Token = ?", token)
		if err != nil {
			http.Error(w, "Failed to delete token", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}