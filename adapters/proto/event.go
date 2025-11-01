package proto

import (
	"time"

	"github.com/makkalot/eskit/lib/types"
	pb "github.com/makkalot/eskit/generated/grpc/go/eventstore"
)

// EventToProto converts a native Event to protobuf format
func EventToProto(e *types.Event) *pb.Event {
	if e == nil {
		return nil
	}

	var occurredOn int64
	if !e.OccurredOn.IsZero() {
		occurredOn = e.OccurredOn.Unix()
	}

	return &pb.Event{
		Originator: OriginatorToProto(e.Originator),
		EventType:  e.EventType,
		Payload:    e.Payload,
		OccuredOn:  occurredOn,
	}
}

// EventFromProto converts a protobuf Event to native format
func EventFromProto(pbEvent *pb.Event) *types.Event {
	if pbEvent == nil {
		return nil
	}

	var occurredOn time.Time
	if pbEvent.OccuredOn > 0 {
		occurredOn = time.Unix(pbEvent.OccuredOn, 0).UTC()
	}

	return &types.Event{
		Originator: OriginatorFromProto(pbEvent.Originator),
		EventType:  pbEvent.EventType,
		Payload:    pbEvent.Payload,
		OccurredOn: occurredOn,
	}
}
