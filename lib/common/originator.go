package common

import (
	uuid "github.com/satori/go.uuid"
)

// NewOriginator creates a new originator with a new uuid
func NewOriginator() *Originator {
	id := uuid.Must(uuid.NewV4()).String()
	return &Originator{
		Id:      id,
		Version: "1",
	}
}
