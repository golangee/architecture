// Code generated by golangee/architecture. DO NOT EDIT.

package rest

import (
	httprouter "github.com/julienschmidt/httprouter"
	log "log"
	http "net/http"
)

// BooksGetContext provides the specific http request and response context including already parsed parameters.
type BooksGetContext struct {
	// Request contains the raw http request.
	Request *http.Request
	// Writer contains a reference to the raw http response writer.
	Writer http.ResponseWriter
	// Session contains the parsed header parameter for 'session'.
	Session string
}

// BooksDeleteContext provides the specific http request and response context including already parsed parameters.
type BooksDeleteContext struct {
	// Request contains the raw http request.
	Request *http.Request
	// Writer contains a reference to the raw http response writer.
	Writer http.ResponseWriter
	// Session contains the parsed header parameter for 'session'.
	Session string
}

// Books represents the REST resource api/v1/books.
// Resource to manage books.
type Books interface {
	// GetBooks represents the http GET request on the /books resource.
	// Returns all books.
	GetBooks(ctx BooksGetContext) error
	// DeleteBooks represents the http DELETE request on the /books resource.
	// Removes all books.
	DeleteBooks(ctx BooksDeleteContext) error
}

// BooksMock is a mock implementation of Books.
// Books represents the REST resource api/v1/books.
// Resource to manage books.
type BooksMock struct {
	// GetBooksFunc mocks the GetBooks function.
	GetBooksFunc func(ctx BooksGetContext) error
	// DeleteBooksFunc mocks the DeleteBooks function.
	DeleteBooksFunc func(ctx BooksDeleteContext) error
}

// GetBooks represents the http GET request on the /books resource.
// Returns all books.
func (m BooksMock) GetBooks(ctx BooksGetContext) error {
	if m.GetBooksFunc != nil {
		return m.GetBooksFunc(ctx)
	}

	panic("mock not available: GetBooks")
}

// DeleteBooks represents the http DELETE request on the /books resource.
// Removes all books.
func (m BooksMock) DeleteBooks(ctx BooksDeleteContext) error {
	if m.DeleteBooksFunc != nil {
		return m.DeleteBooksFunc(ctx)
	}

	panic("mock not available: DeleteBooks")
}

// GetBooks returns the route to register on and the handler to execute.
// Currently, only the httprouter.Router is supported.
func GetBooks(api func(ctx BooksGetContext) error) (route string, handler http.HandlerFunc) {
	return "api/v1/books", func(w http.ResponseWriter, r *http.Request) {
		var err error
		ctx := BooksGetContext{
			Request: r,
			Writer:  w,
		}
		ctx.Session = r.Header.Get("session")
		if err = api(ctx); err != nil {
			log.Println(r.URL.String(), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

// DeleteBooks returns the route to register on and the handler to execute.
// Currently, only the httprouter.Router is supported.
func DeleteBooks(api func(ctx BooksDeleteContext) error) (route string, handler http.HandlerFunc) {
	return "api/v1/books", func(w http.ResponseWriter, r *http.Request) {
		var err error
		ctx := BooksDeleteContext{
			Request: r,
			Writer:  w,
		}
		ctx.Session = r.Header.Get("session")
		if err = api(ctx); err != nil {
			log.Println(r.URL.String(), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

// ConfigureBooks just applies the package wide endpoints into the given router without any other middleware.
func ConfigureBooks(api Books, router httprouter.Router) {
	router.GET(wrap(GetBooks(api.GetBooks)))
	router.DELETE(wrap(DeleteBooks(api.DeleteBooks)))
}
