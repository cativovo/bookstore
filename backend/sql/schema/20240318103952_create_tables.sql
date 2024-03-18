-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE book (
  id  UUID DEFAULT uuid_generate_v4(),
  title VARCHAR(255) NOT NULL,
  author VARCHAR(255) NOT NULL,
  description TEXT,
  cover_image VARCHAR(255),
  price DECIMAL NOT NULL,
  PRIMARY KEY(id)
);

CREATE TABLE genre (
  id UUID DEFAULT uuid_generate_v4(),
  name VARCHAR(255),
  PRIMARY KEY(id)
);

CREATE TABLE book_genre(
  id UUID DEFAULT uuid_generate_v4(),
  book_id UUID,
  genre_id UUID,
  FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE,
  FOREIGN KEY (genre_id) REFERENCES genre(id) ON DELETE CASCADE,
  PRIMARY KEY(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE book;
DROP TABLE genre;
DROP TABLE book_genre;
-- +goose StatementEnd
