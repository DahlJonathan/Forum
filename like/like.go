// like.go
package like

import (
	"net/http"
	"strconv"
	"github.com/gorilla/sessions"
	"lions/database"
	"log"
)

var (
	// Replace with your own secret key for cookie encryption
	key   = []byte("super-secret-key")
	// Create a new session store with the provided secret key
	store = sessions.NewCookieStore(key)
)

// PostLikeCounts holds the counts of likes and dislikes for a post
type PostLikeCounts struct {
	PostID    int
	Likes     int
	Dislikes  int
}

// LikePostHandler handles the liking and disliking of posts.
// It processes POST requests to update the like/dislike status of a post for the current user.
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	// Handle only POST requests
	if r.Method == http.MethodPost {
		// Retrieve the session for the current request
		session, _ := store.Get(r, "session")
		// Get the user ID from the session values
		userID, _ := session.Values["userID"].(int)

		// Retrieve the post ID and like status from the form data
		postID, _ := strconv.Atoi(r.FormValue("post_id"))
		isLike := r.FormValue("is_like") == "true"

		// Insert or update the like/dislike status in the database
		_, err := database.DB.Exec(`
			INSERT INTO PostLikes (UserID, PostID, CommentID, IsLike) 
			VALUES (?, ?, NULL, ?) 
			ON CONFLICT(UserID, PostID, CommentID) 
			DO UPDATE SET IsLike = excluded.IsLike`, userID, postID, isLike)
		if err != nil {
		log.Printf("Error updating like/dislike: %v", err) // Log detailed error
		http.Error(w, "Could not update like: "+err.Error(), http.StatusInternalServerError)
		return
	}

		// Redirect the user to the posts page after updating
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	}
}

// GetPostLikeCounts retrieves the total number of likes and dislikes for each post
func GetPostLikeCounts() ([]PostLikeCounts, error) {
	rows, err := database.DB.Query(`
		SELECT 
			PostID, 
			SUM(CASE WHEN IsLike = TRUE THEN 1 ELSE 0 END) AS Likes, 
			SUM(CASE WHEN IsLike = FALSE THEN 1 ELSE 0 END) AS Dislikes 
		FROM 
			PostLikes 
		GROUP BY 
			PostID
	`)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []PostLikeCounts
	for rows.Next() {
		var pc PostLikeCounts
		if err := rows.Scan(&pc.PostID, &pc.Likes, &pc.Dislikes); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		results = append(results, pc)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return results, nil
}