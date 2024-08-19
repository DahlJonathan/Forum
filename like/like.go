package like

import (
	"database/sql"
	"lions/database"
	"lions/session"
	"log"
	"net/http"
	"strconv"
)

// LikeHandler handles like/dislike requests for posts or comments
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated from the context
	ctx := r.Context()
	authenticated, ok := ctx.Value(session.Authenticated).(bool)
	if !ok || !authenticated {
		http.Error(w, "Unauthorized: User not logged in", http.StatusUnauthorized)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id") // Optional for comments
	isLike := r.FormValue("is_like") == "true"
	userID, ok := ctx.Value(session.UserID).(int) // Get userID from session context

	if !ok {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// No need to convert postID to integer if it's a UUID
	postID := postIDStr // Use postIDStr directly if it's a UUID

	var commentID *int = nil
	if commentIDStr != "" {
		commentIDVal, err := strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}
		commentID = &commentIDVal
	}

	// Call function to handle the like/dislike action
	err = handleLikeDislike(userID, postID, commentID, isLike)
	if err != nil {
		http.Error(w, "Error processing like/dislike", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post view
	http.Redirect(w, r, "/post/view?id="+postIDStr, http.StatusSeeOther)
}

func handleLikeDislike(userID int, postID string, commentID *int, isLike bool) error {
	// Check if the user has already liked/disliked this post/comment
	existingAction, err := getUserPostCommentAction(userID, postID, commentID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Determine the current action
	currentAction := "none"
	if existingAction != "" {
		if existingAction == "like" {
			currentAction = "like"
		} else if existingAction == "dislike" {
			currentAction = "dislike"
		}
	}

	if currentAction == "none" {
		// If no previous action, insert a new like/dislike
		err = insertLikeDislike(userID, postID, commentID, isLike)
	} else if (currentAction == "like" && !isLike) || (currentAction == "dislike" && isLike) {
		// If switching between like and dislike, update the record
		err = updateLikeDislike(userID, postID, commentID, isLike)
	} else {
		// If performing the same action again, do nothing
		return nil
	}
	log.Printf("Existing action: %s", existingAction)
	log.Printf("Performing action: %s", isLike)

	return err
}

func getUserPostCommentAction(userID int, postID string, commentID *int) (string, error) {
	var action bool
	query := "SELECT IsLike FROM PostLikes WHERE UserID = ? AND PostID = ? AND CommentID IS ?"
	err := database.DB.QueryRow(query, userID, postID, commentID).Scan(&action)
	if err != nil {
		return "", err
	}

	// Return "like" or "dislike" based on the IsLike value
	if action {
		return "like", nil
	} else {
		return "dislike", nil
	}
}

func insertLikeDislike(userID int, postID string, commentID *int, isLike bool) error {
	_, err := database.DB.Exec("INSERT INTO PostLikes (UserID, PostID, CommentID, IsLike) VALUES (?, ?, ?, ?)", userID, postID, commentID, isLike)
	return err
}

func updateLikeDislike(userID int, postID string, commentID *int, isLike bool) error {
	_, err := database.DB.Exec("UPDATE PostLikes SET IsLike = ? WHERE UserID = ? AND PostID = ? AND CommentID IS ?", isLike, userID, postID, commentID)
	return err
}
