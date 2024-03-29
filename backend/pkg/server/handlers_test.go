package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cativovo/bookstore/pkg/book"
	"github.com/cativovo/bookstore/pkg/storage/memory"
	"github.com/labstack/echo/v4"
)

var (
	e                = echo.New()
	h                handler
	memoryRepository = memory.NewMemoryRepository()
)

func TestMain(m *testing.M) {
	h = handler{bookService: book.NewBookService(memoryRepository)}
	e.Validator = NewValidator()
	m.Run()
}

type expected struct {
	Err     error
	Payload string
	Body    string
	Code    int
}

func (e expected) test(t *testing.T, rec *httptest.ResponseRecorder, err error) {
	if err != nil {
		// https://github.com/labstack/echo/issues/593#issuecomment-230926351
		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Error() != e.Err.Error() {
				t.Errorf("Expected: %v, Got: %v", e.Err.Error(), err.Error())
			}

			if he.Code != e.Code {
				t.Errorf("Expected: %v, Got: %v", e.Code, he.Code)
			}
		} else {
			t.Fatalf("Invalid error %v", err)
		}
	} else {
		if rec.Code != e.Code {
			t.Errorf("Expected: %v, Got: %v", e.Code, rec.Code)
		}

		body := rec.Body.String()
		if body != e.Body {
			t.Errorf("Expected: %v, Got: %v", e.Body, body)
		}
	}
}

func TestCreateGenre(t *testing.T) {
	tests := []expected{
		{
			Payload: `{"name": "horror"}`,
			Code:    http.StatusCreated,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "invalid json"),
			Payload: `[]`,
			Code:    http.StatusBadRequest,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "genre 'horror' already exists"),
			Payload: `{"name": "horror"}`,
			Code:    http.StatusBadRequest,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'name' cannot be blank"),
			Payload: `{"name": ""}`,
			Code:    http.StatusBadRequest,
		},
	}

	for _, expected := range tests {
		req := httptest.NewRequest(http.MethodPost, "/genre", strings.NewReader(expected.Payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := h.createGenre(ctx)
		expected.test(t, rec, err)
	}

	expectedLen := 1
	gotLen := len(memoryRepository.Genres)
	if gotLen != expectedLen {
		t.Errorf("Expected: %v, Got: %v", expectedLen, gotLen)
	}

	memoryRepository.Cleanup()
}

func TestDeleteGenre(t *testing.T) {
	tests := []expected{
		{
			Payload: "horror",
			Code:    http.StatusNoContent,
		},
		{
			Err:     echo.NewHTTPError(http.StatusNotFound, "genre not found"),
			Payload: "notfound",
			Code:    http.StatusNotFound,
		},
	}

	memoryRepository.Genres = append(memoryRepository.Genres, "horror")

	for _, expected := range tests {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/genre/:name")
		ctx.SetParamNames("name")
		ctx.SetParamValues(expected.Payload)
		err := h.deleteGenre(ctx)
		expected.test(t, rec, err)
	}

	memoryRepository.Cleanup()
}
