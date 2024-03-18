package book

type Book struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Description string  `json:"description"`
	CoverImage  string  `json:"cover_image"`
	Genre       []Genre `json:"genre"`
	Price       float64 `json:"price"`
}
