package postgres

import (
	"context"
	"strconv"

	"github.com/cativovo/bookstore/pkg/book"
	query "github.com/cativovo/bookstore/pkg/storage/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresRepository struct {
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
		queries: query.New(conn),
		ctx:     ctx,
	}, nil
}

func (pr *PostgresRepository) CreateGenre(name string) (book.Genre, error) {
	var genreName pgtype.Text
	if err := genreName.Scan(name); err != nil {
		return book.Genre{}, err
	}

	uuid, err := pr.queries.CreateGenre(pr.ctx, genreName)
	if err != nil {
		return book.Genre{}, err
	}

	id, err := uuid.Value()
	if err != nil {
		return book.Genre{}, err
	}

	return book.Genre{
		Id:   id.(string),
		Name: name,
	}, nil
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
	var description pgtype.Text
	if err := description.Scan(b.Description); err != nil {
		return book.Book{}, err
	}

	var price pgtype.Numeric
	if err := price.Scan(strconv.FormatFloat(b.Price, 'f', 2, 64)); err != nil {
		return book.Book{}, err
	}

	var coverImage pgtype.Text
	if err := coverImage.Scan(b.CoverImage); err != nil {
		return book.Book{}, err
	}

	createBookParams := query.CreateBookParams{
		Title:       b.Title,
		Author:      b.Author,
		Description: description,
		Price:       price,
		CoverImage:  coverImage,
	}
	uuid, err := pr.queries.CreateBook(pr.ctx, createBookParams)
	if err != nil {
		return book.Book{}, err
	}

	id, err := uuid.Value()
	if err != nil {
		return book.Book{}, err
	}

	b.Id = id.(string)
	return b, nil
}
