package main

import (
	"html/template"
	"log"
	"net/http"

	"lions/database"
	"lions/handle"
	"lions/post"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// Define the store variable
var (
	// Replace with your own secret key
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func main() {
	// Initialize the database connection
	database.Init()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Apply session middleware to all routes
	http.Handle("/", handle.SessionMiddleware(http.HandlerFunc(handle.MainPageHandler)))
	http.Handle("/register", handle.SessionMiddleware(http.HandlerFunc(handle.RegisterHandler)))
	http.Handle("/login", handle.SessionMiddleware(http.HandlerFunc(handle.LoginHandler)))
	http.Handle("/logout", handle.SessionMiddleware(http.HandlerFunc(handle.LogoutHandler)))
	http.Handle("/posts", handle.SessionMiddleware(http.HandlerFunc(post.ListPosts)))
	http.Handle("/post/create", handle.SessionMiddleware(http.HandlerFunc(post.CreatePost)))
	http.Handle("/mainpage", handle.SessionMiddleware(http.HandlerFunc(handle.MainPageHandler)))
	http.Handle("/post/view", handle.SessionMiddleware(http.HandlerFunc(post.ViewPost)))
	http.Handle("/profile", handle.SessionMiddleware(http.HandlerFunc(handle.ProfileHandler)))
	http.HandleFunc("/confirm", handle.ConfirmEmailHandler)

	// Serve static HTML files
	http.Handle("/post", handle.SessionMiddleware(http.HandlerFunc(servePage("static/html/post.html"))))
	//http.Handle("/general", handle.SessionMiddleware(http.HandlerFunc(servePage("static/html/general.html"))))
	//http.Handle("/genres", handle.SessionMiddleware(http.HandlerFunc(servePage("static/html/genres.html"))))
	//http.Handle("/book_specific", handle.SessionMiddleware(http.HandlerFunc(servePage("static/html/book_specific.html"))))

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// servePage returns a handler function that serves the static HTML file at the given path
func servePage(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		authenticated := session.Values["authenticated"] == true
		username, _ := session.Values["username"].(string)

		data := map[string]interface{}{
			"Username":      username,
			"Authenticated": authenticated,
		}

		tmpl, err := template.ParseFiles(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
