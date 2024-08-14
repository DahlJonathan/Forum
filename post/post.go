// post.go
package post

import (
	"database/sql"
	"lions/database"
	"lions/session"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// Post represents a blog post with various details.
type Post struct {
	ID            string // Unique identifier for the post
	Title         string // Title of the post
	Content       string // Content of the post
	Username      string // Username of the post author
	Category      string // Category of the post
	Likes         int    // Number of likes on the post
	Dislikes      int    // Number of dislikes on the post
	CategoryID    int    // ID of the category
	UserID        int    // ID of the user who created the post
	RepliesCount  int    // Number of replies to the post
	Views         int    // Number of views of the post
	LastReplyDate string // Date of the last reply
	LastReplyUser string // Username of the user who made the last reply
	IsPopular     bool   // Whether the post is considered popular
}

// Pagination represents pagination data for listing posts.
type Pagination struct {
	CurrentPage int // Current page number
	TotalPages  int // Total number of pages
}

// PageData holds data for rendering post list or details pages.
type PageData struct {
	Authenticated bool   // Whether the user is authenticated
	Username      string // Username of the authenticated user
	Posts         []Post // List of posts to display
	Post          Post   // Single post for detailed view
	Replies       []Reply // List of replies to a post
	Pagination    Pagination // Pagination data
}

// Reply represents a reply to a post with user information.
type Reply struct {
	Content  string // Content of the reply
	Username string // Username of the reply author
}

// PostViewData holds data for rendering a single post with its replies.
type PostViewData struct {
	Post          Post   // Post details
	Replies       []Reply // Replies to the post
	Authenticated bool   // Whether the user is authenticated
	Username      string // Username of the authenticated user
}

// Define template functions for use in HTML templates.
func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

// ViewPost handles displaying a single post and its replies.
func ViewPost(w http.ResponseWriter, r *http.Request) {
	// Retrieve post ID from query parameters
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	var post Post
	// Retrieve post details from the database
	err := database.DB.QueryRow(`
        SELECT PostID, Title, Content, CategoryID, UserID
        FROM Post
        WHERE PostID = ?`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.CategoryID, &post.UserID)
	if err != nil {
		http.Error(w, "Could not retrieve post", http.StatusInternalServerError)
		return
	}

	var categoryName string
	// Retrieve category name for the post
	err = database.DB.QueryRow(`SELECT CategoryName FROM Category WHERE CategoryID = ?`, post.CategoryID).Scan(&categoryName)
	if err != nil {
		http.Error(w, "Could not retrieve category", http.StatusInternalServerError)
		return
	}
	post.Category = categoryName

	var username string
	// Retrieve username of the post author
	err = database.DB.QueryRow(`SELECT Username FROM User WHERE UserID = ?`, post.UserID).Scan(&username)
	if err != nil {
		http.Error(w, "Could not retrieve username", http.StatusInternalServerError)
		return
	}
	post.Username = username

	// Retrieve likes and dislikes for the post
	err = database.DB.QueryRow(`
        SELECT 
            (SELECT COUNT(*) FROM PostLikes WHERE PostID = ? AND IsLike = 1) AS Likes,
            (SELECT COUNT(*) FROM PostLikes WHERE PostID = ? AND IsLike = 0) AS Dislikes
        `, postID, postID).Scan(&post.Likes, &post.Dislikes)
	if err != nil {
		http.Error(w, "Could not retrieve likes/dislikes", http.StatusInternalServerError)
		return
	}

	// Fetch replies and associated user information
	rows, err := database.DB.Query(`
        SELECT c.Content, u.Username 
        FROM Comment c
        JOIN User u ON c.UserID = u.UserID
        WHERE c.PostID = ?`, postID)
	if err != nil {
		http.Error(w, "Could not retrieve comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var reply Reply
		if err := rows.Scan(&reply.Content, &reply.Username); err != nil {
			http.Error(w, "Could not scan comment", http.StatusInternalServerError)
			return
		}
		replies = append(replies, reply)
	}

	// Use session data from the request context
	authenticated := r.Context().Value(session.Authenticated).(bool)
	username = r.Context().Value(session.Username).(string)

	// Prepare data for rendering the template
	data := PostViewData{
		Post:          post,
		Replies:       replies,
		Authenticated: authenticated,
		Username:      username,
	}
	tmpl, err := template.New("view_post.html").Funcs(template.FuncMap{
		"add": add,
		"sub": sub,
	}).ParseFiles("static/html/view_post.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "view_post.html", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// CreatePost handles the creation of a new post.
func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		category := r.FormValue("category")

		// Validate input data
		if title == "" || content == "" || category == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		postID := uuid.New().String() // Generate a unique ID for the new post

		// Check if the user is authenticated
		authenticated := r.Context().Value(session.Authenticated).(bool)
		username := r.Context().Value(session.Username).(string)
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var userID int
		// Retrieve userID based on the username
		err := database.DB.QueryRow(`SELECT UserID FROM User WHERE Username = ?`, username).Scan(&userID)
		if err != nil {
			http.Error(w, "Could not find user", http.StatusInternalServerError)
			return
		}

		var categoryID int
		// Retrieve categoryID or create a new category if it does not exist
		err = database.DB.QueryRow(`SELECT CategoryID FROM Category WHERE CategoryName = ?`, category).Scan(&categoryID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Could not find category", http.StatusInternalServerError)
			return
		}

		if err == sql.ErrNoRows {
			// Insert a new category if it does not exist
			result, err := database.DB.Exec(`INSERT INTO Category (CategoryName) VALUES (?)`, category)
			if err != nil {
				http.Error(w, "Could not create category", http.StatusInternalServerError)
				return
			}
			categoryID64, err := result.LastInsertId()
			if err != nil {
				http.Error(w, "Could not retrieve category ID", http.StatusInternalServerError)
				return
			}
			categoryID = int(categoryID64)
		}

		// Insert the new post into the database
		_, err = database.DB.Exec(`INSERT INTO Post (PostID, Title, Content, UserID, CategoryID) VALUES (?, ?, ?, ?, ?)`,
			postID, title, content, userID, categoryID)
		if err != nil {
			http.Error(w, "Could not create post", http.StatusInternalServerError)
			return
		}

		// Redirect to the list of posts
		http.Redirect(w, r, "/post", http.StatusSeeOther)
	} else {
		// Render the create post template for GET requests
		tmpl, err := template.New("create_post.html").Funcs(template.FuncMap{
			"add": add,
			"sub": sub,
		}).ParseFiles("static/html/create_post.html")
		if err != nil {
			http.Error(w, "Template parsing error", http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "create_post.html", nil); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
			return
		}
	}
}

// ListPosts handles displaying a paginated list of posts.
func ListPosts(w http.ResponseWriter, r *http.Request) {
	// Extract page number from query parameters
	pageParam := r.URL.Query().Get("page")
	currentPage := 1
	if pageParam != "" {
		var err error
		currentPage, err = strconv.Atoi(pageParam)
		if err != nil || currentPage < 1 {
			currentPage = 1
		}
	}

	// Set up pagination parameters
	postsPerPage := 10
	offset := (currentPage - 1) * postsPerPage

	// Fetch total number of posts
	var totalPosts int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM Post").Scan(&totalPosts)
	if err != nil {
		http.Error(w, "Could not retrieve total post count", http.StatusInternalServerError)
		log.Printf("Error retrieving total post count: %v", err)
		return
	}

	// Fetch posts for the current page
	rows, err := database.DB.Query(`
        SELECT PostID, Title, Content, CategoryID, UserID
        FROM Post
        LIMIT ? OFFSET ?`, postsPerPage, offset)
	if err != nil {
		http.Error(w, "Could not retrieve posts", http.StatusInternalServerError)
		log.Printf("Error retrieving posts: %v", err)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CategoryID, &post.UserID)
		if err != nil {
			http.Error(w, "Could not scan post", http.StatusInternalServerError)
			log.Printf("Error scanning post: %v", err)
			return
		}

		// Fetch the category name for each post
		var categoryName string
		err = database.DB.QueryRow(`SELECT CategoryName FROM Category WHERE CategoryID = ?`, post.CategoryID).Scan(&categoryName)
		if err != nil {
			http.Error(w, "Could not retrieve category", http.StatusInternalServerError)
			log.Printf("Error retrieving category: %v", err)
			return
		}
		post.Category = categoryName

		// Fetch the username for each post
		var username string
		err = database.DB.QueryRow(`SELECT Username FROM User WHERE UserID = ?`, post.UserID).Scan(&username)
		if err != nil {
			http.Error(w, "Could not retrieve username", http.StatusInternalServerError)
			log.Printf("Error retrieving username: %v", err)
			return
		}
		post.Username = username

		posts = append(posts, post)
	}

	// Calculate the total number of pages
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage

	// Set up the pagination data
	pagination := Pagination{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	// Use session data from the request context
	authenticated := r.Context().Value(session.Authenticated).(bool)
	username := r.Context().Value(session.Username).(string)

	// Prepare data for rendering the template
	data := PageData{
		Posts:         posts,
		Pagination:    pagination,
		Authenticated: authenticated,
		Username:      username,
	}

	tmpl, err := template.New("post.html").Funcs(template.FuncMap{
		"add": add,
		"sub": sub,
	}).ParseFiles("static/html/post.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// AddReply handles adding a reply to a post.
func AddReply(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Check if the user is authenticated using session context
	authenticated := r.Context().Value(session.Authenticated).(bool)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve username from the session
	username := r.Context().Value(session.Username).(string)

	// Get the post ID and reply content from the form data
	postID := r.FormValue("postID")
	content := r.FormValue("content")
	log.Printf("Received postID: %s, content: %s", postID, content)

	// Validate the input data
	if postID == "" || content == "" {
		http.Error(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	// Retrieve the user ID from the database based on the username
	var userID int
	err := database.DB.QueryRow(`SELECT UserID FROM User WHERE Username = ?`, username).Scan(&userID)
	if err != nil {
		http.Error(w, "Could not retrieve user ID", http.StatusInternalServerError)
		return
	}

	// Insert the reply into the database
	_, err = database.DB.Exec(`INSERT INTO Comment (PostID, UserID, Content) VALUES (?, ?, ?)`,
		postID, userID, content)
	if err != nil {
		log.Printf("Error adding reply: %v", err) // Log the actual error
		http.Error(w, "Could not add reply: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the last reply date and user for the post
	_, err = database.DB.Exec(`
        UPDATE Post 
        SET LastReplyDate = ?, LastReplyUser = ?
        WHERE PostID = ?`, time.Now(), username, postID)
	if err != nil {
		http.Error(w, "Could not update post with last reply info", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post view page
	redirectURL := "/post/view?id=" + url.QueryEscape(postID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
