# Simple Data Server

A lightweight, file-based data storage server with REST API and WebSocket support. Originally designed as a faster alternative to Resonite's CloudVars.

## ⚠️ SECURITY NOTICE
This project stores data as plaintext files on the server host. If you are concerned about security, please do not store unencrypted sensitive data on the server.

## Features

- **File-based Storage**: No database required - data stored as JSON files
- **Multi-tenant**: Separate data namespaces using BasicAuth credentials  
- **REST API**: Standard HTTP methods (GET, POST, DELETE)
- **WebSocket Support**: Real-time communication and live updates
- **Hierarchical Data**: Organized using `object/table/key` structure
- **Health Monitoring**: Built-in health check endpoint
- **Docker Support**: Easy deployment with Docker and docker-compose
- **Configurable**: Environment-based configuration
- **Thread-safe**: File locking prevents concurrent access issues
- **Input Validation**: Prevents directory traversal and validates input

## Quick Start

### Using Go
```bash
# Clone and run
git clone https://github.com/Merith-TK/simple-data-server
cd simple-data-server
go run .
```

### Using Docker
```bash
# Build and run with docker-compose
docker-compose up -d
```

### Configuration
Copy `.env.example` to `.env` and modify as needed:
```bash
PORT=8080
DATA_DIR=./data
LOG_LEVEL=info
```

## API Endpoints

### REST API
- `GET /api/{object}/{table}` - Returns entire table as JSON
- `GET /api/{object}/{table}/{key}` - Returns specific key value as text
- `POST /api/{object}/{table}/{key}` - Sets key to value from request body
- `DELETE /api/{object}/{table}/{key}` - Removes key from table
- `GET /health` - Health check endpoint

### WebSocket
- `GET /api/{object}/{table}/ws` - WebSocket endpoint for real-time operations

#### WebSocket Commands
- `set key value` - Store data (broadcasts UPDATE to all connected clients)
- `get key` - Retrieve data (returns `key: value`)
- `del key` - Delete data (broadcasts DELETED to all connected clients)
- `exit` - Close connection

#### WebSocket Events
- `uuid: {hash}` - Sent on connection with user identifier
- `UPDATE: key: value` - Broadcast when data is updated
- `DELETED: key` - Broadcast when data is deleted

## Authentication

### BasicAuth (Optional)
Supply BasicAuth credentials to create a separate data namespace:
- **With auth**: Data stored in `./data/{userhash}/{object}/{table}.json`
- **Without auth**: Data stored in `./data/default/{object}/{table}.json`

User hash is SHA256 of `username + password`.

## Data Structure

```
./data/
├── default/           # No auth users
│   └── {object}/
│       └── {table}.json
└── {userhash}/        # Authenticated users
    └── {object}/
        └── {table}.json
```

Each JSON file contains:
```json
{
  "data": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

## Examples

### REST API
```bash
# Set a value
curl -X POST http://localhost:8080/api/game/players/player1 -d "John Doe"

# Get a value  
curl http://localhost:8080/api/game/players/player1
# Returns: John Doe

# Get all values in table
curl http://localhost:8080/api/game/players
# Returns: {"data":{"player1":"John Doe"}}

# Delete a value
curl -X DELETE http://localhost:8080/api/game/players/player1

# With BasicAuth
curl -u username:password -X POST http://localhost:8080/api/game/players/player1 -d "Jane Doe"
```

### WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/api/game/players/ws');

ws.onopen = () => {
    ws.send('set player1 John Doe');
    ws.send('get player1');
    ws.send('del player1');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};
```

## Improvements Made

- ✅ **Better Error Handling**: Comprehensive error responses with proper HTTP status codes
- ✅ **Input Validation**: Prevents directory traversal and validates object/table/key names
- ✅ **Configuration**: Environment-based configuration with sensible defaults
- ✅ **Thread Safety**: File locking prevents concurrent access issues
- ✅ **Middleware**: Logging and CORS support
- ✅ **Health Checks**: Built-in endpoint for monitoring
- ✅ **Docker Support**: Easy deployment with containers
- ✅ **Better WebSocket Handling**: Improved connection management and error handling
- ✅ **Structured Responses**: JSON responses for better API consistency

## Development

```bash
# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Build binary
go build -o simple-data-server .

# Run tests (if any)
go test ./...
```

## License

This project is open source. Please check the repository for license details.
