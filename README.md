[![Build Status](https://travis-ci.com/makkalot/eskit.svg?branch=master)](https://travis-ci.com/makkalot/eskit)

## ESKIT (Event Sourcing Kit)

ESKIT is an Event Sourcing toolkit designed to be used as a **pure Go library** for embedding event sourcing capabilities directly into your applications.

As Greg Young mentions in his talks, if you write an event sourcing framework you're probably doing something wrong.
ESKIT is not a framework - it's a lightweight library with example services that demonstrate how to build REST APIs on top of it.

### Version 2.0 - Library-First Architecture

v2.0 is a complete refactoring where the core library (`lib/`) is **completely independent** of any RPC framework.
The library uses pure Go types with no external protocol dependencies (no gRPC, no protobuf).

#### What is ESKIT:

- **Event Store Library** - Append-only event storage with Application Log pattern (from Vaughn Vernon's book)
- **CRUD Store Library** - Built on event store with predefined event types that handles replay automatically
- **Consumer Library** - Process events from Application Log with offset tracking
- **Example User Service** - Demonstrates building a JSON REST API on top of the library
- Easy way to implement CQRS with event sourcing
- Supports both in-memory and PostgreSQL storage backends

#### What ESKIT is not

- It's not an Event Sourcing Framework
- It's not production ready
- It's not a message broker like Kafka


## Usage

### Library Usage (Recommended)

Use ESKIT as an embedded library in your Go application. **No code generation or external protocols required.**

```go
package main

import (
    "context"
    "time"

    "github.com/makkalot/eskit/lib/eventstore"
    "github.com/makkalot/eskit/lib/crudstore"
    "github.com/makkalot/eskit/lib/types"
)

func main() {
    // Create an in-memory event store (or use SQL store with PostgreSQL)
    store := eventstore.NewInMemoryStore()

    // Create events with native Go types
    event := &types.Event{
        Originator: &types.Originator{
            ID:      "user-123",
            Version: "1",
        },
        EventType:  "User.Created",
        Payload:    `{"email":"user@example.com"}`,
        OccurredOn: time.Now().UTC(),
    }

    // Append event to store
    store.Append(event)

    // Or use CRUD store for automatic event replay
    ctx := context.Background()
    crudClient, _ := crudstore.NewClient(ctx, "postgres://...")

    // Define your entity type
    type User struct {
        Originator *types.Originator
        Email      string
        FirstName  string
    }

    user := &User{
        Email:     "user@example.com",
        FirstName: "John",
    }

    // CRUD operations automatically create events
    originator, _ := crudClient.Create(user)
}
```

**Benefits:**
- ✅ No code generation needed
- ✅ Pure Go types (no proto dependencies)
- ✅ Easy to test and embed
- ✅ Framework-agnostic (use with any HTTP framework, gRPC, GraphQL, etc.)

### Example REST API Service

The `services/users` directory contains an example REST API built on top of the ESKIT library, demonstrating how to create a real-world service.

**REST API Endpoints (Port 8080):**
```bash
# Health check
GET /v1/health

# User CRUD operations
POST   /v1/users                    # Create user
GET    /v1/users?id=X&version=Y    # Get user
PUT    /v1/users?id=X&version=Y    # Update user
DELETE /v1/users?id=X&version=Y    # Delete user

# Prometheus metrics
GET /metrics
```

**Example API calls:**
```bash
# Create a user
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","firstName":"John","lastName":"Doe"}'

# Get a user
curl "http://localhost:8080/v1/users?id=USER_ID&version=1"

# Update a user
curl -X PUT "http://localhost:8080/v1/users?id=USER_ID&version=1" \
  -H "Content-Type: application/json" \
  -d '{"email":"newemail@example.com","firstName":"Jane","active":true}'
```

**Build and run:**
```bash
# Build the service
make build-go

# Run locally (requires PostgreSQL)
export DB_URI="host=localhost port=5432 user=postgres dbname=eventsourcing password=pass sslmode=disable"
./bin/users

# Or via Docker Compose
make deploy-compose
```

## Migration from v1.x to v2.0

See [MIGRATION.md](MIGRATION.md) for detailed migration guide.

**Key breaking changes:**
- Library now uses `lib/types` instead of `generated/grpc/go/*`
- Field names follow Go conventions: `.Id` → `.ID`
- Timestamp fields use `time.Time` instead of Unix `int64`

## Architecture Overview

ESKIT is organized as a library-first toolkit with the following components:

### Core Libraries (lib/)

**Event Store (`lib/eventstore/`)**
- Append-only event storage supporting `Append` and `Get` operations
- Implements Application Log pattern (event stream similar to Kafka)
- Events are written to both event store and application log in the same transaction
- Consumers can poll the Application Log to process all events flowing through the system
- Storage backends: In-memory (for testing) and PostgreSQL (for production)

**CRUD Store (`lib/crudstore/`)**
- Built on top of Event Store with automatic event replay
- Supports 4 basic CRUD operations: Create, Read, Update, Delete
- Uses 3 predefined event types: `Created`, `Updated`, `Deleted`
- Automatically handles event replay to reconstruct current entity state
- Stores only diffs (JSON Merge Patches) on updates, keeping storage efficient
- Works like a NoSQL database with full history
- Using CRUD Store is optional - you can use Event Store directly and handle replay yourself
- Note: Snapshotting not yet implemented (planned for future)

**Consumer Store (`lib/consumerstore/`)**
- Tracks consumer progress when reading the Application Log
- Stores consumer offsets so consumers can resume after crashes
- Supports both in-memory and SQL storage backends

**Consumer Library (`lib/consumer/`)**
- Reference implementation for processing events from Application Log
- Automatically manages offset tracking via Consumer Store

### Example Service (services/users/)

The User Service demonstrates how to build a REST API on top of ESKIT:
- Uses CRUD Store library for data persistence
- Exposes JSON REST API for user management
- Shows how to integrate ESKIT into a real service
- Includes Prometheus metrics integration

**Architecture Benefits:**
- ✅ **No network calls needed** - Library runs in-process
- ✅ **Framework agnostic** - Build REST APIs, gRPC services, GraphQL servers, or CLI tools
- ✅ **Simple deployment** - Single binary with embedded library
- ✅ **Easy testing** - Use in-memory store for fast unit tests


 
## Development and Testing

### Requirements
- Go 1.23 or later
- Docker (with compose plugin) - for running integration tests
- PostgreSQL - for production usage (optional, in-memory store available for development)

### Build

```bash
# Build the example user service
make build-go

# Binary will be created at ./bin/users
```

### Run Tests

```bash
# Run unit tests (no Docker required)
make test-go-unit

# Run integration tests (requires Docker)
make test

# Run tests in Docker Compose
make test-compose-go
```

### Run Example User Service Locally

**Option 1: Using Docker Compose (easiest)**
```bash
make deploy-compose
```
This will start PostgreSQL and the user service. The API will be available at `http://localhost:8080`.

**Option 2: Run locally (requires PostgreSQL)**
```bash
# Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=t00r \
  -e POSTGRES_DB=eventsourcing \
  postgres

# Set database URI
export DB_URI="host=localhost port=5432 user=postgres dbname=eventsourcing password=t00r sslmode=disable"

# Run the service
./bin/users
```

The service will start on port 8080 (configurable via `listenAddr` in config.yaml or environment).


### How to contribute ?

- Create a fork
- Clone your own fork : `git clone git@gitlab.com:<you-user-name>/eskit.git`
- cd `eskit`
- Add upstream : `git remote add upstream git@github.com/makkalot/eskit.git`
- Fetch The latest changes from the upstream : `git fetch upstream`
- Merge them to your master : `git rebase upstream/master`
- Create the branch you want to work on : `git checkout -b <branch-name>`
- When ready with your changes push them to your branch : `git push origin -u <branch-name>`
- Before open a Merge Request : rebase against the upstream master branch (if someone did some change)
    - `git checkout master`
    - `git fetch upstream`
    - `git rebase upstream/master`
    - `git checkout <branch-name>`
    - `git rebase -i master`
- Push latest changes to your branch (after the rebase):
    - `git push origin -f <branch-name>` (Note that we apply -f because of the rebase in prev step, careful !)
- Create a MR.

For more information about rebasing :
    - https://www.atlassian.com/git/tutorials/rewriting-history
    - https://www.atlassian.com/git/tutorials/merging-vs-rebasing

Maintainer :
- Denis Kyorov   : makkalot at gmail dot com

Contributors:

- Korhan Yazgan  : korhanyazgan at gmail dot com

