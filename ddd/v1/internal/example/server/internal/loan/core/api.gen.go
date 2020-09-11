// Code generated by golangee/architecture. DO NOT EDIT.

package core

import (
	uuid "github.com/golangee/uuid"
)

// Book is a book to loan or rent.
type Book struct {
	// ID is the unique id of a book.
	ID uuid.UUID `json:"id"`
	// ISBN the international number.
	ISBN int64 `json:"iSBN"`
	// LoanedBy is either nil or the user id.
	LoanedBy *uuid.UUID `json:"loanedBy"`
}

// User is a library customer.
type User struct {
	// ID is the unique id of the user.
	ID uuid.UUID `json:"id"`
}

// LoanService provides stuff to loan all the things.
type LoanService interface {
	// LoanIt loans a book.
	LoanIt()
}
