// password.go
package password

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	DB *sql.DB // Global database connection object
)

// HashPassword hashes a plaintext password using bcrypt.
// It returns the hashed password and an error if the hashing fails.
func HashPassword(password string) (string, error) {
	// Generate a hashed password with the default cost factor
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a hashed password with a plaintext password.
// It returns an error if the passwords do not match.
func CheckPassword(hashedPassword, password string) error {
	// Compare the hashed password with the plaintext password
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateResetToken generates a random reset token and stores it in the database.
// It returns the generated token and an error if token creation or storage fails.
func GenerateResetToken(userID int) (string, error) {
	// Create a 32-byte random token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	tokenStr := fmt.Sprintf("%x", token) // Convert token to hexadecimal string

	// Set the token expiration time to 1 hour from now
	expiration := time.Now().Add(1 * time.Hour)

	// Insert the token into the database
	_, err = DB.Exec("INSERT INTO PasswordResetToken (UserID, Token, Expiration) VALUES (?, ?, ?)", userID, tokenStr, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	return tokenStr, nil
}

// ResetPassword resets a user's password using a provided reset token.
// It returns an error if the token is invalid, expired, or if updating the password fails.
func ResetPassword(tokenStr, newPassword string) error {
	// Retrieve the user ID and expiration time for the provided token
	var userID int
	var expiration time.Time
	err := DB.QueryRow("SELECT UserID, Expiration FROM PasswordResetToken WHERE Token = ?", tokenStr).Scan(&userID, &expiration)
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	// Check if the token has expired
	if time.Now().After(expiration) {
		return fmt.Errorf("token has expired")
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the user's password in the database
	_, err = DB.Exec("UPDATE User SET Password = ? WHERE UserID = ?", hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete the used reset token from the database
	_, err = DB.Exec("DELETE FROM PasswordResetToken WHERE Token = ?", tokenStr)
	if err != nil {
		return fmt.Errorf("failed to delete reset token: %w", err)
	}

	return nil
}

// generateToken generates a random token and returns it as a hexadecimal string.
// It returns an error if the token generation fails.
func generateToken() (string, error) {
	// Create a 32-byte random token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	// Convert the token to a hexadecimal string
	return hex.EncodeToString(token), nil
}

// SendResetEmail sends a password reset email with the reset token to the user.
// It returns an error if sending the email fails.
func SendResetEmail(email, token string) error {
	from := "no-reply@yourdomain.com"       // Sender email address
	password := "your-email-password"        // Email account password
	to := email                             // Recipient email address
	subject := "Password Reset Request"      // Email subject
	body := fmt.Sprintf("Click the following link to reset your password: http://yourdomain.com/reset-password?token=%s", token)

	// Prepare the email message
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	smtpServer := "smtp.yourdomain.com:587" // SMTP server address
	auth := smtp.PlainAuth("", from, password, "smtp.yourdomain.com") // SMTP authentication

	// Send the email
	err := smtp.SendMail(smtpServer, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// RequestPasswordReset handles a request for a password reset.
// It generates a reset token, stores it in the database, and sends a reset email to the user.
// It returns an error if any step of this process fails.
func RequestPasswordReset(email string) error {
	// Check if the provided email address exists in the database
	var userID int
	err := DB.QueryRow("SELECT UserID FROM User WHERE Email = ?", email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Generate a reset token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store the token and associated user ID in the database
	_, err = DB.Exec("INSERT INTO PasswordResetToken (UserID, Token, Expiration) VALUES (?, ?, ?)", userID, token, time.Now().Add(1*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Send a password reset email to the user
	err = SendResetEmail(email, token)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}
