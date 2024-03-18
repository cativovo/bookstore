package server

import (
	"errors"
	"net/http"

	"github.com/cativovo/bookstore/pkg/book"
	"github.com/labstack/echo/v4"
)

type controller struct {
	bookService *book.BookService
}

const messageGenericError = "oops something went wrong"

func (s *Server) registerControllers() {
	c := controller{
		bookService: s.bookService,
	}

	s.echo.GET("/books", c.getBooks)
	s.echo.POST("/genre", c.createGenre)
	s.echo.DELETE("/genre/:id", c.deleteGenre)
}

func (c *controller) getBooks(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"test": "test",
	})
}

type payloadCreateGenre struct {
	Name string `json:"name" validate:"required"`
}

func (c *controller) createGenre(ctx echo.Context) error {
	var payload payloadCreateGenre
	if err := ctx.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ctx.Validate(&payload); err != nil {
		return err
	}

	genre, err := c.bookService.CreateGenre(payload.Name)
	if err != nil {
		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.JSON(http.StatusOK, genre)
}

func (c *controller) deleteGenre(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := c.bookService.DeleteGenre(id); err != nil {
		if errors.Is(err, book.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "genre not found")
		}

		ctx.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, messageGenericError)
	}

	return ctx.NoContent(http.StatusNoContent)
}
