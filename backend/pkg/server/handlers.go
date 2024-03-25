package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cativovo/bookstore/pkg/book"
	"github.com/labstack/echo/v4"
)

type handler struct {
	bookService *book.BookService
}

const messageGenericError = "oops something went wrong"

func (s *Server) registerHandlers() {
	h := handler{
		bookService: s.bookService,
	}

	s.echo.GET("/books", h.getBooks)
	s.echo.POST("/genre", h.createGenre)
	s.echo.DELETE("/genre/:id", h.deleteGenre)
	s.echo.POST("/book", h.createBook)
}

func (h *handler) getBooks(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"test": "test",
	})
}

type payloadCreateGenre struct {
	Name string `json:"name" validate:"required"`
}

func (h *handler) createGenre(ctx echo.Context) error {
	var payload payloadCreateGenre
	if err := ctx.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	if err := ctx.Validate(&payload); err != nil {
		return err
	}

	err := h.bookService.CreateGenre(payload.Name)
	if err != nil {
		if errors.Is(err, book.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("genre '%s' already exists", payload.Name))
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (h *handler) deleteGenre(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := h.bookService.DeleteGenre(id); err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "genre not found")
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

type payloadCreateBook struct {
	// https://github.com/go-playground/validator/issues/692#issuecomment-737039536
	Price       *float64 `json:"price" validate:"required,number"`
	Title       string   `json:"title" validate:"required"`
	Author      string   `json:"author" validate:"required"`
	Description string   `json:"description"`
	CoverImage  string   `json:"cover_image"`
	Genres      []string `json:"genres" validate:"required"`
}

func (h *handler) createBook(ctx echo.Context) error {
	var payload payloadCreateBook
	if err := ctx.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ctx.Validate(&payload); err != nil {
		return err
	}

	b, err := h.bookService.CreateBook(book.Book{
		Title:       payload.Title,
		Author:      payload.Author,
		Description: payload.Description,
		CoverImage:  payload.CoverImage,
		Price:       *payload.Price,
		Genres:      payload.Genres,
	})
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid genre")
		}
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.JSON(http.StatusCreated, b)
}
