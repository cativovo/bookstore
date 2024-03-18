package book

type BookRepository interface {
	// GetBooks() ([]Book, error)
	CreateGenre(name string) (Genre, error)
}

type BookService struct {
	repository BookRepository
}

func NewBookService(r BookRepository) *BookService {
	return &BookService{
		repository: r,
	}
}

// func (bs *BookService) GetBooks() ([]Book, error) {
// 	return bs.repository.GetBooks()
// }

func (bs *BookService) CreateGenre(name string) (Genre, error) {
	return bs.repository.CreateGenre(name)
}
