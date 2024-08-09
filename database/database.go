package database

import (
	"database/sql"
	"bufio"
	"log"
	"os"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB       *sql.DB
	Sessions = make(map[string]string) // Session ID -> User ID
)

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./user.db")
	if err != nil {
		log.Fatal(err)
	}

	// Open schema.sql file
	file, err := os.Open("database/schema.sql")
	if err != nil {
		log.Fatalf("Failed to open schema.sql: %v", err)
	}
	defer file.Close()

	// Read file content
	var schema string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		schema += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}

	// Execute schema creation
	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	log.Println("Database schema and indexes created or already exist.")
}

func InsertUser(username, email, password string) error {
	// Check if the username or email already exists
	var usernameExists, emailExists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Username = ?)", username).Scan(&usernameExists)
	if err != nil {
		return fmt.Errorf("failed to check for existing username: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("username already exists")
	}

	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Email = ?)", email).Scan(&emailExists)
	if err != nil {
		return fmt.Errorf("failed to check for existing email: %w", err)
	}
	if emailExists {
		return fmt.Errorf("email already exists")
	}

	// Insert new user
	_, err = DB.Exec("INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetUserStats retrieves the number of posts and comments for a user
func GetUserStats(userID int) (numPosts, numComments int, err error) {
	// Count posts
	err = DB.QueryRow(`SELECT COUNT(*) FROM Post WHERE UserID = ?`, userID).Scan(&numPosts)
	if err != nil {
		return 0, 0, err
	}

	// Count comments
	err = DB.QueryRow(`SELECT COUNT(*) FROM Comment WHERE UserID = ?`, userID).Scan(&numComments)
	if err != nil {
		return 0, 0, err
	}

	return numPosts, numComments, nil
}
