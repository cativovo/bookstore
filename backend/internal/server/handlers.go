package server

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/cativovo/bookstore/internal/book"
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

	s.echo.GET("/health", h.healthCheck)
	s.echo.GET("/books", h.getBooks)
	s.echo.GET("/book/:id", h.getBookById)
	s.echo.GET("/genres", h.getGenres)
	s.echo.POST("/genre", h.createGenre)
	s.echo.DELETE("/genre/:name", h.deleteGenre)
	s.echo.POST("/book", h.createBook)
}

func (h *handler) healthCheck(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ok")
}

type getBooksQueryParam struct {
	// json tag is used by err.Field() of validator, the order of tags should always be query > json
	OrderBy string `query:"order_by" json:"order_by"`
	Desc    bool   `query:"desc" json:"desc"`
	Page    int    `query:"page" json:"page" validate:"gte=1"`
}

func (h *handler) getBooks(ctx echo.Context) error {
	var queryParam getBooksQueryParam
	if err := ctx.Bind(&queryParam); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ctx.Validate(&queryParam); err != nil {
		return err
	}

	const limit = 10

	books, count, err := h.bookService.GetBooks(
		book.GetBooksOptions{
			Limit:   limit,
			Offset:  (queryParam.Page - 1) * limit,
			OrderBy: queryParam.OrderBy,
			Desc:    queryParam.Desc,
		},
	)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}
	pages := math.Ceil(float64(count) / limit)

	return ctx.JSON(http.StatusOK, map[string]any{
		"books": books,
		"pages": pages,
	})
}

func (h *handler) getBookById(ctx echo.Context) error {
	id := ctx.Param("id")
	b, err := h.bookService.GetBookById(id)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.JSON(http.StatusOK, b)
}

func (h *handler) getGenres(ctx echo.Context) error {
	genres, err := h.bookService.GetGenres()
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.JSON(http.StatusOK, genres)
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
	name := ctx.Param("name")
	if err := h.bookService.DeleteGenre(name); err != nil {
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
