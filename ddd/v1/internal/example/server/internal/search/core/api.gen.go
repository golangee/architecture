// Code generated by golangee/architecture. DO NOT EDIT.

package core

import (
	context "context"
	uuid "github.com/golangee/uuid"
	uuid1 "github.com/google/uuid"
)

// Book is a book with meta data to index and find.
type Book struct {
	// ID is the unique id of a book.
	ID uuid.UUID
	// Title is the title for the book.
	Title string
	// Special is a test for importing a custom type.
	Special uuid1.UUID
	// Tags to describe a book.
	Tags []string
}

// BookRepository is a repository to handle books.
type BookRepository interface {
	// ReadAll returns all books.
	//
	// The parameter 'ctx' is the context to control timeouts and cancellations.
	//
	// The result '[]Book' is the list of books.
	//
	// The result 'error' indicates a violation of pre- or invariants and represents an implementation specific failure.
	ReadAll(ctx context.Context) ([]Book, error)
}

// SearchService is the domain specific service API.
type SearchService interface {
	// Search inspects each book for the key words.
	//
	// The parameter 'ctx' is the context to control timeouts and cancellations.
	//
	// The parameter 'query' contains the query to search for.
	//
	// The result '[]Book' is the list of found books.
	//
	// The result 'error' indicates a violation of pre- or invariants and represents an implementation specific failure.
	Search(ctx context.Context, query string) ([]Book, error)
}
