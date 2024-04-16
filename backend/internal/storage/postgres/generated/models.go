// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package query

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Book struct {
	ID          pgtype.UUID
	Title       string
	Author      string
	Description pgtype.Text
	CoverImage  pgtype.Text
	Price       pgtype.Numeric
}

type BookGenre struct {
	ID      pgtype.UUID
	BookID  pgtype.UUID
	GenreID pgtype.UUID
}

type Genre struct {
	ID   pgtype.UUID
	Name pgtype.Text
}