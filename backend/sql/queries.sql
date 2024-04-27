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
WITH
filtered_books as (
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
      book.author ILIKE @keyword_author
    AND
      book.title ILIKE @keyword_title
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
          genre.name ILIKE ANY(@genres::text[])
        GROUP BY 1
      )
    GROUP BY
      book.id
)
SELECT (
  SELECT
    COUNT(filtered_books.id)
  FROM
  filtered_books
) AS count,
(
  SELECT 
    JSON_AGG(rows.*)
  FROM
    (
      select 
        id,
        title,
        description,
        author,
        price,
        cover_image,
        genres
      from 
        filtered_books
      ORDER BY 
        -- will produce title ASC/DESC, author ASC/DESC OR author ASC/DESC, title ASC/DESC
        CASE
          WHEN @descending::boolean AND @order_by::text = 'title' THEN title
          WHEN @descending::boolean AND @order_by::text = 'author' THEN author
          WHEN @descending::boolean THEN title
        END DESC,
        CASE
          WHEN @descending::boolean AND @order_by::text = 'author' THEN title
          WHEN @descending::boolean THEN author
        END DESC,
        CASE
          WHEN NOT @descending::boolean AND @order_by::text = 'title' THEN title
          WHEN NOT @descending::boolean AND @order_by::text = 'author' THEN author
          WHEN NOT @descending::boolean THEN title
        END ASC,
        CASE
          WHEN NOT @descending::boolean AND @order_by::text = 'author' THEN title
          WHEN NOT @descending::boolean THEN author
        END ASC
      LIMIT 
        $1
      OFFSET 
        $2
    ) AS rows
) AS books;

-- name: GetBookById :one
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
  book.id;

-- name: GetGenres :many
SELECT name FROM genre;

-- name: test :many
SELECT name FROM genre where name ilike @genres::text[];
