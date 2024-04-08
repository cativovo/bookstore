package server

import (
	"bytes"
	"io"
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

type test struct {
	Err     *echo.HTTPError
	Cb      func()
	Payload string
	Body    string
	Code    int
}

func (tt test) assert(t *testing.T, rec *httptest.ResponseRecorder, err error) {
	if tt.Err != nil && err == nil {
		t.Errorf("Expected: %v, Got: %v", tt.Err, err)
		return
	}

	if err != nil {
		// https://github.com/labstack/echo/issues/593#issuecomment-230926351
		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Error() != tt.Err.Error() {
				t.Errorf("Expected: %v, Got: %v", tt.Err.Error(), err.Error())
			}

			if he.Code != tt.Err.Code {
				t.Errorf("Expected: %v, Got: %v", tt.Code, he.Code)
			}
		} else {
			t.Fatalf("Invalid error %v", err)
		}
		return
	}

	if rec.Code != tt.Code {
		t.Errorf("Expected: %v, Got: %v", tt.Code, rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != tt.Body {
		t.Errorf("Expected: %v, Got: %v", tt.Body, body)
	}
}

func newEchoContext(method string, target string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Logger().SetOutput(bytes.NewBuffer([]byte{}))

	return ctx, rec
}

func TestCreateGenre(t *testing.T) {
	tests := []test{
		{
			Payload: `{"name": "horror"}`,
			Code:    http.StatusCreated,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "invalid json"),
			Payload: `[]`,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "genre 'horror' already exists"),
			Payload: `{"name": "horror"}`,
		},
		{
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'name' is required"),
			Payload: `{"name": ""}`,
		},
		{
			Err:     echo.NewHTTPError(http.StatusInternalServerError, "oops something went wrong"),
			Payload: `{"name": "error"}`,
			Cb: func() {
				memoryRepository.ReturnError = true
			},
		},
	}

	for _, test := range tests {
		if test.Cb != nil {
			test.Cb()
		}

		ctx, rec := newEchoContext(http.MethodPost, "/genre", strings.NewReader(test.Payload))
		err := h.createGenre(ctx)
		test.assert(t, rec, err)
		memoryRepository.ReturnError = false
	}

	expectedLen := 1
	gotLen := len(memoryRepository.Genres)
	if gotLen != expectedLen {
		t.Errorf("Expected: %v, Got: %v", expectedLen, gotLen)
	}

	memoryRepository.Cleanup()
}

func TestDeleteGenre(t *testing.T) {
	memoryRepository.Genres = []string{"Horror", "Comedy"}

	tests := []test{
		{
			Payload: "Horror",
			Code:    http.StatusNoContent,
		},
		{
			Err:     echo.NewHTTPError(http.StatusNotFound, "genre not found"),
			Payload: "Romance",
		},
		{
			Err:     echo.NewHTTPError(http.StatusInternalServerError, "oops something went wrong"),
			Payload: "error",
			Cb: func() {
				memoryRepository.ReturnError = true
			},
		},
	}

	memoryRepository.Genres = append(memoryRepository.Genres, "horror")

	for _, test := range tests {
		if test.Cb != nil {
			test.Cb()
		}

		ctx, rec := newEchoContext(http.MethodDelete, "/", nil)
		ctx.SetPath("/genre/:name")
		ctx.SetParamNames("name")
		ctx.SetParamValues(test.Payload)
		err := h.deleteGenre(ctx)
		test.assert(t, rec, err)
		memoryRepository.ReturnError = false
	}

	memoryRepository.Cleanup()
}

func TestCreateBook(t *testing.T) {
	tests := []test{
		{
			Payload: `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			Code:    http.StatusCreated,
			Body:    `{"id":"1","title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
		},
		{
			Payload: `{"title":"","author":"john doe","description":"","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'title' is required"),
		},
		{
			Payload: `{"title":"this is a title","author":"","description":"","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'author' is required"),
		},
		{
			Payload: `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"]}`,
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'price' is required"),
		},
		{
			Payload: `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","price":69}`,
			Err:     echo.NewHTTPError(http.StatusBadRequest, "'genres' is required"),
		},
		{
			Payload: `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			Err:     echo.NewHTTPError(http.StatusInternalServerError, "oops something went wrong"),
			Cb: func() {
				memoryRepository.ReturnError = true
			},
		},
	}

	for _, test := range tests {
		if test.Cb != nil {
			test.Cb()
		}

		ctx, rec := newEchoContext(http.MethodPost, "/book", strings.NewReader(test.Payload))
		err := h.createBook(ctx)
		test.assert(t, rec, err)
		memoryRepository.ReturnError = false
	}

	memoryRepository.Cleanup()
}

func TestGetGenres(t *testing.T) {
	tests := []test{
		{
			Code: http.StatusOK,
			Body: `["Horror","Comedy"]`,
			Cb: func() {
				memoryRepository.Genres = []string{"Horror", "Comedy"}
			},
		},
		{
			Code: http.StatusOK,
			Body: `[]`,
			Cb: func() {
				memoryRepository.Genres = []string{}
			},
		},
		{
			Err: echo.NewHTTPError(http.StatusInternalServerError, "oops something went wrong"),
			Cb: func() {
				memoryRepository.ReturnError = true
			},
		},
	}

	for _, test := range tests {
		if test.Cb != nil {
			test.Cb()
		}

		ctx, rec := newEchoContext(http.MethodGet, "/genres", nil)
		err := h.getGenres(ctx)
		test.assert(t, rec, err)
		memoryRepository.ReturnError = false
	}

	memoryRepository.Cleanup()
}
