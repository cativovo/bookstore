package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"slices"
	"sync"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cativovo/bookstore/internal/book"
	"github.com/cativovo/bookstore/internal/storage/postgres"
)

const (
	BOOK_COUNT  = 1000
	GENRE_COUNT = 10
)

func main() {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	repository, err := postgres.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	log.Println("seeding genres...")
	genres := getGenres()
	for _, genre := range genres {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := repository.CreateGenre(genre); err != nil {
				log.Printf("%s %s", genre, err)
			}
		}()
	}
	wg.Wait()
	log.Println("seeding genres completed")

	log.Println("seeding books...")
	for range BOOK_COUNT {
		wg.Add(1)
		go func() {
			defer wg.Done()

			genreCount := rand.IntN(len(genres))
			bookGenres := make([]string, 0, genreCount)

			for len(bookGenres) != genreCount {
				randomIndex := rand.IntN(len(genres))
				genre := genres[randomIndex]
				if !slices.Contains(bookGenres, genre) {
					bookGenres = append(bookGenres, genre)
				}
			}

			b := book.Book{
				Title:       gofakeit.BookTitle(),
				Author:      gofakeit.BookAuthor(),
				Description: gofakeit.Product().Description,
				CoverImage:  "https://placehold.co/600x400",
				Price:       gofakeit.Price(0.99, 69.99),
				Genres:      bookGenres,
			}
			repository.CreateBook(b)
		}()
	}

	wg.Wait()
	log.Println("seeding books completed")
}

func getGenres() []string {
	genres := make([]string, 0, GENRE_COUNT)

	for len(genres) != GENRE_COUNT {
		genre := gofakeit.BookGenre()

		if !slices.Contains(genres, genre) {
			genres = append(genres, genre)
		}
	}

	return genres
}
