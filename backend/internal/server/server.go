package server

import (
	"github.com/cativovo/bookstore/internal/book"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo        *echo.Echo
	bookService *book.BookService
}

func NewServer(bs *book.BookService) *Server {
	e := echo.New()
	e.Validator = NewValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s := &Server{
		echo:        e,
		bookService: bs,
	}

	s.registerHandlers()

	return s
}

func (s *Server) ListenAndServe(addr string) error {
	return s.echo.Start(addr)
}
