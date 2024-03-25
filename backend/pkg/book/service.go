package book

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type BookRepository interface {
	// GetBooks() ([]Book, error)
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
