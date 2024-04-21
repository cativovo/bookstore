package book

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type GetBooksFilter struct {
	Author string
	Title  string
	// Genres []string
}

type GetBooksOptions struct {
	OrderBy string
	Filter  GetBooksFilter
	Limit   int
	Offset  int
	Desc    bool
}

type BookRepository interface {
	GetBooks(ctx context.Context, options GetBooksOptions) (books []Book, count int, err error)
	GetBookById(ctx context.Context, id string) (Book, error)
	GetGenres(ctx context.Context) ([]string, error)
	CreateGenre(ctx context.Context, name string) error
	DeleteGenre(ctx context.Context, name string) error
	CreateBook(ctx context.Context, b Book) (Book, error)
}

type BookService struct {
	repository BookRepository
}

func NewBookService(r BookRepository) *BookService {
	return &BookService{
		repository: r,
	}
}

func (bs *BookService) CreateGenre(ctx context.Context, name string) error {
	return bs.repository.CreateGenre(ctx, name)
}

func (bs *BookService) DeleteGenre(ctx context.Context, name string) error {
	return bs.repository.DeleteGenre(ctx, name)
}

func (bs *BookService) CreateBook(ctx context.Context, b Book) (Book, error) {
	return bs.repository.CreateBook(ctx, b)
}

func (bs *BookService) GetBooks(ctx context.Context, options GetBooksOptions) (books []Book, count int, err error) {
	return bs.repository.GetBooks(ctx, options)
}

func (bs *BookService) GetBookById(ctx context.Context, id string) (Book, error) {
	return bs.repository.GetBookById(ctx, id)
}

func (bs *BookService) GetGenres(ctx context.Context) ([]string, error) {
	return bs.repository.GetGenres(ctx)
}
