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

const getBookById = `-- name: GetBookById :one
SELECT
  book.id,
  book.title,
  book.description,
  book.author,
  book.price,
  book.cover_image,
  COALESCE(ARRAY_AGG(genre.name) FILTER (WHERE genre.name IS NOT NULL), '{}') AS genres
FROM
  book
LEFT JOIN
  book_genre ON book_genre.book_id = book.id
LEFT JOIN
  genre ON genre.id = book_genre.genre_id
WHERE
  book.id = $1
GROUP BY
  book.id
`

type GetBookByIdRow struct {
	ID          pgtype.UUID
	Title       string
	Description pgtype.Text
	Author      string
	Price       pgtype.Numeric
	CoverImage  pgtype.Text
	Genres      interface{}
}

func (q *Queries) GetBookById(ctx context.Context, id pgtype.UUID) (GetBookByIdRow, error) {
	row := q.db.QueryRow(ctx, getBookById, id)
	var i GetBookByIdRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Author,
		&i.Price,
		&i.CoverImage,
		&i.Genres,
	)
	return i, err
}

const getBooks = `-- name: GetBooks :one
SELECT (
  SELECT
    COUNT(DISTINCT book.id)
  FROM
    book
  LEFT JOIN
    book_genre ON book_genre.book_id = book.id
  LEFT JOIN
    genre ON genre.id = book_genre.genre_id
  WHERE 
    book.author ILIKE $3
  AND
    book.title ILIKE $4
  AND
    book.id
  IN
    (
      SELECT
      book_genre.book_id
      FROM
        genre
      INNER JOIN
        book_genre
      ON
        book_genre.genre_id = genre.id 
      AND
        genre.name ILIKE ANY($5::text[])
      GROUP BY 1
    )
) AS count,
(
  SELECT 
    JSON_AGG(rows.*)
  FROM
    (
      SELECT
        book.id AS id,
        book.title AS title,
        book.description AS description,
        book.author AS author,
        book.price AS price,
        book.cover_image AS cover_image,
        COALESCE(ARRAY_AGG(genre.name) FILTER (WHERE genre.name IS NOT NULL), '{}') AS genres
      FROM
        book
      LEFT JOIN
        book_genre ON book_genre.book_id = book.id
      LEFT JOIN
        genre ON genre.id = book_genre.genre_id
      WHERE 
        book.author ILIKE $3
      AND
        book.title ILIKE $4
      AND
        book.id
      IN
        (
          SELECT
          book_genre.book_id
          FROM
            genre
          INNER JOIN
            book_genre
          ON
            book_genre.genre_id = genre.id 
          AND
            genre.name ILIKE ANY($5::text[])
          GROUP BY 1
        )
      GROUP BY
        book.id
      ORDER BY 
        -- will produce title ASC/DESC, author ASC/DESC OR author ASC/DESC, title ASC/DESC
        CASE
          WHEN $6::boolean AND $7::text = 'title' THEN title
          WHEN $6::boolean AND $7::text = 'author' THEN author
          WHEN $6::boolean THEN title
        END DESC,
        CASE
          WHEN $6::boolean AND $7::text = 'author' THEN title
          WHEN $6::boolean THEN author
        END DESC,
        CASE
          WHEN NOT $6::boolean AND $7::text = 'title' THEN title
          WHEN NOT $6::boolean AND $7::text = 'author' THEN author
          WHEN NOT $6::boolean THEN title
        END ASC,
        CASE
          WHEN NOT $6::boolean AND $7::text = 'author' THEN title
          WHEN NOT $6::boolean THEN author
        END ASC
      LIMIT 
        $1
      OFFSET 
        $2
    ) AS rows
) AS books
`

type GetBooksParams struct {
	Limit         int32
	Offset        int32
	KeywordAuthor string
	KeywordTitle  string
	Genres        []string
	Descending    bool
	OrderBy       string
}

type GetBooksRow struct {
	Count int64
	Books []byte
}

func (q *Queries) GetBooks(ctx context.Context, arg GetBooksParams) (GetBooksRow, error) {
	row := q.db.QueryRow(ctx, getBooks,
		arg.Limit,
		arg.Offset,
		arg.KeywordAuthor,
		arg.KeywordTitle,
		arg.Genres,
		arg.Descending,
		arg.OrderBy,
	)
	var i GetBooksRow
	err := row.Scan(&i.Count, &i.Books)
	return i, err
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

const getGenres = `-- name: GetGenres :many
SELECT name FROM genre
`

func (q *Queries) GetGenres(ctx context.Context) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, getGenres)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var name pgtype.Text
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const test = `-- name: test :many
SELECT name FROM genre where name ilike $1::text[]
`

func (q *Queries) test(ctx context.Context, genres []string) ([]pgtype.Text, error) {
	rows, err := q.db.Query(ctx, test, genres)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Text
	for rows.Next() {
		var name pgtype.Text
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
