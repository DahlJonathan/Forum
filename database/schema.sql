-- schema.sql
-- Create tables

-- Table to store categories of posts
CREATE TABLE IF NOT EXISTS Category (
    CategoryID INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for each category
    CategoryName TEXT NOT NULL UNIQUE -- Name of the category, must be unique
);

-- Table to store user information
CREATE TABLE IF NOT EXISTS User (
    UserID INTEGER PRIMARY KEY, -- Unique identifier for each user
    Email TEXT UNIQUE NOT NULL, -- User's email address, must be unique
    Username TEXT UNIQUE NOT NULL, -- User's username, must be unique
    Password TEXT NOT NULL -- User's hashed password
);

CREATE TABLE IF NOT EXISTS Post (
    PostID TEXT PRIMARY KEY, -- UUID as TEXT
    Title TEXT,
    Content TEXT,
    UserID INTEGER,
    CategoryID INTEGER,
    LastReplyUser TEXT,
    LastReplyDate DATETIME,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (UserID) REFERENCES User(UserID) ON DELETE SET NULL,
    FOREIGN KEY (CategoryID) REFERENCES Category(CategoryID)
);

-- Table to store comments on posts
CREATE TABLE IF NOT EXISTS Comment (
    CommentID INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for each comment
    PostID INTEGER, -- ID of the post to which the comment belongs
    UserID INTEGER, -- ID of the user who made the comment
    Content TEXT NOT NULL, -- Content of the comment
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (PostID) REFERENCES Post(PostID) ON DELETE CASCADE, -- Foreign key to Post table
    FOREIGN KEY (UserID) REFERENCES User(UserID) ON DELETE SET NULL -- Foreign key to User table
);

CREATE TABLE IF NOT EXISTS PostLikes (
    ID INTEGER PRIMARY KEY AUTOINCREMENT, 
    UserID INTEGER NOT NULL, 
    PostID TEXT, -- UUID as TEXT
    CommentID INTEGER, 
    IsLike BOOLEAN NOT NULL, 
    FOREIGN KEY (UserID) REFERENCES User(UserID) ON DELETE CASCADE,
    FOREIGN KEY (PostID) REFERENCES Post(PostID) ON DELETE CASCADE,
    FOREIGN KEY (CommentID) REFERENCES Comment(CommentID) ON DELETE CASCADE,
    UNIQUE (UserID, PostID, CommentID)
);



-- Table to store password reset tokens
CREATE TABLE IF NOT EXISTS PasswordReset (
    Email TEXT NOT NULL, -- Email of the user requesting a password reset
    Token TEXT NOT NULL PRIMARY KEY, -- Unique token for the password reset request
    Expiry DATETIME NOT NULL -- Expiry date and time of the token
);

-- Table to store user sessions
CREATE TABLE IF NOT EXISTS Session (
    SessionID TEXT PRIMARY KEY, -- Unique identifier for each session
    UserID INTEGER NOT NULL, -- ID of the user associated with the session
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the session was created
    FOREIGN KEY (UserID) REFERENCES User(UserID) -- Foreign key to User table
);

-- Create indexes to improve query performance
CREATE INDEX IF NOT EXISTS idx_post_user ON Post(UserID); -- Index on UserID in Post table
CREATE INDEX IF NOT EXISTS idx_post_category ON Post(CategoryID); -- Index on CategoryID in Post table
CREATE INDEX IF NOT EXISTS idx_comment_post ON Comment(PostID); -- Index on PostID in Comment table
CREATE INDEX IF NOT EXISTS idx_comment_user ON Comment(UserID); -- Index on UserID in Comment table
CREATE INDEX IF NOT EXISTS idx_like_user ON PostLikes(UserID); -- Index on UserID in PostLikes table
CREATE INDEX IF NOT EXISTS idx_like_post ON PostLikes(PostID); -- Index on PostID in PostLikes table
CREATE INDEX IF NOT EXISTS idx_like_comment ON PostLikes(CommentID); -- Index on CommentID in PostLikes table






