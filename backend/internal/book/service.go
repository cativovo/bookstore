package book

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type GetBooksFilter struct {
	Author string
	Title  string
	Genres []string
}

type GetBooksOptions struct {
	OrderBy string
	Filter  GetBooksFilter
	Limit   int
	Offset  int
	Desc    bool
}

type BookRepository interface {
	GetBooks(options GetBooksOptions) (books []Book, count int, err error)
	GetBookById(id string) (Book, error)
	GetGenres() ([]string, error)
	CreateGenre(name string) error
	DeleteGenre(name string) error
	CreateBook(b Book) (Book, error)
}

type BookService struct {
	repository BookRepository
}

func NewBookService(r BookRepository) *BookService {
	return &BookService{
		repository: r,
	}
}

// func (bs *BookService) GetBooks() ([]Book, error) {
// 	return bs.repository.GetBooks()
// }

func (bs *BookService) CreateGenre(name string) error {
	return bs.repository.CreateGenre(name)
}

func (bs *BookService) DeleteGenre(name string) error {
	return bs.repository.DeleteGenre(name)
}

func (bs *BookService) CreateBook(b Book) (Book, error) {
	return bs.repository.CreateBook(b)
}

func (bs *BookService) GetBooks(options GetBooksOptions) (books []Book, count int, err error) {
	return bs.repository.GetBooks(options)
}

func (bs *BookService) GetBookById(id string) (Book, error) {
	return bs.repository.GetBookById(id)
}

func (bs *BookService) GetGenres() ([]string, error) {
	return bs.repository.GetGenres()
}
