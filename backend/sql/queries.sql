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
