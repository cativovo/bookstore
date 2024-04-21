package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/cativovo/bookstore/internal/book"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) GetBooks(ctx context.Context, options book.GetBooksOptions) (books []book.Book, count int, err error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]book.Book), args.Int(1), args.Error(2)
}

func (m *MockBookRepository) GetBookById(ctx context.Context, id string) (book.Book, error) {
	return book.Book{}, nil
}

func (m *MockBookRepository) GetGenres(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockBookRepository) CreateGenre(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockBookRepository) DeleteGenre(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockBookRepository) CreateBook(ctx context.Context, b book.Book) (book.Book, error) {
	args := m.Called(ctx, b)
	return args.Get(0).(book.Book), args.Error(1)
}

var e = echo.New()

func TestMain(m *testing.M) {
	e.Validator = NewValidator()
	m.Run()
}

func newEchoContext(t *testing.T, method string, target string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Logger().SetOutput(bytes.NewBuffer([]byte{}))

	return ctx, rec
}

func TestCreateGenre(t *testing.T) {
	tests := []struct {
		name               string
		expectedOutput     any
		serviceReturn      error
		payload            string
		expectedServiceArg string
		expectedStatusCode int
	}{
		{
			name:               "Success",
			payload:            `{"name":"horror"}`,
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     "",
		},
		{
			name:               "Empty name",
			payload:            `{"name":""}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'name' is required"),
		},
		{
			name:               "Empty json",
			payload:            "{}",
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'name' is required"),
		},
		{
			name:               "Invalid json",
			payload:            "[]",
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, msgInvalidPayload),
		},
		{
			name:               "Internal server error",
			serviceReturn:      errors.New("internal server error"),
			payload:            `{"name":"horror"}`,
			expectedServiceArg: "horror",
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, rec := newEchoContext(t, http.MethodPost, "/genre", strings.NewReader(test.payload))

			mockRepository := new(MockBookRepository)
			mockRepository.On("CreateGenre", ctx.Request().Context(), test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			err := h.createGenre(ctx)

			if err != nil {
				if !assert.Equal(t, test.expectedOutput, err) {
					return
				}

				switch test.expectedOutput.(*echo.HTTPError).Code {
				case http.StatusBadRequest:
					mockRepository.AssertNotCalled(t, "CreateGenre", test.expectedServiceArg)
					return
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				// why use strings.TrimSpace - https://stackoverflow.com/questions/36319918/why-does-json-encoder-add-an-extra-line/36320146#36320146
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestDeleteGenre(t *testing.T) {
	tests := []struct {
		name               string
		expectedOutput     any
		serviceReturn      error
		genre              string
		expectedServiceArg string
		expectedStatusCode int
	}{
		{
			name:               "Success",
			genre:              "horror",
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusNoContent,
			expectedOutput:     "",
		},
		{
			name:               "Not found",
			serviceReturn:      book.ErrNotFound,
			genre:              "notfound",
			expectedServiceArg: "notfound",
			expectedOutput:     echo.NewHTTPError(http.StatusNotFound, "genre not found"),
		},
		{
			name:               "Internal server error",
			serviceReturn:      errors.New("internal server error"),
			genre:              "horror",
			expectedServiceArg: "horror",
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, rec := newEchoContext(t, http.MethodDelete, "/genre/:name", nil)

			mockRepository := new(MockBookRepository)
			mockRepository.On("DeleteGenre", ctx.Request().Context(), test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx.SetParamNames("name")
			ctx.SetParamValues(test.genre)
			err := h.deleteGenre(ctx)

			if err != nil {
				if !assert.Equal(t, test.expectedOutput, err) {
					return
				}

				switch test.expectedOutput.(*echo.HTTPError).Code {
				case http.StatusBadRequest:
					mockRepository.AssertNotCalled(t, "DeleteGenre", test.expectedServiceArg)
					return
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				// why use strings.TrimSpace - https://stackoverflow.com/questions/36319918/why-does-json-encoder-add-an-extra-line/36320146#36320146
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestCreateBook(t *testing.T) {
	successBook := book.Book{
		Id:          "1234",
		Title:       "this is a title",
		Author:      "john doe",
		Description: "this is a description",
		CoverImage:  "coverimage.com",
		Genres:      []string{"horror"},
		Price:       69.0,
	}

	successBookJson, err := json.Marshal(successBook)
	if err != nil {
		t.Fatal(err)
	}

	emptyGenresBook := successBook
	emptyGenresBook.Genres = []string{}
	emptyGenresBookJson, err := json.Marshal(emptyGenresBook)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name               string
		payload            string
		expectedOutput     any
		serviceReturn      []any
		expectedServiceArg book.Book
		expectedStatusCode int
	}{
		{
			name:               "Success",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      []any{successBook, nil},
			expectedServiceArg: successBook,
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     string(successBookJson),
		},
		{
			name:               "Empty genres",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":[],"price":69}`,
			serviceReturn:      []any{emptyGenresBook, nil},
			expectedServiceArg: emptyGenresBook,
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     string(emptyGenresBookJson),
		},
		{
			name:           "Empty title",
			payload:        `{"title":"","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'title' is required"),
		},
		{
			name:           "Empty author",
			payload:        `{"title":"this is a title","author":"","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'author' is required"),
		},
		{
			name:           "No genres",
			payload:        `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","price":69}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'genres' is required"),
		},
		{
			name:           "No price",
			payload:        `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"]}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'price' is required"),
		},
		{
			name:           "Zero price",
			payload:        `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"], "price": 0}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'price' should be greater than 0"),
		},
		{
			name:           "Invalid type",
			payload:        `{"title":69,"author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"], "price": 69}`,
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "'title' should be string"),
		},
		{
			name:           "Invalid json",
			payload:        "[]",
			serviceReturn:  []any{},
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, msgInvalidPayload),
		},
		{
			name:               "Internal server error",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      []any{book.Book{}, errors.New("internal server error")},
			expectedServiceArg: successBook,
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, rec := newEchoContext(t, http.MethodPost, "/book", strings.NewReader(test.payload))

			test.expectedServiceArg.Id = ""
			mockRepository := new(MockBookRepository)
			mockRepository.On("CreateBook", ctx.Request().Context(), test.expectedServiceArg).Return(test.serviceReturn...)
			h := handler{bookService: book.NewBookService(mockRepository)}

			err := h.createBook(ctx)

			if err != nil {
				if !assert.Equal(t, test.expectedOutput, err) {
					return
				}

				switch test.expectedOutput.(*echo.HTTPError).Code {
				case http.StatusBadRequest:
					mockRepository.AssertNotCalled(t, "CreateBook", test.expectedServiceArg)
					return
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				// why use strings.TrimSpace - https://stackoverflow.com/questions/36319918/why-does-json-encoder-add-an-extra-line/36320146#36320146
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestGetGenres(t *testing.T) {
	data, err := os.ReadFile("../../testdata/genres.json")
	if err != nil {
		t.Fatal(err)
	}
	var testdata []string

	if err := json.Unmarshal(data, &testdata); err != nil {
		t.Fatal(err)
	}

	testdataBytes, err := json.Marshal(testdata)
	if err != nil {
		t.Fatal(err)
	}

	testdataJson := string(testdataBytes)

	tests := []struct {
		name               string
		expectedOutput     any
		serviceReturn      []any
		expectedStatusCode int
	}{
		{
			name:               "Success",
			serviceReturn:      []any{testdata, nil},
			expectedOutput:     testdataJson,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:           "Internal server error",
			serviceReturn:  []any{testdata, errors.New("internal server error")},
			expectedOutput: echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr), expectedStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, rec := newEchoContext(t, http.MethodGet, "/genres", nil)

			mockRepository := new(MockBookRepository)
			mockRepository.On("GetGenres", ctx.Request().Context()).Return(test.serviceReturn...)
			h := handler{bookService: book.NewBookService(mockRepository)}

			err := h.getGenres(ctx)

			if err != nil {
				if !assert.Equal(t, test.expectedOutput, err) {
					return
				}

				switch test.expectedOutput.(*echo.HTTPError).Code {
				case http.StatusBadRequest:
					mockRepository.AssertNotCalled(t, "GetGenres")
					return
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestGetBooks(t *testing.T) {
	data, err := os.ReadFile("../../testdata/books.json")
	if err != nil {
		t.Fatal(err)
	}

	var testdata []book.Book

	if err := json.Unmarshal(data, &testdata); err != nil {
		t.Fatal(err)
	}

	type response struct {
		Books []book.Book `json:"books"`
		Pages int         `json:"pages"`
	}

	success := response{
		Books: testdata,
		Pages: 11,
	}

	successBytes, err := json.Marshal(success)
	if err != nil {
		t.Fatal(err)
	}

	successEmptyBooks := response{
		Pages: 11,
	}

	successEmptyBooksBytes, err := json.Marshal(successEmptyBooks)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		expectedOutput     any
		name               string
		query              string
		serviceReturn      []any
		expectedServiceArg book.GetBooksOptions
		expectedStatusCode int
	}{
		{
			name:          "Success without query",
			serviceReturn: []any{success.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit: 10,
			},
			expectedOutput:     string(successBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success page",
			query:         "?page=1",
			serviceReturn: []any{success.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit:  10,
				Offset: 0,
			},
			expectedOutput:     string(successBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success exceed pages",
			query:         "?page=6969",
			serviceReturn: []any{successEmptyBooks.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit:  10,
				Offset: 6968 * 10,
			},
			expectedOutput:     string(successEmptyBooksBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success desc",
			query:         "?desc=true",
			serviceReturn: []any{successEmptyBooks.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit: 10,
				Desc:  true,
			},
			expectedOutput:     string(successEmptyBooksBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success order_by",
			query:         "?order_by=author",
			serviceReturn: []any{successEmptyBooks.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit:   10,
				OrderBy: "author",
			},
			expectedOutput:     string(successEmptyBooksBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:          "Success filter",
			query:         "?title=title",
			serviceReturn: []any{successEmptyBooks.Books, 101, nil},
			expectedServiceArg: book.GetBooksOptions{
				Limit: 10,
				Filter: book.GetBooksFilter{
					Title: "title",
				},
			},
			expectedOutput:     string(successEmptyBooksBytes),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:           "Invalid page",
			query:          "?page=j",
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "invalid value for 'page'"),
		},
		{
			name:           "Invalid desc",
			query:          "?desc=j",
			expectedOutput: echo.NewHTTPError(http.StatusBadRequest, "invalid value for 'desc'"),
		},
		{
			name:          "Internal server error",
			serviceReturn: []any{success.Books, 101, errors.New("internal server error")},
			expectedServiceArg: book.GetBooksOptions{
				Limit: 10,
			},
			expectedOutput: echo.NewHTTPError(http.StatusInternalServerError, msgInternalServerErr),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, rec := newEchoContext(t, http.MethodGet, "/books"+test.query, nil)

			mockRepository := new(MockBookRepository)
			mockRepository.On("GetBooks", ctx.Request().Context(), test.expectedServiceArg).Return(test.serviceReturn...)
			h := handler{bookService: book.NewBookService(mockRepository)}

			err := h.getBooks(ctx)

			if err != nil {
				if !assert.Equal(t, test.expectedOutput, err) {
					return
				}

				switch test.expectedOutput.(*echo.HTTPError).Code {
				case http.StatusBadRequest:
					mockRepository.AssertNotCalled(t, "GetBooks", test.expectedServiceArg)
					return
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

//
// func TestGetBookById(t *testing.T) {
// 	memoryRepository.Seed()
// 	expectedBook := memoryRepository.Books[rand.IntN(len(memoryRepository.Books))]
// 	expectedBookBytes, err := json.Marshal(expectedBook)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	tests := []testold{
// 		{
// 			Payload: expectedBook.Id,
// 			Code:    http.StatusOK,
// 			Body:    string(expectedBookBytes),
// 		},
// 		{
// 			Payload: "404",
// 			Err:     echo.NewHTTPError(http.StatusNotFound, "book not found"),
// 		},
// 		{
// 			Payload: expectedBook.Id,
// 			Err:     echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
// 			Cb: func() {
// 				memoryRepository.ReturnError = true
// 			},
// 		},
// 	}
//
// 	for _, test := range tests {
// 		if test.Cb != nil {
// 			test.Cb()
// 		}
//
// 		ctx, rec := newEchoContext(http.MethodGet, "/book/:id", nil)
// 		ctx.SetParamNames("id")
// 		ctx.SetParamValues(test.Payload)
// 		err := hold.getBookById(ctx)
// 		test.assert(t, rec, err)
// 		memoryRepository.ReturnError = false
// 	}
//
// 	memoryRepository.Cleanup()
// }
