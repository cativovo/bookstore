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
	e = echo.New()
	h = handler{bookService: book.NewBookService(memory.NewMemoryRepository())}
)

func TestMain(m *testing.M) {
	e.Validator = NewValidator()
	m.Run()
}

type test struct {
	Err     error
	Payload string
	Body    string
	Code    int
}

func TestCreateGenre(t *testing.T) {
	tests := []test{
		{
			Err:     nil,
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
		if err != nil {
			// https://github.com/labstack/echo/issues/593#issuecomment-230926351
			he, ok := err.(*echo.HTTPError)
			if ok {
				if he.Error() != expected.Err.Error() {
					t.Errorf("Expected: %v, Got: %v", expected.Err.Error(), err.Error())
				}

				if he.Code != expected.Code {
					t.Errorf("Expected: %v, Got: %v", expected.Code, he.Code)
				}
			} else {
				t.Fatalf("Invalid error %v", err)
			}
		} else {
			if rec.Code != expected.Code {
				t.Errorf("Expected: %v, Got: %v", expected.Code, rec.Code)
			}

			body := rec.Body.String()
			if body != expected.Body {
				t.Errorf("Expected: %v, Got: %v", expected.Body, body)
			}
		}

	}
}
