package postgres

import (
	"context"
	"errors"
	"strconv"

	"github.com/cativovo/bookstore/pkg/book"
	query "github.com/cativovo/bookstore/pkg/storage/postgres/generated"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresRepository struct {
	conn    *pgx.Conn
	queries *query.Queries
	ctx     context.Context
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		conn:    conn,
		queries: query.New(conn),
		ctx:     ctx,
	}, nil
}

func (pr *PostgresRepository) CreateGenre(name string) error {
	var genreName pgtype.Text
	if err := genreName.Scan(name); err != nil {
		return err
	}

	_, err := pr.queries.CreateGenre(pr.ctx, genreName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return book.ErrAlreadyExists
			}
		}

		return err
	}

	return nil
}

func (pr *PostgresRepository) DeleteGenre(id string) error {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return book.ErrNotFound
	}

	rows, err := pr.queries.DeleteGenre(pr.ctx, uuid)
	if err != nil {
		return err
	}

	if rows == 0 {
		return book.ErrNotFound
	}

	return nil
}

func (pr *PostgresRepository) CreateBook(b book.Book) (book.Book, error) {
	tx, err := pr.conn.Begin(pr.ctx)
	if err != nil {
		return book.Book{}, err
	}
	defer tx.Rollback(pr.ctx)
	qtx := pr.queries.WithTx(tx)

	genreUuids := make([]pgtype.UUID, 0)

	// check if genre exists in db
	for _, v := range b.Genres {
		name := pgtype.Text{String: v, Valid: true}
		genre, err := qtx.GetGenreByName(pr.ctx, name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return book.Book{}, book.ErrNotFound
			}

			return book.Book{}, err
		}

		genreUuids = append(genreUuids, genre.ID)
	}

	// create book
	description := pgtype.Text{String: b.Description, Valid: true}
	coverImage := pgtype.Text{String: b.CoverImage, Valid: true}

	var price pgtype.Numeric
	if err := price.Scan(strconv.FormatFloat(b.Price, 'f', 2, 64)); err != nil {
		return book.Book{}, err
	}

	createBookParams := query.CreateBookParams{
		Title:       b.Title,
		Author:      b.Author,
		Description: description,
		Price:       price,
		CoverImage:  coverImage,
	}
	bookUuid, err := qtx.CreateBook(pr.ctx, createBookParams)
	if err != nil {
		return book.Book{}, err
	}

	// create bookgenre
	for _, genreUuid := range genreUuids {
		err := qtx.CreateBookGenre(pr.ctx, query.CreateBookGenreParams{
			BookID:  bookUuid,
			GenreID: genreUuid,
		})
		if err != nil {
			return book.Book{}, err
		}
	}

	if err := tx.Commit(pr.ctx); err != nil {
		return book.Book{}, err
	}

	id, err := bookUuid.Value()
	if err != nil {
		return book.Book{}, err
	}

	b.Id = id.(string)

	return b, nil
}
