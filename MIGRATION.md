# Migration Guide: v1.x → v2.0

## Overview

ESKIT v2.0 introduces a major architectural change where the core library (`lib/`) is now **completely independent** of gRPC/protobuf. This enables using ESKIT as a pure Go library without any code generation requirements.

## Breaking Changes

### 1. Import Paths Changed

**Before (v1.x):**
```go
import (
    "github.com/makkalot/eskit/generated/grpc/go/common"
    store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
)
```

**After (v2.0):**
```go
import (
    "github.com/makkalot/eskit/lib/types"
    "github.com/makkalot/eskit/lib/eventstore"
)
```

### 2. Type Changes

#### Originator

**Before:**
```go
originator := &common.Originator{
    Id:      "user-123",
    Version: "1",
}
```

**After:**
```go
originator := &types.Originator{
    ID:      "user-123",  // Note: Id → ID (Go naming conventions)
    Version: "1",
}
```

#### Event

**Before:**
```go
event := &store.Event{
    Originator: originator,
    EventType:  "User.Created",
    Payload:    `{"email":"user@example.com"}`,
    OccurredOn: time.Now().Unix(),  // Unix timestamp (int64)
}
```

**After:**
```go
event := &types.Event{
    Originator: originator,
    EventType:  "User.Created",
    Payload:    `{"email":"user@example.com"}`,
    OccurredOn: time.Now().UTC(),  // time.Time instead of int64
}
```

#### AppLogEntry

**Before:**
```go
import store "github.com/makkalot/eskit/generated/grpc/go/eventstore"

entry := &store.AppLogEntry{
    Id:    "123",
    Event: event,
}
```

**After:**
```go
import "github.com/makkalot/eskit/lib/types"

entry := &types.AppLogEntry{
    ID:    "123",  // Note: Id → ID
    Event: event,
}
```

### 3. Field Name Changes

All field names now follow Go naming conventions:

| Before (proto) | After (native Go) |
|----------------|-------------------|
| `.Id`          | `.ID`             |
| `.id`          | `.ID`             |

### 4. Timestamp Handling

**Before:** Timestamps were `int64` (Unix seconds)
```go
event.OccurredOn = time.Now().Unix()
timestamp := time.Unix(event.OccurredOn, 0)
```

**After:** Timestamps are `time.Time`
```go
event.OccurredOn = time.Now().UTC()
timestamp := event.OccurredOn  // Already time.Time
```

### 5. Event Store Interface

**Before:**
```go
import store "github.com/makkalot/eskit/generated/grpc/go/eventstore"

type Store interface {
    Append(event *store.Event) error
    Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error)
}
```

**After:**
```go
import "github.com/makkalot/eskit/lib/types"

type Store interface {
    Append(event *types.Event) error
    Get(originator *types.Originator, fromVersion bool) ([]*types.Event, error)
}
```

### 6. CRUD Store Client

**Before:**
```go
import "github.com/makkalot/eskit/generated/grpc/go/common"

type User struct {
    Originator *common.Originator
    Email      string
}
```

**After:**
```go
import "github.com/makkalot/eskit/lib/types"

type User struct {
    Originator *types.Originator
    Email      string
}
```

### 7. Consumer Callbacks

**Before:**
```go
import store "github.com/makkalot/eskit/generated/grpc/go/eventstore"

consumer.Consume(ctx, func(entry *store.AppLogEntry) error {
    // Process entry
    return nil
})
```

**After:**
```go
import "github.com/makkalot/eskit/lib/types"

consumer.Consume(ctx, func(entry *types.AppLogEntry) error {
    // Process entry
    return nil
})
```

## Migration Steps

### Step 1: Update Imports

Replace all proto imports with `lib/types` imports:

```bash
# Find all files using proto types
grep -r "github.com/makkalot/eskit/generated/grpc/go" your_project/

# Replace with lib/types imports
```

### Step 2: Update Type References

1. Change `common.Originator` → `types.Originator`
2. Change `store.Event` → `types.Event`
3. Change `store.AppLogEntry` → `types.AppLogEntry`

### Step 3: Fix Field Names

Update all field accesses:
```bash
# Find and replace .Id with .ID
sed -i 's/\.Id\b/.ID/g' your_files.go
```

### Step 4: Update Timestamp Handling

Replace Unix timestamp handling:

**Before:**
```go
event.OccurredOn = time.Now().Unix()
t := time.Unix(event.OccurredOn, 0)
```

**After:**
```go
event.OccurredOn = time.Now().UTC()
t := event.OccurredOn
```

### Step 5: Update Tests

Update test data creation:

**Before:**
```go
testEvent := &store.Event{
    Originator: &common.Originator{Id: "test", Version: "1"},
    OccurredOn: 1234567890,
}
```

**After:**
```go
testEvent := &types.Event{
    Originator: &types.Originator{ID: "test", Version: "1"},
    OccurredOn: time.Now().UTC(),
}
```

## For Service Implementations

If you're implementing gRPC services (microservices mode), you'll need to use the adapter layer:

### Before (v1.x)
```go
func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
    user := &pb.User{
        Email: req.Email,
    }
    s.crudStore.Create(user)  // Directly passed proto type
    return &pb.CreateResponse{User: user}, nil
}
```

### After (v2.0)
```go
import (
    "github.com/makkalot/eskit/adapters/proto"
    "github.com/makkalot/eskit/lib/types"
)

// Define native type
type User struct {
    Originator *types.Originator
    Email      string
}

func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
    // Convert request to native type
    nativeUser := &User{
        Email: req.Email,
    }

    // Use library with native types
    s.crudStore.Create(nativeUser)

    // Convert back to proto for response
    return &pb.CreateResponse{
        User: &pb.User{
            Originator: proto.OriginatorToProto(nativeUser.Originator),
            Email:      nativeUser.Email,
        },
    }, nil
}
```

## Compatibility

### Go Module Version

Update your `go.mod`:
```go
require github.com/makkalot/eskit v2.0.0
```

### Backward Compatibility

**v2.0 is NOT backward compatible** with v1.x. This is a breaking change that requires code updates.

If you need to maintain v1.x compatibility:
1. Pin to v1.x in go.mod: `github.com/makkalot/eskit v1.x.x`
2. Plan migration to v2.0 when ready

## Benefits of Upgrading

✅ **No protobuf compilation required** for library usage
✅ **Pure Go types** - easier to work with and test
✅ **Better Go conventions** - `.ID` instead of `.Id`
✅ **Simpler dependencies** - no gRPC/genproto in library code
✅ **Better IDE support** - native Go types work better with tools
✅ **Offline development** - no network required for library usage

## Getting Help

If you encounter issues during migration:

1. Check the examples in the `examples/` directory (coming soon)
2. Review the updated tests in `lib/` for usage patterns
3. Open an issue on GitHub with your migration question

## Summary Checklist

- [ ] Update import statements (`generated/grpc/go/*` → `lib/types`)
- [ ] Change field names (`.Id` → `.ID`)
- [ ] Update timestamp handling (`int64` → `time.Time`)
- [ ] Update type references (`common.Originator` → `types.Originator`)
- [ ] Update service implementations to use adapters (if using microservices)
- [ ] Update tests with new types
- [ ] Run tests to verify migration
- [ ] Update `go.mod` to v2.0

---

**Note:** This is a one-time migration cost. Once migrated, your code will be cleaner and easier to maintain.
