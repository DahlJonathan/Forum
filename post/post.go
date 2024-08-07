package post

import (
	"database/sql"
	"lions/database"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"
)

type Post struct {
	ID       int
	Title    string
	Content  string
	Username string
	Category string
}

type Reply struct {
	Content string
}

type PostViewData struct {
	Post          Post
	Replies       []Reply
	Authenticated bool
}

var (
	// Replace with your own secret key
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func ViewPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")

	var post Post
	err := database.DB.QueryRow(`SELECT PostID, Title, Content FROM Post WHERE PostID = ?`, postID).Scan(&post.ID, &post.Title, &post.Content)
	if err != nil {
		http.Error(w, "Could not retrieve post", http.StatusInternalServerError)
		return
	}

	rows, err := database.DB.Query(`SELECT Content FROM Comment WHERE PostID = ?`, postID)
	if err != nil {
		http.Error(w, "Could not retrieve replies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var reply Reply
		if err := rows.Scan(&reply.Content); err != nil {
			http.Error(w, "Could not scan reply", http.StatusInternalServerError)
			return
		}
		replies = append(replies, reply)
	}

	session, _ := store.Get(r, "session")
	authenticated := session.Values["username"] != nil

	data := PostViewData{
		Post:          post,
		Replies:       replies,
		Authenticated: authenticated,
	}

	tmpl, err := template.ParseFiles("static/html/view_post.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get form values
		title := r.FormValue("title")
		content := r.FormValue("content")
		category := r.FormValue("category")

		// Get session
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"].(string)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Retrieve user ID
		var userID int
		err := database.DB.QueryRow(`SELECT UserID FROM User WHERE Username = ?`, username).Scan(&userID)
		if err != nil {
			http.Error(w, "Could not find user", http.StatusInternalServerError)
			return
		}

		// Retrieve or create category ID
		var categoryID int
		err = database.DB.QueryRow(`SELECT CategoryID FROM Category WHERE CategoryName = ?`, category).Scan(&categoryID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Could not find category", http.StatusInternalServerError)
			return
		}

		if err == sql.ErrNoRows {
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

		// Insert post
		_, err = database.DB.Exec(`INSERT INTO Post (Title, Content, UserID, CategoryID) VALUES (?, ?, ?, ?)`,
			title, content, userID, categoryID)
		if err != nil {
			http.Error(w, "Could not create post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func ListPosts(w http.ResponseWriter, r *http.Request) {
	// Retrieve posts from database
	rows, err := database.DB.Query(`
		SELECT Post.PostID, Post.Title, Post.Content, User.Username, Category.CategoryName
		FROM Post
		JOIN User ON Post.UserID = User.UserID
		LEFT JOIN Category ON Post.CategoryID = Category.CategoryID
		ORDER BY Post.PostID DESC
	`)
	if err != nil {
		http.Error(w, "Could not retrieve posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Scan posts into slice
	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Username, &post.Category); err != nil {
			http.Error(w, "Could not scan post", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Retrieve session data
	session, _ := store.Get(r, "session")
	authenticated := session.Values["username"] != nil
	username, _ := session.Values["username"].(string)

	// Prepare data for template
	data := struct {
		Posts         []Post
		Authenticated bool
		Username      string
	}{
		Posts:         posts,
		Authenticated: authenticated,
		Username:      username,
	}

	// Render template
	tmpl, err := template.ParseFiles("static/html/post.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
