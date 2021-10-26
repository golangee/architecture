package ticket

import uuid "github.com/google/uuid"

type TicketBase struct {
	id          uuid.UUID
	title       string
	description string
	// The User that created this ticket.
	creator uuid.UUID
	// The User that should work on this ticket.
	assignee uuid.UUID
}
type TicketCreate struct{}
type TicketShow struct{}
type TicketById struct{}
type SpecialThing struct {
	timestamp int
	user      User
}
type User struct {
	id   uuid.UUID
	name string
}
type UserById struct{}
