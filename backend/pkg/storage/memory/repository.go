package memory

import (
	"slices"
	"strconv"

	"github.com/cativovo/bookstore/pkg/book"
)

type MemoryRepository struct {
	Books  map[string]book.Book
	Genres []string
}

var bookId = 0

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		Books:  make(map[string]book.Book),
		Genres: make([]string, 0),
	}
}

func (m *MemoryRepository) CreateGenre(name string) error {
	if slices.Contains(m.Genres, name) {
		return book.ErrAlreadyExists
	}
	m.Genres = append(m.Genres, name)
	return nil
}

func (m *MemoryRepository) DeleteGenre(name string) error {
	genres := make([]string, 0)

	for _, v := range m.Genres {
		if v != name {
			genres = append(genres, v)
		}
	}

	if len(genres) == len(m.Genres) {
		return book.ErrNotFound
	}

	m.Genres = genres

	return nil
}

func (m *MemoryRepository) CreateBook(b book.Book) (book.Book, error) {
	bookId++
	b.Id = strconv.Itoa(bookId)
	return b, nil
}

func (m *MemoryRepository) Cleanup() {
	m.Books = make(map[string]book.Book)
	m.Genres = make([]string, 0)
	bookId = 0
}
