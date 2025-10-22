package proto

import (
	"github.com/makkalot/eskit/lib/types"
	pb "github.com/makkalot/eskit/generated/grpc/go/eventstore"
)

// AppLogEntryToProto converts a native AppLogEntry to protobuf format
func AppLogEntryToProto(entry *types.AppLogEntry) *pb.AppLogEntry {
	if entry == nil {
		return nil
	}
	return &pb.AppLogEntry{
		Id:    entry.ID,
		Event: EventToProto(entry.Event),
	}
}

// AppLogEntryFromProto converts a protobuf AppLogEntry to native format
func AppLogEntryFromProto(pbEntry *pb.AppLogEntry) *types.AppLogEntry {
	if pbEntry == nil {
		return nil
	}
	return &types.AppLogEntry{
		ID:    pbEntry.Id,
		Event: EventFromProto(pbEntry.Event),
	}
}
