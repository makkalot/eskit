package proto

import (
	"github.com/makkalot/eskit/lib/types"
	pb "github.com/makkalot/eskit/generated/grpc/go/common"
)

// OriginatorToProto converts a native Originator to protobuf format
func OriginatorToProto(o *types.Originator) *pb.Originator {
	if o == nil {
		return nil
	}
	return &pb.Originator{
		Id:      o.ID,
		Version: o.Version,
	}
}

// OriginatorFromProto converts a protobuf Originator to native format
func OriginatorFromProto(pbOrig *pb.Originator) *types.Originator {
	if pbOrig == nil {
		return nil
	}
	return &types.Originator{
		ID:      pbOrig.Id,
		Version: pbOrig.Version,
	}
}
