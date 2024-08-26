-- data.sql

-- Insert initial data into Category table
INSERT INTO Category (CategoryName) 
VALUES ('New Category') 
ON CONFLICT(CategoryName) 
DO UPDATE SET CategoryName = excluded.CategoryName;

-- Explanation:
-- This statement attempts to insert a new row into the Category table with the name 'New Category'.
-- If a row with the same CategoryName already exists, it updates the existing row with the new CategoryName.
-- The ON CONFLICT clause handles cases where a conflict occurs due to a unique constraint on CategoryName.

-- Insert initial data into User table
INSERT INTO User (UserID, Email, Username, Password) 
VALUES (1, 'user@example.com', 'username', 'hashed_password') 
ON CONFLICT(UserID) 
DO UPDATE SET Email = excluded.Email, Username = excluded.Username, Password = excluded.Password;

-- Explanation:
-- This statement inserts a new user with a specific UserID, Email, Username, and Password into the User table.
-- If a user with the same UserID already exists, it updates the existing user's Email, Username, and Password.
-- The ON CONFLICT clause handles conflicts based on the UserID, which must be unique.

-- Insert initial data into User table with email conflict handling
INSERT INTO User (Email, Username, Password) 
VALUES ('user@example.com', 'username', 'hashed_password') 
ON CONFLICT(Email) 
DO UPDATE SET Username = excluded.Username, Password = excluded.Password;

-- Explanation:
-- This statement attempts to insert a new user based on the Email address.
-- If a user with the same Email already exists, it updates the Username and Password for that user.
-- The ON CONFLICT clause handles conflicts based on the Email column, which must be unique.

-- Insert initial data into PostLikes table
INSERT INTO PostLikes (UserID, PostID, CommentID, IsLike) 
VALUES (?, ?, ?, ?) 
ON CONFLICT(UserID, PostID, CommentID) 
DO UPDATE SET IsLike = excluded.IsLike;

-- Explanation:
-- This statement inserts a like for a post into the PostLikes table.
-- If a like already exists for the same UserID, PostID, and CommentID, it updates the IsLike value.
-- The ON CONFLICT clause handles conflicts based on the combination of UserID, PostID, and CommentID.

-- Query to count likes and dislikes for each post
SELECT 
    PostID, 
    SUM(CASE WHEN IsLike = TRUE THEN 1 ELSE 0 END) AS Likes, -- Count of likes
    SUM(CASE WHEN IsLike = FALSE THEN 1 ELSE 0 END) AS Dislikes -- Count of dislikes
FROM 
    PostLikes 
GROUP BY 
    PostID; -- Group results by PostID

-- Explanation:
-- This query calculates the total number of likes and dislikes for each post.
-- It uses the SUM function with CASE statements to count the number of TRUE (likes) and FALSE (dislikes) values for each PostID.
-- The results are grouped by PostID to aggregate the counts for each post.

-- Add CreatedAt column to Post table
ALTER TABLE Post ADD COLUMN CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP;

-- Explanation:
-- This statement adds a new column, CreatedAt, to the Post table.
-- The column is of type DATETIME and defaults to the current timestamp when a new row is inserted.
-- This column is intended to store the creation time of each post.

-- Create a trigger to set CreatedAt for Comment table
CREATE TRIGGER set_created_at
AFTER INSERT ON Comment
FOR EACH ROW
BEGIN
    UPDATE Comment
    SET CreatedAt = datetime('now')
    WHERE rowid = NEW.rowid;
END;

-- Explanation:
-- This trigger sets the CreatedAt column of the Comment table to the current timestamp whenever a new row is inserted.
-- The trigger updates the CreatedAt column for the newly inserted row (identified by NEW.rowid) to ensure it has the correct timestamp.

-- Select query to retrieve comments and associated usernames
SELECT c.Content, u.Username
FROM Comment c
JOIN User u ON c.UserID = u.UserID
WHERE c.PostID = ?;

-- Explanation:
-- This query retrieves the content of comments and the usernames of the users who made the comments.
-- It joins the Comment table with the User table based on the UserID to get the username.
-- The WHERE clause filters the results to only include comments related to a specific PostID (provided as a parameter).
