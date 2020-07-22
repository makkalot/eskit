package common

import (
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/satori/go.uuid"
)

func NewOriginator() *common.Originator {
	id := uuid.Must(uuid.NewV4()).String()
	return &common.Originator{
		Id:      id,
		Version: "1",
	}
}
