package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/cativovo/bookstore/internal/book"
	query "github.com/cativovo/bookstore/internal/storage/postgres/generated"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *query.Queries
	ctx     context.Context
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		pool:    pool,
		queries: query.New(pool),
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
	tx, err := pr.pool.Begin(pr.ctx)
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
			switch err {
			case pgx.ErrNoRows:
				return book.Book{}, book.ErrNotFound
			default:
				return book.Book{}, err
			}
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

func (pr *PostgresRepository) GetBooks(opts book.GetBooksOptions) ([]book.Book, int, error) {
	row, err := pr.queries.GetBooks(pr.ctx, query.GetBooksParams{
		Limit:      int32(opts.Limit),
		Offset:     int32(opts.Offset),
		OrderBy:    opts.OrderBy,
		Descending: opts.Desc,
		FilterBy:   opts.Filter.By,
		Keyword:    fmt.Sprintf("%%%s%%", opts.Filter.Keyword),
	})
	if err != nil {
		return nil, 0, err
	}

	books := make([]book.Book, 0)

	if len(row.Books) > 0 {
		if err := json.Unmarshal(row.Books, &books); err != nil {
			return nil, 0, err
		}
	}

	return books, int(row.Count), nil
}

func (pr *PostgresRepository) GetGenres() ([]string, error) {
	genreRows, err := pr.queries.GetGenres(pr.ctx)
	if err != nil {
		return nil, err
	}

	genres := make([]string, len(genreRows))

	for i, v := range genreRows {
		genres[i] = v.String
	}

	return genres, nil
}

func (pr *PostgresRepository) GetBookById(id string) (book.Book, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return book.Book{}, book.ErrNotFound
	}

	b, err := pr.queries.GetBookById(pr.ctx, uuid)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return book.Book{}, book.ErrNotFound
		default:
			return book.Book{}, err
		}
	}

	priceFloatValue, err := b.Price.Float64Value()
	if err != nil {
		return book.Book{}, err
	}

	genresInterface := b.Genres.([]interface{})
	genres := make([]string, len(genresInterface))

	for i, genre := range genresInterface {
		genres[i] = genre.(string)
	}

	return book.Book{
		Id:          id,
		Author:      b.Author,
		Title:       b.Title,
		Price:       priceFloatValue.Float64,
		CoverImage:  b.CoverImage.String,
		Description: b.Description.String,
		Genres:      genres,
	}, nil
}
