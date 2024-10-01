package data

import (
	"context"
	"database/sql"
	"errors"
	"oynas/internal/validator"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"-"`
	ToyID     int64     `json:"toy_id"`
	UserName  string    `json:"user_name"`
	Text      string    `json:"text"`
	Rating    Rating    `json:"rating"`
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Text != "", "text", "text must be provided")
	v.Check(len(comment.Text) <= 1000, "text", "text must not be bigger than 1000 bytes")
}

var (
	ErrDuplicateComment = errors.New("duplicate comment")
)

type CommentModel struct {
	DB *sql.DB
}

func (c CommentModel) Insert(comment *Comment) error {
	query := `INSERT INTO comments (toy_id, user_name, text, rating)
	VALUES ($1, $2, $3, $4) 
	RETURNING id, created_at
	`

	args := []any{comment.ToyID, comment.UserName, comment.Text, comment.Rating}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == "pq: повторяющееся значение ключа нарушает ограничение уникальности \"users_email_key\"":
			return ErrDuplicateEmail
		default:
			return err
		}

	}
	return nil
}

func (c CommentModel) GetCommentsFromId(id int64, text string, rating Rating) ([]*Comment, error) {
	query := `SELECT id, toy_id, text, rating, user_name from comments 
WHERE toy_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []*Comment{}

	for rows.Next() {
		var comment Comment

		err := rows.Scan(
			&comment.ID,
			&comment.ToyID,
			&comment.Text,
			&comment.Rating,
			&comment.UserName,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, err

}
