# Forum

Users can post & comment, view posts & comments, like or dislike posts & comments, and filter posts.


## Before starting the program

- Instal Golang

- Install SQLite:
```
sudo apt update
sudo apt install sqlite3
```

HOW TO USE SQLite

Open SQLite Command Line Interface (split the terminal window):
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
Query Data:
```
SELECT * FROM user;
```
Exit SQLite CLI:
```
.exit
```

- Dockerize the Application

Build the Docker Image:
```
sudo docker build -t literary-lions .
```
Run the Docker Container:
```
sudo docker run -p 8080:8080 literary-lions
```

HOW TO USE Docker

View Running Containers (split the terminal window):
```
sudo docker ps
```
View Docker Logs:
```
sudo docker logs <container_id>
```


## Using SQLite Browser (optional)

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


## Starting the program

Type in terminal window 
```
go run .
```

A server will start at localhost:8080.
If you type localhost:8080 in your web browser our Literary Lions forum will open.


## Register

You can register by entering 

- Password
- Email
- Username
  
When you have registered you will get a confirmation email from literary.lions.verf@gmail.com.

Here are 2 email addresses you can test the forum with

Email: lionsreviewer1@gmail.com  
Password: Reviewer1-

Email: lionsreviewer2@gmail.com  
Password: Reviewer2test


## Login

You can now log in using your

- Email
- Password

If you don't remember your password you can click on the reset password link below login form.
Then you can enter your email and literary.lions.verf@gmail.com will send you an email with a reset password link.
Click the link and you will get to a page where you can input your new password.


## Forum

If you are not logged in you can

- filter posts by category and replies
- read posts and comments 

If you are logged in you can

- create posts
- reply on posts
- edit your post
- edit your reply
- tag users in posts
- like/dislike posts
- like/dislike comments
- filter posts by category/replies/likes/dislikes/time
- delete your post

## My Page / need to be logged in

Here you can see your

- username
- email
- number of posts you have liked
- number of posts you have disliked
- Number of comments you have replied
- Number of posts you have posted

And delete your account.

## My Posts / need to be logged in

Here you can see all your posts and likes, and if you click on them the post opens.

