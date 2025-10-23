// Package proto provides adapters for converting between native Go types and protobuf types.
//
// This package bridges the gap between ESKIT's pure Go library types (lib/types)
// and the gRPC protobuf types (generated/grpc/go/*) used for network communication.
//
// # Purpose
//
// When running ESKIT as microservices, gRPC handlers receive protobuf-generated types
// from the network. This package provides conversion functions to:
//
//  1. Convert incoming proto types to native Go types
//  2. Pass native types to the library (which has no proto dependencies)
//  3. Convert results back to proto types for responses
//
// # Usage Pattern
//
// In a gRPC service implementation:
//
//	import (
//	    "github.com/makkalot/eskit/adapters/proto"
//	    "github.com/makkalot/eskit/lib/types"
//	    pb "github.com/makkalot/eskit/generated/grpc/go/users"
//	)
//
//	func (s *Service) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
//	    // 1. Convert proto Originator to native
//	    nativeOrig := proto.OriginatorFromProto(req.Originator)
//
//	    // 2. Use library with native types (no proto!)
//	    result, err := s.library.DoWork(nativeOrig)
//
//	    // 3. Convert native result back to proto
//	    return &pb.CreateResponse{
//	        Originator: proto.OriginatorToProto(result),
//	    }, nil
//	}
//
// # Conversion Functions
//
// Each type has bidirectional converters:
//
//   - OriginatorToProto / OriginatorFromProto
//   - EventToProto / EventFromProto
//   - AppLogEntryToProto / AppLogEntryFromProto
//   - CrudEntityToProto / CrudEntityFromProto
//
// All functions handle nil values gracefully, returning nil if input is nil.
//
// # Architecture
//
// This adapter layer enables a clean separation:
//
//	Network (proto) → Adapter → Library (native Go) → Adapter → Network (proto)
//	     gRPC            ↓         lib/eventstore        ↓         gRPC
//	                  ToProto()                    FromProto()
//
// The library layer (lib/*) remains completely independent of gRPC/protobuf,
// while services use these adapters at their boundaries.
//
// # When to Use
//
// Use this package when:
//
//   - Implementing gRPC service handlers (services/)
//   - Converting between microservice boundaries
//   - Bridging proto and native type systems
//
// Do NOT use this package when:
//
//   - Using ESKIT as an embedded library (use lib/types directly)
//   - Writing pure library code (lib/* should never import this)
//   - Writing tests for library code (use native types)
package proto
