# TEENet Signature Tool

A comprehensive signature tool for TEENet key management system that provides both web interface and REST API for digital signature operations.

## Overview

The TEENet Signature Tool is a Go-based web application that provides digital signature functionality using the TEENet key management client. It supports multiple cryptographic protocols and curves, offering both a user-friendly web interface and a comprehensive REST API.

## Features

- **Web Interface**: Interactive HTML interface for easy signature operations
- **REST API**: Complete API for programmatic access
- **Multiple Protocols**: Support for ECDSA and Schnorr signature protocols
- **Multiple Curves**: Support for ED25519, SECP256K1, and SECP256R1 curves
- **App ID Integration**: Simplified operations using App ID instead of raw public keys
- **Signature Verification**: Comprehensive signature verification capabilities
- **Docker Support**: Containerized deployment with Docker

## Prerequisites

- Go 1.24 or higher
- TEENet TEE Configuration Server running on `localhost:50052` (default)
- Valid App ID configured in environment

## Installation

### Local Development

1. **Clone the repository and navigate to the signature tool directory:**
   ```bash
   cd go/example/signature-tool
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set required environment variables:**
   ```bash
   export APP_ID="your-app-id-here"
   export TEE_CONFIG_ADDR="localhost:50052"
   export PORT="8080"
   ```

4. **Build and run:**
   ```bash
   go build -o signature-tool main.go
   ./signature-tool
   ```

### Docker Deployment

#### Build and Run Locally

1. **Build the Docker image:**
   ```bash
   docker build -t teenet-signature-tool .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name signature-tool \
     -p 8080:8080 \
     -e APP_ID="your-app-id-here" \
     -e TEE_CONFIG_ADDR="localhost:50052" \
     -e PORT="8080" \
     teenet-signature-tool
   ```

#### Build and Compress Docker Image

To build and compress the Docker image for distribution:

1. **Use the build script:**
   ```bash
   ./build-image.sh
   ```

2. **The script will:**
   - Build the Docker image as `teenet-signature-tool:latest`
   - Compress it to `teenet-signature-tool-latest.tar.gz`
   - Display file size and usage instructions

3. **To load the compressed image elsewhere:**
   ```bash
   docker load < teenet-signature-tool-latest.tar.gz
   ```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_ID` | Application ID for signature operations | - | Yes |
| `TEE_CONFIG_ADDR` | TEE configuration server address | `localhost:50052` | No |
| `PORT` | Web server port | `8080` | No |

## Usage

### Web Interface

Once the server is running, access the web interface at:
```
http://localhost:8080
```

The web interface provides the following functions:

1. **Sign Message**: Sign a message using the configured App ID
2. **Verify Signature**: Verify a signature using the App ID
3. **Get Public Key**: Retrieve the public key associated with the App ID
4. **Advanced Sign**: Sign with custom public key, protocol, and curve
5. **Advanced Verify**: Verify with custom public key, protocol, and curve

### REST API

The tool provides a comprehensive REST API with the following endpoints:

#### Health Check
```http
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "TEENet Signature Tool",
  "node_id": 123
}
```

#### Get Public Key by App ID
```http
POST /api/get-public-key
Content-Type: application/json

{
  "app_id": "your-app-id"
}
```

**Response:**
```json
{
  "success": true,
  "app_id": "your-app-id",
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519"
}
```

#### Sign Message with App ID
```http
POST /api/sign-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Hello, World!",
  "app_id": "your-app-id",
  "signature": "hex-encoded-signature"
}
```

#### Verify Signature with App ID
```http
POST /api/verify-with-appid
Content-Type: application/json

{
  "app_id": "your-app-id",
  "message": "Hello, World!",
  "signature": "hex-encoded-signature"
}
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "message": "Hello, World!",
  "signature": "hex-encoded-signature",
  "app_id": "your-app-id",
  "protocol": "schnorr",
  "curve": "ed25519"
}
```

#### Sign Message with Public Key (Advanced)
```http
POST /api/sign
Content-Type: application/json

{
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519",
  "message": "Hello, World!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Hello, World!",
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519",
  "signature": "hex-encoded-signature"
}
```

#### Verify Signature with Public Key (Advanced)
```http
POST /api/verify
Content-Type: application/json

{
  "public_key": "base64-encoded-public-key",
  "protocol": "schnorr",
  "curve": "ed25519",
  "message": "Hello, World!",
  "signature": "hex-encoded-signature"
}
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "message": "Hello, World!",
  "public_key": "base64-encoded-public-key",
  "signature": "hex-encoded-signature",
  "protocol": "schnorr",
  "curve": "ed25519"
}
```

## Supported Protocols and Curves

### Protocols
- `ecdsa`: Elliptic Curve Digital Signature Algorithm
- `schnorr`: Schnorr signature scheme

### Curves
- `ed25519`: Edwards25519 curve
- `secp256k1`: SECP256K1 curve (Bitcoin curve)
- `secp256r1`: SECP256R1 curve (NIST P-256)

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "success": false,
  "error": "Error description"
}
```

Common error scenarios:
- Invalid request format
- Missing required fields
- Invalid public key format (must be base64)
- Invalid signature format (must be hex)
- Unsupported protocol or curve
- TEE server communication errors

## Security Considerations

1. **Environment Variables**: Store sensitive configuration in environment variables
2. **Network Security**: Ensure TEE configuration server is properly secured
3. **Input Validation**: All inputs are validated before processing
4. **Error Messages**: Error messages do not expose sensitive information
5. **CORS**: CORS is enabled for web interface access

## Development

### Project Structure
```
signature-tool/
├── main.go          # Main application file
├── go.mod           # Go module file
├── go.sum           # Go module checksums
├── Dockerfile       # Docker configuration
├── build-image.sh   # Docker image build and compression script
└── README.md        # This file
```

### Dependencies
- `github.com/TEENet-io/tee-dao-key-management-client/go`: TEENet key management client
- `github.com/gin-gonic/gin`: Web framework
- Standard Go crypto libraries

### Building
```bash
go build -o signature-tool main.go
```

### Testing
The tool includes comprehensive error handling and validation. Test the API endpoints using curl or the web interface.

## Troubleshooting

### Common Issues

1. **"APP_ID environment variable is required"**
   - Solution: Set the `APP_ID` environment variable

2. **"Failed to initialize TEE client"**
   - Solution: Ensure the TEE configuration server is running and accessible

3. **"Invalid public key format"**
   - Solution: Ensure public keys are base64 encoded

4. **"Invalid signature format"**
   - Solution: Ensure signatures are hex encoded

### Logs
The application logs important events including:
- Server startup information
- API request processing
- Error conditions
- Signature operation results

## License

This project is part of the TEENet key management client and follows the same license terms.

## Contributing

When contributing to this tool:
1. Follow Go coding standards
2. Add appropriate error handling
3. Update documentation for new features
4. Test thoroughly before submitting changes 