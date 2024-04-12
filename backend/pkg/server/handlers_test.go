package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cativovo/bookstore/pkg/book"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) GetBooks(options book.GetBooksOptions) (books []book.Book, count int, err error) {
	return nil, 0, nil
}

func (m *MockBookRepository) GetBookById(id string) (book.Book, error) {
	return book.Book{}, nil
}

func (m *MockBookRepository) GetGenres() ([]string, error) {
	return nil, nil
}

func (m *MockBookRepository) CreateGenre(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockBookRepository) DeleteGenre(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockBookRepository) CreateBook(b book.Book) (book.Book, error) {
	args := m.Called(b)
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
		expectedErr        error
		serviceReturn      error
		payload            string
		expectedServiceArg string
		expectedStatusCode int
	}{
		{
			name:               "Success",
			expectedErr:        nil,
			serviceReturn:      nil,
			payload:            `{"name":"horror"}`,
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "Empty name",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'name' is required"),
			serviceReturn:      nil,
			payload:            `{"name":""}`,
			expectedServiceArg: "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Empty json",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'name' is required"),
			serviceReturn:      nil,
			payload:            "{}",
			expectedServiceArg: "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Internal server error",
			expectedErr:        echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
			serviceReturn:      errors.New("internal server error"),
			payload:            `{"name":"horror"}`,
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			mockRepository.On("CreateGenre", test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx, rec := newEchoContext(t, http.MethodPost, "/genre", strings.NewReader(test.payload))
			err := h.createGenre(ctx)
			assert.Equal(t, test.expectedErr, err)

			hErr, ok := err.(*echo.HTTPError)
			if ok {
				switch hErr.Code {
				case http.StatusBadRequest:
					return
				}
			}

			if err == nil {
				assert.Equal(t, "", rec.Body.String())
				assert.Equal(t, test.expectedStatusCode, rec.Code)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestDeleteGenre(t *testing.T) {
	tests := []struct {
		name               string
		expectedErr        error
		serviceReturn      error
		payload            string
		expectedServiceArg string
		expectedStatusCode int
	}{
		{
			name:               "Success",
			expectedErr:        nil,
			serviceReturn:      nil,
			payload:            "horror",
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "Not found",
			expectedErr:        echo.NewHTTPError(http.StatusNotFound, "genre not found"),
			serviceReturn:      book.ErrNotFound,
			payload:            "notfound",
			expectedServiceArg: "notfound",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "Internal server error",
			expectedErr:        echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
			serviceReturn:      errors.New("internal server error"),
			payload:            "horror",
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			mockRepository.On("DeleteGenre", test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx, rec := newEchoContext(t, http.MethodDelete, "/genre/:name", nil)
			ctx.SetParamNames("name")
			ctx.SetParamValues(test.payload)
			err := h.deleteGenre(ctx)
			assert.Equal(t, test.expectedErr, err)

			hErr, ok := err.(*echo.HTTPError)
			if ok {
				switch hErr.Code {
				case http.StatusNotFound:
					return
				}
			}

			if err == nil {
				assert.Equal(t, "", rec.Body.String())
				assert.Equal(t, test.expectedStatusCode, rec.Code)
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
		serviceReturn      book.Book
		serviceError       error
		expectedServiceArg book.Book
		expectedStatusCode int
		expectedOutput     string
		expectedErr        error
	}{
		{
			name:               "Success",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      successBook,
			serviceError:       nil,
			expectedServiceArg: successBook,
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     string(successBookJson),
			expectedErr:        nil,
		},
		{
			name:               "Empty genres",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":[],"price":69}`,
			serviceReturn:      emptyGenresBook,
			serviceError:       nil,
			expectedServiceArg: emptyGenresBook,
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     string(emptyGenresBookJson),
			expectedErr:        nil,
		},
		{
			name:               "Empty title",
			payload:            `{"title":"","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      book.Book{},
			serviceError:       nil,
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     "",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'title' is required"),
		},
		{
			name:               "Empty author",
			payload:            `{"title":"this is a title","author":"","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      book.Book{},
			serviceError:       nil,
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     "",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'author' is required"),
		},
		{
			name:               "No genres",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","price":69}`,
			serviceReturn:      book.Book{},
			serviceError:       nil,
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     "",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'genres' is required"),
		},
		{
			name:               "No price",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"]}`,
			serviceReturn:      book.Book{},
			serviceError:       nil,
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     "",
			expectedErr:        echo.NewHTTPError(http.StatusBadRequest, "'price' is required"),
		},
		{
			name:               "Internal server error",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      book.Book{},
			serviceError:       errors.New("internal server error"),
			expectedServiceArg: successBook,
			expectedStatusCode: http.StatusInternalServerError,
			expectedOutput:     "",
			expectedErr:        echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			test.expectedServiceArg.Id = ""
			mockRepository.On("CreateBook", test.expectedServiceArg).Return(test.serviceReturn, test.serviceError)
			h := handler{bookService: book.NewBookService(mockRepository)}
			ctx, rec := newEchoContext(t, http.MethodPost, "/book", strings.NewReader(test.payload))
			err := h.createBook(ctx)
			assert.Equal(t, test.expectedErr, err, "here")

			if err != nil {
				hErr, ok := err.(*echo.HTTPError)
				if ok {
					switch hErr.Code {
					case http.StatusBadRequest:
						mockRepository.AssertNotCalled(t, "CreateBook")
						return
					}

					assert.Equal(t, hErr.Code, test.expectedErr.(*echo.HTTPError).Code)
					assert.Equal(t, hErr.Error(), test.expectedErr.Error())
				}
			} else {
				// why use strings.TrimSpace - https://stackoverflow.com/questions/36319918/why-does-json-encoder-add-an-extra-line/36320146#36320146
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
				assert.Equal(t, test.expectedStatusCode, rec.Code)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

// func TestGetGenres(t *testing.T) {
// 	tests := []testold{
// 		{
// 			Code: http.StatusOK,
// 			Body: `["Horror","Comedy"]`,
// 			Cb: func() {
// 				memoryRepository.Genres = []string{"Horror", "Comedy"}
// 			},
// 		},
// 		{
// 			Code: http.StatusOK,
// 			Body: `[]`,
// 			Cb: func() {
// 				memoryRepository.Genres = []string{}
// 			},
// 		},
// 		{
// 			Err: echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
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
// 		ctx, rec := newEchoContext(http.MethodGet, "/genres", nil)
// 		err := hold.getGenres(ctx)
// 		test.assert(t, rec, err)
// 		memoryRepository.ReturnError = false
// 	}
//
// 	memoryRepository.Cleanup()
// }
//
// func TestGetBooks(t *testing.T) {
// 	memoryRepository.Seed()
// 	expectedBooks, err := json.Marshal(memoryRepository.Books[:10])
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	tests := []testold{
// 		{
// 			Code:    http.StatusOK,
// 			Body:    fmt.Sprintf(`{"books":%s,"pages":11}`, string(expectedBooks)),
// 			Payload: "?page=1",
// 		},
// 		{
// 			Err:     echo.NewHTTPError(http.StatusBadRequest, "'page' should be greater than or equal to 1"),
// 			Payload: "?page=0",
// 		},
// 		{
// 			Err:     echo.NewHTTPError(http.StatusInternalServerError, messageGenericError),
// 			Payload: "?page=1",
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
// 		ctx, rec := newEchoContext(http.MethodGet, "/books"+test.Payload, nil)
// 		err := hold.getBooks(ctx)
// 		test.assert(t, rec, err)
// 		memoryRepository.ReturnError = false
// 	}
//
// 	memoryRepository.Cleanup()
// }
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
