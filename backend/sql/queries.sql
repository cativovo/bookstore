-- name: CreateGenre :one
INSERT INTO genre (
  name
) VALUES ( 
  $1 
)
RETURNING id;

-- name: DeleteGenre :execrows
DELETE FROM genre WHERE id = $1;

-- name: GetGenreByName :one
SELECT * FROM genre WHERE name = $1;

-- name: CreateBook :one
INSERT INTO book (
  title, author, description, price, cover_image
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id;

-- name: CreateBookGenre :exec
INSERT INTO book_genre (
  book_id, genre_id
) VALUES ( 
  $1, $2
);

-- name: GetBooks :one
SELECT (
  SELECT
    COUNT(id)
  FROM
    book
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
      GROUP BY
        book.id
      ORDER BY
        title ASC
      LIMIT 
        $1
      OFFSET 
        $2
    ) AS rows
) AS books;

-- name: GetGenres :many
SELECT name FROM genre;
