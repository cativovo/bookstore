package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

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
}

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		pool:    pool,
		queries: query.New(pool),
	}, nil
}

func (pr *PostgresRepository) CreateGenre(ctx context.Context, name string) error {
	var genreName pgtype.Text
	if err := genreName.Scan(name); err != nil {
		return err
	}

	_, err := withTimeout(ctx, func(ctxWithTimeout context.Context) (pgtype.UUID, error) {
		return pr.queries.CreateGenre(ctxWithTimeout, genreName)
	})
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

func (pr *PostgresRepository) DeleteGenre(ctx context.Context, id string) error {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return book.ErrNotFound
	}

	rows, err := withTimeout(ctx, func(ctxWithTimeout context.Context) (int64, error) {
		return pr.queries.DeleteGenre(ctxWithTimeout, uuid)
	})
	if err != nil {
		return err
	}

	if rows == 0 {
		return book.ErrNotFound
	}

	return nil
}

func (pr *PostgresRepository) CreateBook(ctx context.Context, b book.Book) (book.Book, error) {
	return withTimeout(ctx, func(ctxWithTimeout context.Context) (book.Book, error) {
		tx, err := pr.pool.Begin(ctxWithTimeout)
		if err != nil {
			return book.Book{}, err
		}
		defer tx.Rollback(ctxWithTimeout)
		qtx := pr.queries.WithTx(tx)

		genreUuids := make([]pgtype.UUID, 0)

		// check if genre exists in db
		for _, v := range b.Genres {
			name := pgtype.Text{String: v, Valid: true}
			genre, err := qtx.GetGenreByName(ctxWithTimeout, name)
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
		bookUuid, err := qtx.CreateBook(ctxWithTimeout, createBookParams)
		if err != nil {
			return book.Book{}, err
		}

		// create bookgenre
		for _, genreUuid := range genreUuids {
			err := qtx.CreateBookGenre(ctxWithTimeout, query.CreateBookGenreParams{
				BookID:  bookUuid,
				GenreID: genreUuid,
			})
			if err != nil {
				return book.Book{}, err
			}
		}

		if err := tx.Commit(ctxWithTimeout); err != nil {
			return book.Book{}, err
		}

		id, err := bookUuid.Value()
		if err != nil {
			return book.Book{}, err
		}

		b.Id = id.(string)

		return b, nil
	})
}

func (pr *PostgresRepository) GetBooks(ctx context.Context, opts book.GetBooksOptions) ([]book.Book, int, error) {
	genres := opts.Filter.Genres
	if len(genres) == 0 {
		genres = []string{"%%"}
	}

	row, err := withTimeout(ctx, func(ctxWithTimeout context.Context) (query.GetBooksRow, error) {
		return pr.queries.GetBooks(ctxWithTimeout, query.GetBooksParams{
			Limit:         int32(opts.Limit),
			Offset:        int32(opts.Offset),
			Descending:    opts.Desc,
			OrderBy:       opts.OrderBy,
			KeywordAuthor: appendPatternWildcard(opts.Filter.Author),
			KeywordTitle:  appendPatternWildcard(opts.Filter.Title),
			Genres:        genres,
		})
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

func (pr *PostgresRepository) GetGenres(ctx context.Context) ([]string, error) {
	genreRows, err := withTimeout(ctx, func(ctxWithTimeout context.Context) ([]pgtype.Text, error) {
		return pr.queries.GetGenres(ctxWithTimeout)
	})
	if err != nil {
		return nil, err
	}

	genres := make([]string, len(genreRows))

	for i, v := range genreRows {
		genres[i] = v.String
	}

	return genres, nil
}

func (pr *PostgresRepository) GetBookById(ctx context.Context, id string) (book.Book, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return book.Book{}, book.ErrNotFound
	}

	b, err := withTimeout(ctx, func(ctxWithTimeout context.Context) (query.GetBookByIdRow, error) {
		return pr.queries.GetBookById(ctxWithTimeout, uuid)
	})
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

func appendPatternWildcard(s string) string {
	return fmt.Sprintf("%%%s%%", s)
}

type withTimeoutResult[T any] struct {
	result T
	err    error
}

func withTimeout[T any](ctx context.Context, cb func(ctxWithTimeout context.Context) (T, error)) (T, error) {
	timeout := time.Second * 5
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	withoutTimeoutChan := make(chan withTimeoutResult[T])

	go func() {
		result, err := cb(ctxWithTimeout)
		withoutTimeoutChan <- withTimeoutResult[T]{
			err:    err,
			result: result,
		}
	}()

	select {
	case <-ctxWithTimeout.Done():
		var defaultValue T
		return defaultValue, errors.New("operation timed out")
	case result := <-withoutTimeoutChan:
		return result.result, result.err
	}
}
