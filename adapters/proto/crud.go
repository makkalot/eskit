package proto

import (
	"github.com/makkalot/eskit/lib/types"
	pb "github.com/makkalot/eskit/generated/grpc/go/crudstore"
)

// SchemaSpecToProto converts a native SchemaSpec to protobuf format
func SchemaSpecToProto(spec *types.SchemaSpec) *pb.SchemaSpec {
	if spec == nil {
		return nil
	}
	return &pb.SchemaSpec{
		SchemaVersion: spec.SchemaVersion,
		JsonSchema:    spec.JSONSchema,
	}
}

// SchemaSpecFromProto converts a protobuf SchemaSpec to native format
func SchemaSpecFromProto(pbSpec *pb.SchemaSpec) *types.SchemaSpec {
	if pbSpec == nil {
		return nil
	}
	return &types.SchemaSpec{
		SchemaVersion: pbSpec.SchemaVersion,
		JSONSchema:    pbSpec.JsonSchema,
	}
}

// CrudEntitySpecToProto converts a native CrudEntitySpec to protobuf format
func CrudEntitySpecToProto(spec *types.CrudEntitySpec) *pb.CrudEntitySpec {
	if spec == nil {
		return nil
	}
	return &pb.CrudEntitySpec{
		EntityType: spec.EntityType,
		SchemaSpec: SchemaSpecToProto(spec.SchemaSpec),
	}
}

// CrudEntitySpecFromProto converts a protobuf CrudEntitySpec to native format
func CrudEntitySpecFromProto(pbSpec *pb.CrudEntitySpec) *types.CrudEntitySpec {
	if pbSpec == nil {
		return nil
	}
	return &types.CrudEntitySpec{
		EntityType: pbSpec.EntityType,
		SchemaSpec: SchemaSpecFromProto(pbSpec.SchemaSpec),
	}
}
