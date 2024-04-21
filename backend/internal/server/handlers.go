package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"

	"github.com/cativovo/bookstore/internal/book"
	"github.com/labstack/echo/v4"
)

type handler struct {
	bookService *book.BookService
}

const (
	msgInternalServerErr = "oops something went wrong"
	msgInvalidPayload    = "unable to parse the request"
)

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
	OrderBy string `query:"order_by"`
	Author  string `query:"author"`
	Genres  string `query:"genres"`
	Title   string `query:"title"`
	Page    int    `query:"page"`
	Desc    bool   `query:"desc"`
}

func (h *handler) getBooks(ctx echo.Context) error {
	var queryParam getBooksQueryParam

	err := echo.QueryParamsBinder(ctx).
		Int("page", &queryParam.Page).
		Bool("desc", &queryParam.Desc).
		String("order_by", &queryParam.OrderBy).
		String("author", &queryParam.Author).
		String("genres", &queryParam.Genres).
		String("title", &queryParam.Title).
		BindError()
	if err != nil {
		bindingErr := err.(*echo.BindingError)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid value for '%s'", bindingErr.Field))
	}

	if queryParam.Page <= 0 {
		queryParam.Page = 1
	}

	const limit = 10

	genres := strings.Split(queryParam.Genres, ",")
	for i, genre := range genres {
		genres[i] = strings.TrimSpace(genre)
	}

	books, count, err := h.bookService.GetBooks(
		ctx.Request().Context(),
		book.GetBooksOptions{
			Limit:   limit,
			Offset:  (queryParam.Page - 1) * limit,
			OrderBy: queryParam.OrderBy,
			Desc:    queryParam.Desc,
			Filter: book.GetBooksFilter{
				Author: queryParam.Author,
				Title:  queryParam.Title,
				// Genres: genres,
			},
		},
	)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}
	pages := math.Ceil(float64(count) / limit)

	return ctx.JSON(http.StatusOK, map[string]any{
		"books": books,
		"pages": pages,
	})
}

func (h *handler) getBookById(ctx echo.Context) error {
	id := ctx.Param("id")
	b, err := h.bookService.GetBookById(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}

	return ctx.JSON(http.StatusOK, b)
}

func (h *handler) getGenres(ctx echo.Context) error {
	genres, err := h.bookService.GetGenres(ctx.Request().Context())
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}

	return ctx.JSON(http.StatusOK, genres)
}

type payloadCreateGenre struct {
	Name string `json:"name" validate:"required"`
}

func (h *handler) createGenre(ctx echo.Context) error {
	var payload payloadCreateGenre
	if err := ctx.Bind(&payload); err != nil {
		return getBindErr(err)
	}

	if err := ctx.Validate(&payload); err != nil {
		return err
	}

	err := h.bookService.CreateGenre(ctx.Request().Context(), payload.Name)
	if err != nil {
		if errors.Is(err, book.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("genre '%s' already exists", payload.Name))
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (h *handler) deleteGenre(ctx echo.Context) error {
	name := ctx.Param("name")
	if err := h.bookService.DeleteGenre(ctx.Request().Context(), name); err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "genre not found")
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}

	return ctx.NoContent(http.StatusNoContent)
}

type payloadCreateBook struct {
	// https://github.com/go-playground/validator/issues/692#issuecomment-737039536
	Price       *float64 `json:"price" validate:"required,gt=0"`
	Title       string   `json:"title" validate:"required"`
	Author      string   `json:"author" validate:"required"`
	Description string   `json:"description"`
	CoverImage  string   `json:"cover_image"`
	Genres      []string `json:"genres" validate:"required"`
}

func (h *handler) createBook(ctx echo.Context) error {
	var payload payloadCreateBook
	if err := ctx.Bind(&payload); err != nil {
		return getBindErr(err)
	}

	if err := ctx.Validate(&payload); err != nil {
		return err
	}

	b, err := h.bookService.CreateBook(ctx.Request().Context(), book.Book{
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
		return echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr)
	}

	return ctx.JSON(http.StatusCreated, b)
}

func getBindErr(err error) *echo.HTTPError {
	defaultStatusCode := http.StatusBadRequest

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if ute, ok := httpErr.Internal.(*json.UnmarshalTypeError); ok && ute.Type.Kind() != reflect.Struct {
			return echo.NewHTTPError(defaultStatusCode, fmt.Sprintf("'%s' should be %s", ute.Field, ute.Type))
		}
	}

	return echo.NewHTTPError(defaultStatusCode, msgInvalidPayload)
}
