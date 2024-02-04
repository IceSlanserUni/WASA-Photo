package database

import (
	"database/sql"
)

func (db *appdbimpl) GetLikes(myUID uint64, postID uint64) ([]uint64, error) {
	// Store likes
	query := `SELECT OwnerID 
				FROM likes
				WHERE PostID = ? AND OwnerID NOT IN (SELECT BannerUID FROM bans WHERE BannedUID = ?)`
	rows, err := db.c.Query(query, postID, myUID)
	var likeOwners []uint64
	for rows.Next() {
		var ownerID uint64
		err = rows.Scan(&ownerID)
		if err != nil {
			return nil, err
		}
		likeOwners = append(likeOwners, ownerID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return likeOwners, nil
}

func (db *appdbimpl) PutLike(UID uint64, postID uint64) (uint64, bool, error) {

	// Try to insert a like
	var like Like
	res, err := db.c.Exec("INSERT INTO likes(PostID, OwnerID) VALUES (?, ?)",
		postID, UID)
	if err != nil {
		err = db.c.QueryRow("SELECT * FROM likes WHERE PostID = ? AND OwnerID = ?",
			postID, UID).Scan(&like.ID, &like.PostID, &like.OwnerID)
		// There is already an existent like
		if err != nil {
			if err == sql.ErrNoRows {
				return like.ID, false, err
			}
		}
		return like.ID, true, nil
	}

	// Get the new like ID
	ID, err := res.LastInsertId()
	if err != nil {
		return like.ID, false, err
	}
	like.ID = uint64(ID)
	return like.ID, false, nil
}

func (db *appdbimpl) DeleteLike(UID uint64, likeID uint64) (bool, error) {
	// Check if there is an existent like
	err := db.c.QueryRow("SELECT * FROM likes WHERE ID = ? AND OwnerID = ?", likeID, UID).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}
	}

	_, err = db.c.Exec("DELETE FROM likes WHERE ID = ?", likeID)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (db *appdbimpl) GetComments(myUID uint64, postID uint64) ([]Comment, error) {
	// Store comments
	query := `SELECT * 
				FROM comments 
				WHERE PostID = ? AND OwnerID NOT IN (SELECT BannerUID FROM bans WHERE BannedUID = ?)`
	rows, err := db.c.Query(query, postID, myUID)
	var comments []Comment
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.ID, &comment.PostID, &comment.OwnerID, &comment.Text, &comment.DateTime)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (db *appdbimpl) PostComment(UID uint64, postID uint64, text string) (uint64, error) {
	// Post new comment
	res, err := db.c.Exec("INSERT INTO comments VALUES (?, ?, ?)", postID, UID, text)
	if err != nil {
		return 0, err
	}

	// Get new comment ID
	ID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(ID), nil
}

func (db *appdbimpl) DeleteComment(UID uint64, commentID uint64) (bool, error) {
	// Check if there is an existent comment
	var ownerID uint64
	var postID uint64
	var userID uint64
	err := db.c.QueryRow("SELECT PostID, OwnerID FROM comments WHERE ID = ?", commentID).Scan(&postID, &ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}
	}

	// Find the post owner
	err = db.c.QueryRow("SELECT ProfileID FROM posts WHERE ID = ?", postID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}
	}

	// Check if the user are allowed to delete that comment
	if ownerID == UID || ownerID == userID {
		_, err = db.c.Exec("DELETE FROM comments WHERE ID = ?", commentID)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

/** func (db *appdbimpl) getPostOwner(postID uint64) (uint64, error) {
	var owner User

	err := db.c.QueryRow("SELECT profileID FROM posts WHERE ID = ?", postID).Scan(&owner.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return owner.ID, err
		}
	}
	return owner.ID, nil
} */
