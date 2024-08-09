# Literary Lions Forum

A web forum for Literary Lions allowing users to communicate, associate categories with posts, like/dislike posts & comments, and filter posts.
Test

## Before starting the program

- Instal Golang
- go.mod
```
go mod init lions
```

Step-by-Step Guide to Create Database and Tables with SQLite


Install SQLite:
```
sudo apt update
sudo apt install sqlite3
```

Create the Database and Tables:

Save the following schema in a file named schema.sql:
```
-- Create tables
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    category TEXT,
    content TEXT,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER,
    user_id INTEGER,
    content TEXT,
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS likes_dislikes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    post_id INTEGER,
    comment_id INTEGER,
    type TEXT,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(comment_id) REFERENCES comments(id)
);
```

Apply the Schema to Create the Database:
```
 sqlite3 forum.db < database/schema.sql
```
This command will create the forum.db database and apply the schema, creating all the tables defined.


Dockerize the Application

Create a Dockerfile:
In your project directory, create a file named Dockerfile with the following contents:

Dockerfile
```
# Use an official Golang runtime as a parent image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Command to run the executable
CMD ["./main"]
```

Build the Docker Image:
```
sudo docker build -t literary-lions .
```
Run the Docker Container:
```
sudo docker run -p 8080:8080 literary-lions
```


HOW TO USE SQLite

Open SQLite Command Line Interface:
```
sqlite3 user.db
```

Basic SQLite Commands:

View All Tables:
```
.tables
```
Describe Table Schema:
```
.schema tablename
```
Insert Data:
```
INSERT INTO users (email, username, password) VALUES ('example@example.com', 'exampleuser', 'examplepass');
```
Query Data:
```
SELECT * FROM user;
```
Exit SQLite CLI:
```
.exit
```

HOW TO USE Docker

View Running Containers (split terminal window):
```
sudo docker ps
```
View All Containers (Including Stopped):
```
sudo docker ps -a
```
Stop a Container:
```
sudo docker stop <container_id>
```
Remove a Container:
```
sudo docker rm <container_id>
```
View Docker Logs:
```
sudo docker logs <container_id>
```

- go.sum
```
go mod tidy
```

for "golang.org/x/crypto/bcrypt" package
```
go get golang.org/x/crypto/bcrypt

```

for "github.com/gorilla/sessions" package
```
go get github.com/gorilla/sessions

```


## Using SQLite Browser

Install SQLite Browser:
```
sudo add-apt-repository -y ppa:linuxgndu/sqlitebrowser
sudo apt update
sudo apt install sqlitebrowser
```
Open SQLite Browser:
```
sqlitebrowser
```


## Directory tree

```
lions/
│
├── database/
│     ├── database.go
│     └── schema.sql
├── handle/
│     └── handlers.go
├── like/
│     └── like.go
├── password/
│     └── password.go
├── post/
│     └── post.go
├── static/
│     ├── css/
│     │    └── 
│     └── html/
│          ├── book_specific.html
│          ├── general.html
│          ├── genres.go
│          ├── login.html
│          ├── mainpage.html
│          ├── post.html
│          └── register.html
│              
├── Dockerfile
├── forum.db
├── go.sum
├── go.sum 
├── lions_projekti.txt
├── main.go
└── README.md
└── user.db
```               



## Coders

Laura Levistö - Jonathan Dahl - 9/24
