package book

import "errors"

var ErrNotFound = errors.New("not found")

type BookRepository interface {
	// GetBooks() ([]Book, error)
	CreateGenre(name string) (Genre, error)
	DeleteGenre(id string) error
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

func (bs *BookService) CreateGenre(name string) (Genre, error) {
	return bs.repository.CreateGenre(name)
}

func (bs *BookService) DeleteGenre(id string) error {
	return bs.repository.DeleteGenre(id)
}
