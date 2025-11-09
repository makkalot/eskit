# ESKit Admin Interface

A web-based administration interface for inspecting data created with the ESKit event sourcing library. This service provides a read-only interface to query and browse events, application logs, and CRUD entities.

## Features

### 1. Raw Events Query (`/events`)
- Query events directly from the event store
- **Filters:**
  - Originator ID (entity UUID)
  - Event Type (e.g., `User.Created`, `CamConfig.Updated`)
  - Date range (from/to)
- Pagination support
- Expandable JSON payloads

### 2. Application Log Query (`/applog`)
- Sequential log of all events across all entities
- **Filters:**
  - Partition ID (entity type like `User`, `CamConfig`)
  - Event Type
  - Date range (from/to)
  - Start from specific Log ID (for offset-based querying)
- Useful for event replay and debugging

### 3. CRUD Entities Browser (`/crud`)
- **Auto-discovery** of entity types from the database
- Entity type listing with counts
- Browse entities of a specific type
- View current state (after event replay)
- View complete event history for each entity
- Timeline visualization with event versioning
- Status indicators (Active/Deleted)

## Technology Stack

- **Backend:** Go with standard `net/http`
- **Templates:** Go `html/template`
- **Frontend:** HTMX + Alpine.js
- **Database:** PostgreSQL or SQLite (via GORM)
- **Styling:** Custom CSS with responsive design

## Installation

### Prerequisites
- Go 1.24 or higher
- PostgreSQL (or SQLite for testing)
- An ESKit-based application with existing data

### Build

```bash
cd services/admin
go build -o bin/admin ./cmd/admin
```

## Configuration

The admin service can be configured via:
1. Environment variables
2. Configuration file (`config.json` or `config.yaml`)
3. Default values

### Configuration Options

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| `listenAddr` | `LISTEN_ADDR` | `:8082` | HTTP server address |
| `dbUri` | `DB_URI` | `inmemory://` | Database connection string |
| `dbDialect` | `DB_DIALECT` | `postgres` | Database dialect (`postgres` or `sqlite3`) |

### Example Configuration File

**config.yaml:**
```yaml
listenAddr: ":8082"
dbUri: "host=localhost port=5432 user=postgres dbname=eventsourcing password=yourpass sslmode=disable"
dbDialect: "postgres"
```

**config.json:**
```json
{
  "listenAddr": ":8082",
  "dbUri": "host=localhost port=5432 user=postgres dbname=eventsourcing password=yourpass sslmode=disable",
  "dbDialect": "postgres"
}
```

### Environment Variables

```bash
export DB_URI="host=localhost port=5432 user=postgres dbname=eventsourcing password=yourpass sslmode=disable"
export DB_DIALECT="postgres"
export LISTEN_ADDR=":8082"
```

## Running the Service

### Using Docker Compose (Recommended)

The easiest way to run the admin service is with Docker Compose, which will also start the database and other services:

```bash
# From the repository root
# Build all services
make build

# Start all services (including admin on port 8082)
docker compose up admin

# Or start everything
docker compose up
```

The admin interface will be available at `http://localhost:8082/`

### From Source
```bash
cd services/admin
go run ./cmd/admin
```

### From Binary
```bash
# First build from repository root
make build-go

# Then run
./bin/admin
```

### With Custom Config
```bash
# Place config.yaml in the current directory
./bin/admin

# Or use environment variables
DB_URI="..." DB_DIALECT="postgres" ./bin/admin
```

### Docker Only

To build and run just the admin service with Docker:

```bash
# From repository root
# Build the Docker image
docker build -f services/admin/Dockerfile -t eskit-admin .

# Run with PostgreSQL connection
docker run -p 8082:8082 \
  -e DB_URI="host=your-db port=5432 user=postgres dbname=eventsourcing password=yourpass sslmode=disable" \
  -e DB_DIALECT="postgres" \
  eskit-admin
```

## Usage

Once the service is running, access the web interface at:

```
http://localhost:8082/
```

### Available Endpoints

- **Home:** `http://localhost:8082/` (redirects to `/events`)
- **Raw Events:** `http://localhost:8082/events`
- **Application Log:** `http://localhost:8082/applog`
- **CRUD Entities:** `http://localhost:8082/crud`

### Example Workflows

#### 1. Find All Events for a Specific Entity
1. Go to `/events`
2. Enter the entity's UUID in "Originator ID" filter
3. Click "Apply Filters"
4. Click on JSON payloads to expand them

#### 2. Browse User Entities
1. Go to `/crud`
2. Click on the "User" entity type card
3. Browse the list of user entities
4. Click "View History" to see the complete event timeline

#### 3. Debug Event Replay
1. Go to `/applog`
2. Filter by partition ID (e.g., "User")
3. Review the sequential order of events
4. Use "Start from Log ID" to resume from a specific point

## Project Structure

```
services/admin/
├── cmd/admin/
│   └── main.go                 # Application entry point
├── provider/
│   ├── provider.go             # Service provider
│   └── handlers.go             # HTTP handlers and business logic
├── templates/
│   ├── layout.html             # Base layout with navigation
│   ├── events.html             # Raw events page
│   ├── applog.html             # Application log page
│   ├── crud.html               # CRUD entities listing
│   └── crud_entity.html        # Entity detail page
├── static/
│   └── css/
│       └── styles.css          # Custom styles
├── bin/                        # Compiled binaries
├── config.yaml                 # Optional configuration file
└── README.md                   # This file
```

## Database Requirements

The admin service requires read access to the following tables:

- **stored_events** - Contains all events indexed by originator ID and version
- **stored_log_entries** - Sequential application log with auto-incrementing IDs

These tables are automatically created by ESKit's event store when using the SQL backend.

## Connecting to Existing Applications

### Example: Connecting to CamConfig Service Database

```bash
# If CamConfig is running on port 8081 with PostgreSQL
cd services/admin
DB_URI="host=localhost port=5432 user=postgres dbname=eventsourcing password=t00r sslmode=disable" \
DB_DIALECT="postgres" \
./bin/admin
```

Now you can inspect all CamConfig entities at `http://localhost:8082/crud?type=CamConfig`

## Limitations

- **Read-only interface** - Currently supports inspection only, no mutations
- **In-memory mode** has limited functionality (no persistent storage)
- **No authentication** - Do not expose to untrusted networks
- **No pagination on entity history** - Large entities with many events may be slow

## Future Enhancements

Potential features for future versions:

- [ ] Write operations (append events, create/update/delete CRUD entities)
- [ ] Authentication and authorization
- [ ] Export functionality (CSV, JSON)
- [ ] Event search across all entities
- [ ] Real-time event streaming with WebSockets
- [ ] Event replay visualization
- [ ] Diff view for entity state changes
- [ ] Performance optimizations for large datasets

## Development

### Adding New Features

1. Add HTTP handler in `provider/handlers.go`
2. Create HTML template in `templates/`
3. Add route in `SetupRoutes()` method
4. Update CSS in `static/css/styles.css` if needed

### Template System

The service uses Go's standard `html/template` package with:
- Layout template (`layout.html`) for consistent UI
- Named templates for content sections
- HTMX for dynamic partial updates
- Alpine.js for client-side interactivity

### Testing Locally

```bash
# Start with in-memory store (limited functionality)
go run ./cmd/admin

# Or connect to a test database
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=test postgres:latest
DB_URI="host=localhost port=5432 user=postgres password=test sslmode=disable" go run ./cmd/admin
```

## Troubleshooting

### "Error loading templates"
- Ensure `templates/` directory exists in the working directory
- Templates must be in the same directory as the binary or use absolute path

### "Database connection failed"
- Verify database is running and accessible
- Check connection string format
- Ensure database contains ESKit tables

### "No entity types found"
- Ensure there's data in the database
- Check that stored_log_entries table has partition_id values
- Verify event types follow the "EntityType.Action" naming convention

## License

This service is part of the ESKit project. See the root LICENSE file for details.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Support

For issues and questions:
- GitHub Issues: [ESKit Issues](https://github.com/makkalot/eskit/issues)
- Documentation: See `/docs` in the repository root
