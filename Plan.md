# ESKIT Library Refactoring Plan: Remove gRPC Dependencies

**Goal**: Decouple `/lib/` from gRPC dependencies, making it a pure Go library with no protobuf/gRPC requirements.

**Status**: Planning Phase
**Created**: 2025-10-22

---

## Current State Analysis

### Dependencies Problem
```
lib/eventstore → generated/grpc/go/eventstore → grpc + genproto + gateway
lib/crudstore  → generated/grpc/go/crudstore  → grpc + genproto + gateway
lib/consumer   → generated/grpc/go/eventstore → grpc + genproto + gateway
```

### Current Imports in Library Code
- `github.com/makkalot/eskit/generated/grpc/go/common` (for `Originator`)
- `github.com/makkalot/eskit/generated/grpc/go/eventstore` (for `Event`, `AppLogEntry`)
- `github.com/makkalot/eskit/generated/grpc/go/crudstore` (for CRUD types)

### Impact
- Heavy dependency chain (grpc, genproto, grpc-gateway)
- Requires code generation for library usage
- Network issues block library compilation
- Confusing for library-only users

---

## Target Architecture

### New Structure
```
eskit/
├── lib/                                    # Pure Go - NO gRPC deps
│   ├── types/                              # Native Go types
│   │   ├── originator.go                   # ID + Version
│   │   ├── event.go                        # Event structure
│   │   ├── applog.go                       # Application log types
│   │   └── crud.go                         # CRUD-related types
│   ├── eventstore/                         # Event store implementations
│   │   ├── store.go                        # Store interface (uses lib/types)
│   │   ├── memorystore.go                  # In-memory impl
│   │   └── sqlstore.go                     # SQL impl
│   ├── crudstore/                          # CRUD wrapper
│   │   ├── client.go                       # Client interface (uses lib/types)
│   │   └── provider.go                     # Implementation
│   ├── consumer/                           # Consumer library
│   │   └── consumer.go                     # Consumer (uses lib/types)
│   └── consumerstore/                      # Consumer offset tracking
│       └── store.go                        # Store interface
│
├── adapters/                               # NEW: Conversion layer
│   ├── proto/                              # Protobuf conversions
│   │   ├── event.go                        # lib.Event ↔ pb.Event
│   │   ├── originator.go                   # lib.Originator ↔ pb.Originator
│   │   └── crud.go                         # lib.CRUD ↔ pb.CRUD
│   └── adapters_test.go                    # Conversion tests
│
├── contracts/                              # Proto definitions (unchanged)
│   ├── eventstore/
│   ├── crudstore/
│   └── users/
│
├── generated/                              # Generated gRPC code (unchanged)
│
└── services/                               # Microservices (uses adapters)
    └── users/
        └── server.go                       # Uses adapters to convert
```

### Dependency Flow
```
Old:
  lib → generated/grpc → grpc dependencies ❌

New:
  lib → stdlib only ✅
  services → lib + adapters + generated/grpc ✅
```

---

## Migration Steps

### Phase 1: Create Native Types (Week 1)

**Task 1.1**: Create `lib/types/originator.go`
```go
package types

// Originator identifies an entity and its version
type Originator struct {
    ID      string
    Version string
}
```

**Task 1.2**: Create `lib/types/event.go`
```go
package types

import "time"

// Event represents a domain event
type Event struct {
    Originator *Originator
    EventType  string
    Payload    string
    Timestamp  time.Time
}
```

**Task 1.3**: Create `lib/types/applog.go`
```go
package types

// AppLogEntry represents an entry in the application log
type AppLogEntry struct {
    ID    string
    Event *Event
}
```

**Task 1.4**: Create `lib/types/crud.go`
```go
package types

// CrudEntitySpec defines entity schema
type CrudEntitySpec struct {
    EntityType string
    Schema     string
}

// CrudEntity represents a CRUD entity with metadata
type CrudEntity struct {
    EntityType string
    Originator *Originator
    Payload    string
    Deleted    bool
}
```

**Task 1.5**: Add JSON/SQL serialization tags
- Add `json` tags for JSON marshaling
- Add `gorm` tags for database mapping (if needed)

---

### Phase 2: Create Adapters (Week 1)

**Task 2.1**: Create `adapters/proto/originator.go`
```go
package proto

import (
    "github.com/makkalot/eskit/lib/types"
    pb "github.com/makkalot/eskit/generated/grpc/go/common"
)

func OriginatorToProto(o *types.Originator) *pb.Originator {
    if o == nil {
        return nil
    }
    return &pb.Originator{
        Id:      o.ID,
        Version: o.Version,
    }
}

func OriginatorFromProto(pb *pb.Originator) *types.Originator {
    if pb == nil {
        return nil
    }
    return &types.Originator{
        ID:      pb.Id,
        Version: pb.Version,
    }
}
```

**Task 2.2**: Create `adapters/proto/event.go`
- `EventToProto(e *types.Event) *pb.Event`
- `EventFromProto(pb *pb.Event) *types.Event`

**Task 2.3**: Create `adapters/proto/applog.go`
- `AppLogEntryToProto(e *types.AppLogEntry) *pb.AppLogEntry`
- `AppLogEntryFromProto(pb *pb.AppLogEntry) *types.AppLogEntry`

**Task 2.4**: Create `adapters/proto/crud.go`
- CRUD-related type conversions

**Task 2.5**: Write comprehensive adapter tests
- Test bidirectional conversion
- Test nil handling
- Test nested structures

---

### Phase 3: Refactor Library Code (Week 2)

**Task 3.1**: Update `lib/eventstore/store.go`
```go
// OLD
import store "github.com/makkalot/eskit/generated/grpc/go/eventstore"

type Store interface {
    Append(event *store.Event) error
}

// NEW
import "github.com/makkalot/eskit/lib/types"

type Store interface {
    Append(event *types.Event) error
}
```

**Task 3.2**: Update `lib/eventstore/memorystore.go`
- Replace proto imports with `lib/types`
- Update all type references
- Ensure all methods use native types

**Task 3.3**: Update `lib/eventstore/sqlstore.go`
- Replace proto imports
- Update GORM mappings if needed
- Update serialization logic

**Task 3.4**: Update `lib/crudstore/client.go`
- Replace proto imports
- Update all CRUD operations

**Task 3.5**: Update `lib/crudstore/provider.go`
- Replace proto imports
- Update event replay logic

**Task 3.6**: Update `lib/consumer/consumer.go`
- Replace proto imports
- Update consumer callback signatures

**Task 3.7**: Update `lib/consumerstore/`
- Replace proto imports
- Update all interfaces

**Task 3.8**: Update all library tests
- Fix imports
- Update test data to use native types
- Ensure all tests pass

---

### Phase 4: Update Services (Week 3)

**Task 4.1**: Update `services/users/server.go`
```go
// Add adapter imports
import (
    "github.com/makkalot/eskit/lib/eventstore"
    "github.com/makkalot/eskit/adapters/proto"
)

// Convert at service boundary
func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
    // Use library with native types
    libEvent := &types.Event{...}
    s.store.Append(libEvent)

    // Convert back to proto for response
    return &pb.CreateResponse{
        User: proto.UserToProto(libUser),
    }, nil
}
```

**Task 4.2**: Update all gRPC service implementations
- Add adapter layer at service boundaries
- Convert incoming proto → native types
- Process with library
- Convert outgoing native types → proto

**Task 4.3**: Update service tests
- Fix integration tests
- Ensure services still work correctly

---

### Phase 5: Update Build & Documentation (Week 3)

**Task 5.1**: Update `go.mod`
- Verify lib dependencies are clean
- No grpc/genproto in lib dependencies

**Task 5.2**: Update `Makefile`
- Library tests don't require proto generation
- Service builds still generate protos first

**Task 5.3**: Update `README.md`
```markdown
## Using as a Library

```go
import "github.com/makkalot/eskit/lib/eventstore"

// No protobuf compilation needed!
store := eventstore.NewInMemoryStore()
```

## Using as Microservices

Requires protobuf generation:
```bash
make generate-grpc
```
```

**Task 5.4**: Create migration guide
- Document breaking changes
- Provide code examples for upgrading
- Show before/after patterns

**Task 5.5**: Update code comments
- Add godoc comments to all exported types
- Explain library vs service usage

---

## Breaking Changes

### For Library Users

**Before:**
```go
import (
    "github.com/makkalot/eskit/generated/grpc/go/common"
    store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
)

originator := &common.Originator{Id: "123", Version: "1"}
event := &store.Event{Originator: originator}
```

**After:**
```go
import "github.com/makkalot/eskit/lib/types"

originator := &types.Originator{ID: "123", Version: "1"}
event := &types.Event{Originator: originator}
```

### Field Name Changes
- `Originator.Id` → `Originator.ID` (Go naming conventions)
- `Event.event_type` → `Event.EventType` (proto snake_case → Go camelCase)

### For Service Implementations
- Must use adapters at boundaries
- Internal logic uses native types
- gRPC handlers convert proto ↔ native

---

## Testing Strategy

### Phase 1-2 Tests
- [ ] Unit tests for all native types
- [ ] Unit tests for all adapters
- [ ] Bidirectional conversion tests
- [ ] Benchmark adapter performance

### Phase 3 Tests
- [ ] All existing library tests pass
- [ ] No gRPC imports in lib/
- [ ] `go mod graph` shows clean dependencies

### Phase 4 Tests
- [ ] All service unit tests pass
- [ ] Integration tests with Docker Compose
- [ ] End-to-end gRPC tests
- [ ] Python client compatibility

### Phase 5 Tests
- [ ] Documentation examples compile
- [ ] Clean `make test` on fresh clone

---

## Rollout Plan

### Version Strategy
- **v1.x**: Current implementation (deprecated)
- **v2.0**: Breaking change with native types
- Tag both for Go module compatibility

### Communication
1. Create GitHub issue explaining changes
2. Update README with migration guide
3. Add CHANGELOG.md with v2.0 notes
4. Consider blog post if project has users

### Backwards Compatibility (Optional)
If needed, could maintain v1 compatibility:
```go
// lib/compat/types.go
import pb "github.com/makkalot/eskit/generated/grpc/go/eventstore"

// Deprecated: Use types.Event instead
type Event = pb.Event
```

---

## Risk Mitigation

### Risk 1: Adapter Performance Overhead
- **Mitigation**: Benchmark conversions, optimize if needed
- **Acceptance**: Overhead only at service boundaries, not in library

### Risk 2: Type Drift (Proto vs Native)
- **Mitigation**: CI checks that adapters are up-to-date
- **Mitigation**: Generate adapter stubs from protos (future)

### Risk 3: Breaking External Users
- **Mitigation**: Semantic versioning (v2.0)
- **Mitigation**: Maintain v1 branch for critical fixes
- **Mitigation**: Migration guide with examples

### Risk 4: Missed Proto Dependencies
- **Mitigation**: `go mod graph | grep grpc` in CI for lib/
- **Mitigation**: Separate module for lib/ (future)

---

## Success Criteria

- [ ] `lib/` has zero gRPC/protobuf dependencies
- [ ] All library tests pass with native types
- [ ] Services work correctly with adapters
- [ ] Integration tests pass
- [ ] Documentation updated
- [ ] Migration guide written
- [ ] Can compile library without `make generate-grpc`

---

## Future Enhancements (Post-Refactor)

1. **Separate Module**: Move `lib/` to separate Go module
   - `github.com/makkalot/eskit-lib` (pure Go)
   - `github.com/makkalot/eskit` (services + adapters)

2. **Generate Adapters**: Auto-generate adapters from proto
   - Add `protoc-gen-go-adapter` plugin
   - Keep proto as source of truth
   - Generate both native types and adapters

3. **Alternative Serialization**: Add more serialization options
   - JSON codec
   - MessagePack
   - Custom binary format

---

## Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Native Types | 2-3 days | None |
| Phase 2: Adapters | 2-3 days | Phase 1 |
| Phase 3: Refactor Library | 3-4 days | Phase 2 |
| Phase 4: Update Services | 3-4 days | Phase 3 |
| Phase 5: Docs & Release | 2-3 days | Phase 4 |
| **Total** | **2-3 weeks** | |

---

## Next Steps

1. Review this plan
2. Create GitHub issue for tracking
3. Set up feature branch: `feature/remove-grpc-from-lib`
4. Begin Phase 1: Create native types
5. Iterate with tests at each phase

---

## Notes

- This is a **breaking change** requiring major version bump
- Consider community feedback before starting
- Can be done incrementally (phase by phase)
- Each phase can be committed separately for easier review
