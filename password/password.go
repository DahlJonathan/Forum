//password.go
package password

import (

    "golang.org/x/crypto/bcrypt"
    "fmt"
	"database/sql"
    "time"
	"net/smtp"
    "crypto/rand"
    "encoding/hex"
    //"crypto/sha256"
   // "encoding/base64"
)

/*/ Hashes passwords using SHA256 (consider using a better approach in production)
func HashPassword(password string) string {
    hash := sha256.Sum256([]byte(password))
    return base64.URLEncoding.EncodeToString(hash[:])
}

*/
var (
	DB *sql.DB
)


func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

/*
func InsertUser(username, email, password string) error {
	// Check if the email already exists
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Email = ?)", email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check for existing email: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	// Insert new user
	_, err = DB.Exec("INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}*/

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

func ResetPassword(tokenStr, newPassword string) error {
	// Validate token
	var userID int
	var expiration time.Time
	err := DB.QueryRow("SELECT UserID, Expiration FROM PasswordResetToken WHERE Token = ?", tokenStr).Scan(&userID, &expiration)
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	if time.Now().After(expiration) {
		return fmt.Errorf("token has expired")
	}

	// Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	_, err = DB.Exec("UPDATE User SET Password = ? WHERE UserID = ?", hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Optionally, delete the token after use
	_, err = DB.Exec("DELETE FROM PasswordResetToken WHERE Token = ?", tokenStr)
	if err != nil {
		return fmt.Errorf("failed to delete reset token: %w", err)
	}

	return nil
}


// Generate a random token
func generateToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// Send a password reset email
func SendResetEmail(email, token string) error {
	from := "no-reply@yourdomain.com"
	password := "your-email-password"
	to := email
	subject := "Password Reset Request"
	body := fmt.Sprintf("Click the following link to reset your password: http://yourdomain.com/reset-password?token=%s", token)
	
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	
	smtpServer := "smtp.yourdomain.com:587"
	auth := smtp.PlainAuth("", from, password, "smtp.yourdomain.com")

	err := smtp.SendMail(smtpServer, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// RequestPasswordReset handles the password reset request
func RequestPasswordReset(email string) error {
	// Check if the email exists
	var userID int
	err := DB.QueryRow("SELECT UserID FROM User WHERE Email = ?", email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Generate a token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store the token and associated userID
	_, err = DB.Exec("INSERT INTO PasswordResetTokens (UserID, Token) VALUES (?, ?)", userID, token)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Send the reset email
	err = SendResetEmail(email, token)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}