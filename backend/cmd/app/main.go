package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cativovo/bookstore/pkg/book"
	"github.com/cativovo/bookstore/pkg/server"
	"github.com/cativovo/bookstore/pkg/storage/postgres"
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

	bookService := book.NewBookService(repository)

	s := server.NewServer(bookService)
	log.Fatal(s.ListenAndServe("127.0.0.1:5000"))
}
