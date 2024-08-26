// database.go
package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB       *sql.DB // Global database connection handle
	Sessions = make(map[string]string) // Session ID to User ID mapping
)

// Init initializes the database connection and sets up the schema
func Init() {
	var err error
	// Open a connection to the SQLite database file "user.db"
	DB, err = sql.Open("sqlite3", "./user.db")
	if err != nil {
		// Log and terminate the program if the database connection fails
		log.Fatal(err)
	}

	// Open the schema.sql file which contains SQL statements to create the schema
	file, err := os.Open("database/schema.sql")
	if err != nil {
		// Log and terminate if opening the file fails
		log.Fatalf("Failed to open schema.sql: %v", err)
	}
	defer file.Close() // Ensure the file is closed after reading

	// Read the content of schema.sql file
	var schema string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		schema += scanner.Text() + "\n"
	}

	// Check for errors during file reading
	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}

	// Execute the SQL statements to set up the database schema
	_, err = DB.Exec(schema)
	if err != nil {
		// Log and terminate if executing the schema fails
		log.Fatalf("Failed to execute schema: %v", err)
	}

	// Log a message indicating the schema setup was successful or already exists
	log.Println("Database schema and indexes created or already exist.")
}

// InsertUser inserts a new user into the database
func InsertUser(username, email, password string) error {
	// Check if the username already exists in the database
	var usernameExists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Username = ?)", username).Scan(&usernameExists)
	if err != nil {
		return fmt.Errorf("failed to check for existing username: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("username already exists")
	}

	// Check if the email already exists in the database
	var emailExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM User WHERE Email = ?)", email).Scan(&emailExists)
	if err != nil {
		return fmt.Errorf("failed to check for existing email: %w", err)
	}
	if emailExists {
		return fmt.Errorf("email already exists")
	}

	// Insert the new user into the User table
	_, err = DB.Exec("INSERT INTO User (Username, Email, Password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetUserStats retrieves the number of posts, comments, likes, and dislikes for a given user
func GetUserStats(userID int) (numPosts, numComments, likes, dislikes int, err error) {
    // Count the number of posts for the specified user
    err = DB.QueryRow(`SELECT COUNT(*) FROM Post WHERE UserID = ?`, userID).Scan(&numPosts)
    if err != nil {
        return 0, 0, 0, 0, err
    }

    // Count the number of comments for the specified user
    err = DB.QueryRow(`SELECT COUNT(*) FROM Comment WHERE UserID = ?`, userID).Scan(&numComments)
    if err != nil {
        return 0, 0, 0, 0, err
    }

    // Count the number of likes given by the user
    err = DB.QueryRow(`
        SELECT COUNT(*) 
        FROM PostLikes 
        WHERE UserID = ? AND IsLike = 1
    `, userID).Scan(&likes)
    if err != nil {
        return 0, 0, 0, 0, err
    }

    // Count the number of dislikes given by the user
    err = DB.QueryRow(`
        SELECT COUNT(*) 
        FROM PostLikes 
        WHERE UserID = ? AND IsLike = 0
    `, userID).Scan(&dislikes)
    if err != nil {
        return 0, 0, 0, 0, err
    }

    return numPosts, numComments, likes, dislikes, nil
}
