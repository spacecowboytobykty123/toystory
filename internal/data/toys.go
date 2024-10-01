package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"oynas/internal/validator"
	"time"
)

var AnonymousUser = &User{}

type Toy struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"-"`
	Title          string    `json:"title"`
	Description    string    `json:"desc"`
	Details        []string  `json:"details,omitempty"`
	Skills         []string  `json:"skills"`
	Images         []string  `json:"image"`
	Categories     []string  `json:"categories"`
	RecommendedAge string    `json:"recommended_age"`
	Manufacturer   string    `json:"manufacturer"`
	Value          int64     `json:"value"`
	IsAvailable    string    `json:"isAvailable"`
	WaitList       []string  `json:"waitList,omitempty"`
	Comments       []Comment `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateToy(v *validator.Validator, toy *Toy) {
	v.Check(toy.Title != "", "title", "title must be provided")
	v.Check(len(toy.Title) <= 500, "title", "title must not be more than 500 bytes long")
	v.Check(len(toy.Description) <= 5000, "desc", "Description must not be more than 5000 bytes long")
	v.Check(len(toy.Details) <= 5, "details", "details must not be more than 5")
	v.Check(v.ImageUrlsCheck(toy.Images), "image", "some of image urls is wrong")
	v.Check(toy.Categories != nil, "categories", "categories must be provided")
	v.Check(toy.Skills != nil, "skills", "skills must be provided")
	v.Check(len(toy.Categories) >= 1, "categories", "at least 1 category")
	v.Check(len(toy.Skills) >= 1, "skills", "at least 1 skill")
	v.Check(len(toy.Categories) <= 7, "categories", "no more than 7 categories")
	v.Check(len(toy.Skills) <= 7, "Skills", "no more than 7 skills")
	v.Check(validator.Unique(toy.Categories), "categories", "categories should not contain duplicate values")
	v.Check(validator.Unique(toy.Skills), "skills", "skills should not contain duplicate values")
	v.Check(toy.RecommendedAge != "", "recAge", "age must be provided")
	v.Check(toy.Manufacturer != "", "manufacturer", "manufacturer must be provided")
	v.Check(toy.Value >= 1000, "value", "toy value must be more than 1000 tenge")
	v.Check(toy.Value <= 150000, "value", "limit of toy's value is 150.000 tenge")
}

type ToyModel struct {
	DB *sql.DB
}

func (t ToyModel) Insert(toy *Toy) error {
	query := `
INSERT INTO toys (title, description, details, skills, categories, images, recommended_age, manufacturer, value, is_available, wait_list)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, created_at`

	args := []any{toy.Title, toy.Description, pq.Array(toy.Details), pq.Array(toy.Skills), pq.Array(toy.Categories), pq.Array(toy.Images), toy.RecommendedAge, toy.Manufacturer, toy.Value, toy.IsAvailable, pq.Array(toy.WaitList)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return t.DB.QueryRowContext(ctx, query, args...).Scan(&toy.ID, &toy.CreatedAt)
}

func (t ToyModel) Get(id int64) (*Toy, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
SELECT id, created_at, title, description, details ,skills, categories, images, recommended_age, manufacturer, value, is_available, wait_list
FROM toys
WHERE id = $1
`

	var toy Toy

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, id).Scan(
		&toy.ID,
		&toy.CreatedAt,
		&toy.Title,
		&toy.Description,
		pq.Array(&toy.Details),
		pq.Array(&toy.Skills),
		pq.Array(&toy.Categories),
		pq.Array(&toy.Images),
		&toy.RecommendedAge,
		&toy.Manufacturer,
		&toy.Value,
		&toy.IsAvailable,
		pq.Array(&toy.WaitList),
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &toy, nil
}

func (t ToyModel) Update(toy *Toy) error {

	query := `UPDATE toys
SET title = $1, description = $2, details = $3, skills = $4, images = $5, categories = $6, recommended_age = $7, manufacturer = $8, value = $9
WHERE id = $10
RETURNING id
`
	args := []any{
		toy.Title,
		toy.Description,
		pq.Array(toy.Details),
		pq.Array(toy.Skills),
		pq.Array(toy.Images),
		pq.Array(toy.Categories),
		toy.RecommendedAge,
		toy.Manufacturer,
		toy.Value,
		toy.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, args...).Scan(&toy.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}

	}
	return nil

}

func (t ToyModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
DELETE FROM toys
WHERE id = $1
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil

}

func (t ToyModel) GetAll(title string, skills []string, categories []string, value int64, from int64, to int64, filters Filters) ([]*Toy, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, title, categories, skills, recommended_age, value from toys 
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (categories @> $2 OR $2 = '{}')
AND (skills @> $3 OR $3 = '{}')
AND (value BETWEEN $4 and $5)
ORDER BY %s %s, id ASC
LIMIT $6 OFFSET $7`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, pq.Array(categories), pq.Array(skills), from, to, filters.limit(), filters.offset()}

	rows, err := t.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	toys := []*Toy{}

	for rows.Next() {
		var toy Toy

		err := rows.Scan(
			&totalRecords,
			&toy.ID,
			&toy.Title,
			pq.Array(&toy.Categories),
			pq.Array(&toy.Skills),
			&toy.RecommendedAge,
			&toy.Value,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		toys = append(toys, &toy)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := filters.calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return toys, metadata, err
}
