# CamConfig Service

A camera configuration management service that demonstrates the eskit library's features including:
- Event sourcing with CRUD operations
- Optimistic locking with versioning
- Audit logging capabilities
- REST API with JSON responses
- Web interface with HTMX

## Features

### Camera Configuration Fields
- **Camera ID**: Unique identifier for the camera
- **Name**: Friendly name for the camera
- **Gamma**: Gamma correction value (0-5)
- **Exposure**: Exposure time in microseconds
- **Saturation**: Saturation level (0-100)
- **Sharpness**: Sharpness level (0-100)
- **Gain**: Gain value (0-100)

### API Endpoints

#### Health Check
```bash
GET /v1/health
```

#### Create Configuration
```bash
POST /v1/camconfigs
Content-Type: application/json

{
  "cameraId": "CAM-001",
  "name": "Front Door Camera",
  "gamma": 1.2,
  "exposure": 1500,
  "saturation": 60,
  "sharpness": 55,
  "gain": 30
}
```

#### Get Configuration
```bash
GET /v1/camconfigs?id={id}&version={version}
```
- `id` (required): Configuration ID
- `version` (optional): Specific version to retrieve
- `fetchDeleted` (optional): Set to "true" to fetch deleted configs

#### Update Configuration
```bash
PUT /v1/camconfigs?id={id}&version={version}
Content-Type: application/json

{
  "gamma": 1.5,
  "exposure": 2000,
  "saturation": 70
}
```
Note: Version parameter is required for optimistic locking.

#### Delete Configuration
```bash
DELETE /v1/camconfigs?id={id}&version={version}
```

### Web Interface

The service includes a web interface accessible at:
- **Main Dashboard**: http://localhost:8081/web/
- **Audit Log**: http://localhost:8081/web/audit

The audit log page shows:
- Complete event history for all camera configurations
- What changed (field-by-field comparison)
- When changes occurred
- Previous values → New values
- Filtering by configuration ID

## Running the Service

### Using Docker Compose
```bash
# Build the service
make build

# Deploy with docker-compose
make deploy-compose

# Or manually
docker-compose up camconfig
```

The service will be available at http://localhost:8081

### Running Locally
```bash
# Build the binary
make build-go

# Run the service
./bin/camconfig
```

### Configuration

The service can be configured via:
1. `config.yaml` file
2. Environment variables

#### config.yaml
```yaml
listenAddr: ":8081"
dbUri: "inmemory://"
templateDir: "./web/templates"
```

#### Environment Variables
- `DB_URI`: Database connection string
  - `inmemory://` for in-memory storage (default)
  - `host=localhost port=5432 user=postgres dbname=eventsourcing password=pass sslmode=disable` for PostgreSQL

## Example Usage

### Creating a Configuration
```bash
curl -X POST http://localhost:8081/v1/camconfigs \
  -H "Content-Type: application/json" \
  -d '{
    "cameraId": "CAM-001",
    "name": "Front Door Camera",
    "gamma": 1.2,
    "exposure": 1500,
    "saturation": 60,
    "sharpness": 55,
    "gain": 30
  }'
```

Response:
```json
{
  "originator": {
    "id": "168d206e-48cb-43b7-bb17-35867da50c54",
    "version": 1
  },
  "cameraId": "CAM-001",
  "name": "Front Door Camera",
  "gamma": 1.2,
  "exposure": 1500,
  "saturation": 60,
  "sharpness": 55,
  "gain": 30
}
```

### Retrieving a Configuration
```bash
curl "http://localhost:8081/v1/camconfigs?id=168d206e-48cb-43b7-bb17-35867da50c54"
```

### Updating a Configuration
```bash
curl -X PUT "http://localhost:8081/v1/camconfigs?id=168d206e-48cb-43b7-bb17-35867da50c54&version=1" \
  -H "Content-Type: application/json" \
  -d '{
    "gamma": 1.5,
    "exposure": 2000
  }'
```

Response shows incremented version:
```json
{
  "originator": {
    "id": "168d206e-48cb-43b7-bb17-35867da50c54",
    "version": 2
  },
  ...
}
```

### Viewing Audit History

Visit http://localhost:8081/web/audit to see:
- All configuration changes over time
- Field-by-field change tracking
- Created/Updated/Deleted events
- Timeline visualization

## What This Example Demonstrates

### 1. Event Sourcing
Every operation (Create, Update, Delete) generates immutable events stored in the event store. The current state is derived by replaying these events.

### 2. Optimistic Locking
The `version` field provides optimistic concurrency control. Updates must specify the current version to succeed, preventing lost updates.

### 3. Audit Trail
All changes are automatically tracked in the application log, providing a complete audit trail without additional code.

### 4. CRUD Abstraction
The eskit library provides a simple CRUD interface that handles:
- UUID generation
- Version management
- Event creation and storage
- State reconstruction from events

### 5. In-Memory Storage
The example uses in-memory storage for simplicity, but can easily switch to PostgreSQL by changing the `dbUri` configuration.

## Architecture

```
┌─────────────────┐
│   Web UI/API    │
│   (HTTP Layer)  │
└────────┬────────┘
         │
┌────────▼────────┐
│  Service Layer  │
│  (Provider)     │
└────────┬────────┘
         │
┌────────▼────────┐
│  CRUD Store     │
│  (eskit lib)    │
└────────┬────────┘
         │
┌────────▼────────┐
│  Event Store    │
│  (In-Memory)    │
└─────────────────┘
```

## Project Structure

```
services/camconfig/
├── cmd/
│   └── camconfig/
│       └── main.go              # Entry point, HTTP server setup
├── provider/
│   ├── service.go               # Domain entity definition
│   ├── rest_handlers.go         # JSON API handlers
│   └── web_handlers.go          # HTML interface handlers
├── web/
│   └── templates/
│       ├── index.html           # Configuration list
│       ├── form.html            # Create/Edit form
│       └── audit.html           # Audit log viewer
├── config.yaml                  # Default configuration
├── Dockerfile                   # Container definition
└── README.md                    # This file
```

## Technologies Used

- **Go**: Service implementation
- **eskit**: Event sourcing library
- **HTMX**: Dynamic HTML updates without build tools
- **Docker**: Containerization
- **In-Memory Store**: Simple event storage (can use PostgreSQL)

## Next Steps

Try these modifications to learn more:
1. Add more camera configuration fields
2. Switch from in-memory to PostgreSQL storage
3. Implement list/pagination on the API
4. Add authentication and user tracking
5. Create a consumer to react to configuration changes
6. Add validation rules for camera settings
