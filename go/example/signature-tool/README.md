# TEENet Signature Tool

A distributed signature and multi-party voting tool based on TEENet Key Management System, providing Web UI and REST API.

## Features

- **Single-party Signature**: Sign messages using TEE key management system
- **Signature Verification**: Verify the validity of digital signatures
- **Multi-party Voting Signature**: Support M-of-N threshold voting mechanism across multiple TEE nodes
- **Multi-protocol Support**: Support for ECDSA and Schnorr protocols
- **Multi-curve Support**: Support for ED25519, SECP256K1, SECP256R1 curves
- **Web Interface**: Intuitive web interface for all operations
- **REST API**: Complete REST API for programmatic access
- **Docker Support**: Easy deployment with Docker

## Quick Start

### Using Docker (Recommended)

```bash
# Pull the pre-built image
docker load < teenet-signature-tool.tar.gz

# Or build from source
docker build -t teenet-signature-tool:latest .

# Run the container
docker run -d \
  --name signature-tool \
  -p 8080:8080 \
  -e APP_ID="your-app-id" \
  -e TEE_CONFIG_ADDR="tee-config-server:50052" \
  teenet-signature-tool:latest
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/TEENet-io/teenet-sdk.git
cd teenet-sdk/go/example/signature-tool

# Build the application
go build -o signature-tool .

# Run with environment variables
APP_ID="your-app-id" \
TEE_CONFIG_ADDR="localhost:50052" \
./signature-tool
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_ID` | Application ID for signing operations | - | Yes |
| `TEE_CONFIG_ADDR` | TEE configuration server address | `localhost:50052` | No |
| `PORT` | Web server port | `8080` | No |
| `FRONTEND_PATH` | Frontend files path | `./frontend` | No |
| `GIN_MODE` | Gin framework mode (debug/release) | `release` | No |

## API Documentation

### Health Check
```http
GET /api/health
```
Returns server health status.

### Get Configuration
```http
GET /api/config
```
Returns current configuration including App ID.

### Get Public Key
```http
POST /api/get-public-key
Content-Type: application/json

{
  "app_id": "your-app-id"
}
```

Response:
```json
{
  "success": true,
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519"
}
```

### Sign Message
```http
POST /api/sign-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!"
}
```

Response:
```json
{
  "success": true,
  "signature": "hex-encoded-signature",
  "message": "Hello, World!",
  "app_id": "your-app-id"
}
```

### Verify Signature
```http
POST /api/verify-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!",
  "signature": "hex-encoded-signature"
}
```

Response:
```json
{
  "success": true,
  "valid": true,
  "message": "Hello, World!",
  "signature": "hex-encoded-signature",
  "app_id": "your-app-id",
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519"
}
```

### Multi-party Voting Signature
```http
POST /api/vote
Content-Type: application/json

{
  "message": "base64-encoded-message",
  "signer_app_id": "app-id-requesting-signature"
}
```

**Note**: The target App IDs and required votes are automatically fetched from the server based on the VotingSign project configuration.

Response:
```json
{
  "success": true,
  "approved": true,
  "app_id": "current-app-id",
  "message": "APPROVED",
  "voting_results": {
    "voting_complete": true,
    "successful_votes": 2,
    "required_votes": 2,
    "total_targets": 2,
    "final_result": "APPROVED",
    "vote_details": [
      {
        "client_id": "app-1",
        "success": true,
        "response": true,
        "error": ""
      },
      {
        "client_id": "app-2",
        "success": true,
        "response": true,
        "error": ""
      }
    ]
  },
  "signature": "hex-encoded-signature",
  "timestamp": "2025-09-09T10:00:00Z"
}
```

## Multi-party Voting Mechanism

### How It Works

The voting mechanism implements a distributed consensus system where multiple TEE nodes participate in approving a signature request:

1. **Voting Initiation**: Client sends a voting request to the signature tool
2. **Vote Distribution**: The tool distributes voting requests to all configured target nodes
3. **Local Decision**: Each node makes a voting decision based on custom logic
4. **Vote Collection**: The system waits for all votes to be collected
5. **Threshold Check**: Verifies if the required number of approvals is met
6. **Signature Generation**: If approved, generates the final signature

### Voting Flow Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │  Signature Tool │    │ TEE DAO Client  │
│                 │    │     Backend     │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ 1. POST /api/vote    │                      │
          ├─────────────────────►│                      │
          │ {                    │                      │
          │   message,           │                      │
          │   signer_app_id      │                      │
          │ }                    │                      │
          │                      │                      │
          │                      │ 2. Extract message   │
          │                      │ 3. Local approval    │
          │                      │    decision          │
          │                      │                      │
          │                      │ 4. client.Sign()     │
          │                      ├─────────────────────►│
          │                      │ {                    │
          │                      │   EnableVoting: true │
          │                      │   LocalApproval: ... │
          │                      │   HTTPRequest: ...   │
          │                      │ }                    │
          │                      │                      │
          │                      │                      │ 5. Fetch targets from
          │                      │                      │    server config
          │                      │                      │
          │                      │                      │ 6. Concurrent voting
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID 1 │
          │                      │                      │ │ (Local vote)    │
          │                      │                      │ └─────────────────┘
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID 2 │
          │                      │                      │ │ (HTTP request)  │
          │                      │                      │ └─────────────────┘
          │                      │                      │ ┌─────────────────┐
          │                      │                      │ │                 │
          │                      │                      ├─┤ Target App ID N │
          │                      │                      │ │ (HTTP request)  │
          │                      │                      │ └─────────────────┘
          │                      │                      │
          │                      │                      │ 7. Collect all votes
          │                      │                      │    (wait for all)
          │                      │                      │
          │                      │                      │ 8. Check threshold
          │                      │                      │    (M of N)
          │                      │                      │
          │                      │ 9. SignResult with   │ 10. If approved,
          │                      │    VotingInfo        │     generate signature
          │                      │◄─────────────────────┤
          │                      │                      │
          │ 11. Final response   │                      │
          │ {                    │                      │
          │   success,           │                      │
          │   approved,          │                      │
          │   voting_results: {  │                      │
          │     vote_details[]   │                      │
          │   },                 │                      │
          │   signature          │                      │
          │ }                    │                      │
          │◄─────────────────────┤                      │
          │                      │                      │
```

### Default Voting Logic

The default implementation approves votes if the message contains the word "test" (case-insensitive). This can be customized by modifying the voting handler in the code.

## Project Structure

```
signature-tool/
├── main.go         # Main application and HTTP routes
├── types.go        # Data structure definitions
├── crypto.go       # Cryptographic operations
├── server.go       # Static file server
├── voting.go       # Voting logic handler
├── go.mod          # Go module configuration
├── go.sum          # Dependency lock file
├── Dockerfile      # Docker build configuration
├── README.md       # This file
└── frontend/       # Web UI files
    ├── index.html  # Main HTML page
    ├── styles.css  # CSS styles
    └── app.js      # JavaScript application
```

## Architecture

### Components

1. **HTTP Server**: Gin-based web server handling API requests
2. **TEE Client**: Interface with TEE key management system
3. **Voting Service**: Handles multi-party voting logic
4. **Frontend**: Web UI for user interaction

### Key Technologies

- **Go 1.24+**: Main programming language
- **Gin Web Framework**: HTTP server and routing
- **gRPC**: Communication with TEE services
- **Docker**: Containerization for deployment

## Security Considerations

1. **Environment Variables**: Sensitive configuration stored in environment variables
2. **Input Validation**: All inputs are validated before processing
3. **Error Handling**: Error messages don't expose sensitive information
4. **TLS Support**: Communication with TEE services uses TLS
5. **CORS**: Properly configured for web access

## Development

### Prerequisites

- Go 1.24 or higher
- Access to TEE configuration server
- Valid App ID configuration

### Building

```bash
# Clone repository
git clone https://github.com/TEENet-io/teenet-sdk.git
cd teenet-sdk/go/example/signature-tool

# Install dependencies
go mod download

# Build application
go build -o signature-tool .

# Run tests
go test ./...
```

### Custom Voting Logic

To customize the voting logic, modify the voting handler in `main.go`:

```go
// Current logic: approve if message contains "test"
localApproval := strings.Contains(strings.ToLower(messageStr), "test")

// Custom examples:
// - Based on message hash
// - Based on timestamp
// - External API validation
// - Complex business rules
```

## Deployment

### Production Deployment

1. **Environment Setup**
   - Ensure TEE configuration server is accessible
   - Configure valid App ID
   - Set up network connectivity

2. **Docker Deployment**
   ```bash
   docker run -d \
     --restart=always \
     --name signature-tool \
     -p 8080:8080 \
     -e APP_ID="production-app-id" \
     -e TEE_CONFIG_ADDR="tee-server:50052" \
     -e GIN_MODE="release" \
     teenet-signature-tool:latest
   ```

3. **Health Monitoring**
   - Use `/api/health` endpoint for health checks
   - Monitor container logs for errors
   - Set up alerts for failures

### Scaling Considerations

- The tool can be deployed across multiple instances
- Each instance requires its own App ID
- Load balancing can be configured for high availability

## Troubleshooting

### Common Issues

1. **"APP_ID environment variable is required"**
   - Solution: Set the `APP_ID` environment variable

2. **"Failed to initialize TEE client"**
   - Check TEE configuration server connectivity
   - Verify server address and port
   - Check network firewall rules

3. **Voting request failures**
   - Verify target applications are running
   - Check network connectivity
   - Review deployment configuration

4. **Signature verification failures**
   - Ensure correct message format
   - Verify App ID matches
   - Check signature encoding (hex)

### Logging

The application logs important events including:
- Server startup information
- API request processing
- Voting flow details
- Error conditions
- Signature operations

## Updates and Versioning

### Latest Updates (v2.0)

- **New Sign API**: Unified signing interface with `Sign()` method
- **SignRequest Structure**: New request structure supporting both regular and voting signatures
- **Enhanced Voting**: Improved voting mechanism with detailed result tracking
- **Better Error Handling**: More informative error messages and status codes
- **Docker Optimization**: Smaller image size with multi-stage build

### Version History

- v2.0 - Unified Sign API with enhanced voting mechanism
- v1.5 - Automatic voting target configuration from server
- v1.0 - Initial release with basic signing and voting

## Support

For issues, questions, or contributions:
- GitHub Issues: [https://github.com/TEENet-io/teenet-sdk/issues](https://github.com/TEENet-io/teenet-sdk/issues)
- Documentation: [TEENet Documentation](https://docs.teenet.io)

## License

Copyright (c) 2025 TEENet Technology (Hong Kong) Limited. All Rights Reserved.

This software is proprietary and confidential. Use is subject to license terms.