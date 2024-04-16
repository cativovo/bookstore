package server

import (
	"bytes"
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

func (m *MockBookRepository) GetBooks(options book.GetBooksOptions) (books []book.Book, count int, err error) {
	return nil, 0, nil
}

func (m *MockBookRepository) GetBookById(id string) (book.Book, error) {
	return book.Book{}, nil
}

func (m *MockBookRepository) GetGenres() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
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
			name:               "Internal server error",
			serviceReturn:      errors.New("internal server error"),
			payload:            `{"name":"horror"}`,
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusInternalServerError,
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, messageInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			mockRepository.On("CreateGenre", test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx, rec := newEchoContext(t, http.MethodPost, "/genre", strings.NewReader(test.payload))
			err := h.createGenre(ctx)

			if err != nil {
				expectedOutput, expectedOutputOk := test.expectedOutput.(*echo.HTTPError)
				if !expectedOutputOk {
					assert.Fail(t, "Unexpected err", err)
					return
				}

				hErr, hErrOk := err.(*echo.HTTPError)
				if hErrOk {
					switch hErr.Code {
					case http.StatusBadRequest:
						return
					}

					assert.Equal(t, expectedOutput.Code, hErr.Code)
					assert.Equal(t, expectedOutput.Error(), hErr.Error())
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
			expectedStatusCode: http.StatusNotFound,
			expectedOutput:     echo.NewHTTPError(http.StatusNotFound, "genre not found"),
		},
		{
			name:               "Internal server error",
			serviceReturn:      errors.New("internal server error"),
			genre:              "horror",
			expectedServiceArg: "horror",
			expectedStatusCode: http.StatusInternalServerError,
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, messageInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			mockRepository.On("DeleteGenre", test.expectedServiceArg).Return(test.serviceReturn)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx, rec := newEchoContext(t, http.MethodDelete, "/genre/:name", nil)
			ctx.SetParamNames("name")
			ctx.SetParamValues(test.genre)
			err := h.deleteGenre(ctx)

			if err != nil {
				expectedOutput, expectedOutputOk := test.expectedOutput.(*echo.HTTPError)
				if !expectedOutputOk {
					assert.Fail(t, "Unexpected err", err)
					return
				}

				hErr, hErrOk := err.(*echo.HTTPError)
				if hErrOk {
					assert.Equal(t, expectedOutput.Code, hErr.Code)
					assert.Equal(t, expectedOutput.Error(), hErr.Error())
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
			name:               "Empty title",
			payload:            `{"title":"","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      []any{book.Book{}, nil},
			expectedServiceArg: book.Book{},
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'title' is required"),
		},
		{
			name:               "Empty author",
			payload:            `{"title":"this is a title","author":"","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      []any{book.Book{}, nil},
			expectedServiceArg: book.Book{},
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'author' is required"),
		},
		{
			name:               "No genres",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","price":69}`,
			serviceReturn:      []any{book.Book{}, nil},
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'genres' is required"),
		},
		{
			name:               "No price",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"]}`,
			serviceReturn:      []any{book.Book{}, nil},
			expectedServiceArg: book.Book{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     echo.NewHTTPError(http.StatusBadRequest, "'price' is required"),
		},
		{
			name:               "Internal server error",
			payload:            `{"title":"this is a title","author":"john doe","description":"this is a description","cover_image":"coverimage.com","genres":["horror"],"price":69}`,
			serviceReturn:      []any{book.Book{}, errors.New("internal server error")},
			expectedServiceArg: successBook,
			expectedStatusCode: http.StatusInternalServerError,
			expectedOutput:     echo.NewHTTPError(http.StatusInternalServerError, messageInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			test.expectedServiceArg.Id = ""
			mockRepository.On("CreateBook", test.expectedServiceArg).Return(test.serviceReturn...)
			h := handler{bookService: book.NewBookService(mockRepository)}
			ctx, rec := newEchoContext(t, http.MethodPost, "/book", strings.NewReader(test.payload))
			err := h.createBook(ctx)

			if err != nil {
				expectedOutput, expectedOutputOk := test.expectedOutput.(*echo.HTTPError)
				if !expectedOutputOk {
					assert.Fail(t, "Unexpected err", err)
					return
				}

				hErr, hErrOk := err.(*echo.HTTPError)
				if hErrOk {
					switch hErr.Code {
					case http.StatusBadRequest:
						return
					}

					assert.Equal(t, expectedOutput.Code, hErr.Code)
					assert.Equal(t, expectedOutput.Error(), hErr.Error())
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
			expectedOutput: echo.NewHTTPError(http.StatusInternalServerError, messageInternalServerError), expectedStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository := new(MockBookRepository)
			mockRepository.On("GetGenres").Return(test.serviceReturn...)
			h := handler{bookService: book.NewBookService(mockRepository)}

			ctx, rec := newEchoContext(t, http.MethodGet, "/genres", nil)
			err := h.getGenres(ctx)

			if err != nil {
				expectedOutput, expectedOutputOk := test.expectedOutput.(*echo.HTTPError)
				if !expectedOutputOk {
					assert.Fail(t, "Unexpected err", err)
					return
				}

				hErr, hErrOk := err.(*echo.HTTPError)
				if hErrOk {
					assert.Equal(t, expectedOutput.Code, hErr.Code)
					assert.Equal(t, expectedOutput.Error(), hErr.Error())
				}
			} else {
				assert.Equal(t, test.expectedStatusCode, rec.Code)
				assert.Equal(t, test.expectedOutput, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}

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
