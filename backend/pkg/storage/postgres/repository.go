package postgres

import (
	"context"

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
