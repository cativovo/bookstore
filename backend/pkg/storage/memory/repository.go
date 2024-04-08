package memory

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cativovo/bookstore/pkg/book"
)

type MemoryRepository struct {
	Books       []book.Book
	Genres      []string
	ReturnError bool
}

var bookId = 0

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		Books:  make([]book.Book, 0),
		Genres: make([]string, 0),
	}
}

func (m *MemoryRepository) CreateGenre(name string) error {
	if m.ReturnError {
		return errors.New("error")
	}
	if slices.Contains(m.Genres, name) {
		return book.ErrAlreadyExists
	}
	m.Genres = append(m.Genres, name)
	return nil
}

func (m *MemoryRepository) DeleteGenre(name string) error {
	if m.ReturnError {
		return errors.New("error")
	}

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
	if m.ReturnError {
		return book.Book{}, errors.New("error")
	}

	bookId++
	b.Id = strconv.Itoa(bookId)
	return b, nil
}

func (m *MemoryRepository) GetBooks(options book.GetBooksOptions) ([]book.Book, int, error) {
	if m.ReturnError {
		return nil, 0, errors.New("error")
	}
	return m.Books[options.Offset : options.Offset+options.Limit], len(m.Books), nil
}

func (m *MemoryRepository) GetBookById(id string) (book.Book, error) {
	if m.ReturnError {
		return book.Book{}, errors.New("error")
	}

	i := slices.IndexFunc(m.Books, func(b book.Book) bool {
		return b.Id == id
	})

	if i < 0 {
		return book.Book{}, book.ErrNotFound
	}

	return m.Books[i], nil
}

func (m *MemoryRepository) GetGenres() ([]string, error) {
	if m.ReturnError {
		return nil, errors.New("error")
	}
	return m.Genres, nil
}

func (m *MemoryRepository) Seed() {
	const (
		genreCount = 6
		bookCount  = 101
	)

	m.Genres = make([]string, 0, genreCount)

	for len(m.Genres) < genreCount {
		genre := gofakeit.BookGenre()
		if !slices.Contains(m.Genres, genre) {
			m.Genres = append(m.Genres, genre)
		}
	}

	for len(m.Books) < bookCount {
		genreCount := rand.IntN(len(m.Genres))
		bookGenres := make([]string, 0, genreCount)

		for len(bookGenres) != genreCount {
			randomIndex := rand.IntN(len(m.Genres))
			genre := m.Genres[randomIndex]
			if !slices.Contains(bookGenres, genre) {
				bookGenres = append(bookGenres, genre)
			}
		}

		bookId++
		b := book.Book{
			Id:          strconv.Itoa(bookId),
			Title:       gofakeit.BookTitle(),
			Author:      gofakeit.BookAuthor(),
			Description: gofakeit.Product().Description,
			CoverImage:  "https://placehold.co/600x400",
			Price:       gofakeit.Price(0.99, 69.99),
			Genres:      bookGenres,
		}

		m.Books = append(m.Books, b)
	}
}

func (m *MemoryRepository) Cleanup() {
	m.Books = make([]book.Book, 0)
	m.Genres = make([]string, 0)
	m.ReturnError = false
	bookId = 0
}
