package common

import (
	"github.com/makkalot/eskit/lib/types"
	"github.com/satori/go.uuid"
)

func NewOriginator() *types.Originator {
	id := uuid.Must(uuid.NewV4()).String()
	return &types.Originator{
		ID:      id,
		Version: "1",
	}
}
