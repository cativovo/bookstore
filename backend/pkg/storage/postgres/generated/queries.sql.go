// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createBook = `-- name: CreateBook :one
INSERT INTO book (
  title, author, description, price, cover_image
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id
`

type CreateBookParams struct {
	Title       string
	Author      string
	Description pgtype.Text
	Price       pgtype.Numeric
	CoverImage  pgtype.Text
}

func (q *Queries) CreateBook(ctx context.Context, arg CreateBookParams) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createBook,
		arg.Title,
		arg.Author,
		arg.Description,
		arg.Price,
		arg.CoverImage,
	)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const createBookGenre = `-- name: CreateBookGenre :exec
INSERT INTO book_genre (
  book_id, genre_id
) VALUES ( 
  $1, $2
)
`

type CreateBookGenreParams struct {
	BookID  pgtype.UUID
	GenreID pgtype.UUID
}

func (q *Queries) CreateBookGenre(ctx context.Context, arg CreateBookGenreParams) error {
	_, err := q.db.Exec(ctx, createBookGenre, arg.BookID, arg.GenreID)
	return err
}

const createGenre = `-- name: CreateGenre :one
INSERT INTO genre (
  name
) VALUES ( 
  $1 
)
RETURNING id
`

func (q *Queries) CreateGenre(ctx context.Context, name pgtype.Text) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createGenre, name)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const deleteGenre = `-- name: DeleteGenre :execrows
DELETE FROM genre WHERE id = $1
`

func (q *Queries) DeleteGenre(ctx context.Context, id pgtype.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteGenre, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getGenreByName = `-- name: GetGenreByName :one
SELECT id, name FROM genre WHERE name = $1
`

func (q *Queries) GetGenreByName(ctx context.Context, name pgtype.Text) (Genre, error) {
	row := q.db.QueryRow(ctx, getGenreByName, name)
	var i Genre
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
