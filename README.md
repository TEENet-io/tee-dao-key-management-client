# TEE DAO Key Management Client & Test Environment

A comprehensive testing environment for TEE DAO key management operations, including multi-language client libraries and a complete mock server infrastructure for local development and testing.

## 🚀 Components

### 1. Client Libraries
- **Go** - Production-ready implementation
- **TypeScript** - Node.js compatible implementation

### 2. Mock Server Environment
- **Mock DAO Server** - Simulates distributed key management operations with real cryptography
- **Mock Config Server** - Provides node discovery and configuration
- **Mock App Node** - Simulates user management system functionality

### 3. TEENet Signature Tool Example
- **Web-based Signature Tool** - Interactive web interface and REST API for digital signature operations
- **Multi-Protocol Support** - ECDSA and Schnorr protocols with ED25519, SECP256K1, and SECP256R1 curves
- **Signature Verification** - Comprehensive verification capabilities with multiple format support
- **Docker Ready** - Containerized deployment for easy integration

## ✨ Mock Server Features

The mock server provides a complete local testing environment with:

- **Cryptographic Signatures**:
  - ECDSA (secp256k1, secp256r1) 
  - Schnorr (ed25519, secp256k1)
- **Semantic App IDs**: 
  - `secure-messaging-app` (Schnorr + ED25519)
  - `financial-trading-platform` (ECDSA + SECP256R1)
  - `digital-identity-service` (Schnorr + SECP256K1)
  - `bitcoin-wallet-app` (ECDSA + SECP256K1)
- **TLS Security**: Mutual certificate authentication
- **Consistent Testing**: Deterministic key generation for reproducible results

## 🏁 Quick Start

### Start Mock Server Environment

```bash
cd mock-server
./start-test-env.sh
```

This will start:
- Config Server on localhost:50052
- DAO Server on localhost:50051  
- App Node on localhost:50053

### Run Client Examples

**Go Example:**
```bash
cd go
go run example/main.go
```

**TypeScript Example:**
```bash
cd typescript
npm install
npm run example
```

### Stop Mock Server

```bash
cd mock-server
./stop-test-env.sh
```

## 🔧 TEENet Signature Tool Example

The signature tool provides a comprehensive web-based interface and REST API for digital signature operations within the TEE DAO ecosystem.

### Features

- **Web Interface**: Interactive HTML interface for signature operations
- **REST API**: Complete API endpoints for programmatic access
- **Multi-Protocol Support**: ECDSA and Schnorr signature protocols
- **Multiple Curves**: Support for ED25519, SECP256K1, and SECP256R1 curves
- **Signature Verification**: Comprehensive verification with multiple format support
- **Docker Ready**: Containerized deployment for easy integration

### Quick Start

**Start the Signature Tool:**

```bash
cd go/example/signature-tool
go run main.go
```

The web interface will be available at `http://localhost:8080`

**Docker Deployment:**

```bash
cd go/example/signature-tool
docker build -t signature-tool .
docker run -p 8080:8080 -e TEE_CONFIG_ADDR=host.docker.internal:50052 signature-tool
```

### API Endpoints

- `GET /api/health` - Health check
- `POST /api/get-public-key` - Get public key by App ID
- `POST /api/sign-with-appid` - Sign message using App ID
- `POST /api/verify-with-appid` - Verify signature using App ID
- `POST /api/sign` - Advanced signing with explicit protocol/curve
- `POST /api/verify` - Advanced verification with explicit protocol/curve

### Configuration

Environment variables:
- `APP_ID`: Application ID for signature operations
- `TEE_CONFIG_ADDR`: TEE configuration server address (default: localhost:50052)
- `PORT`: Web server port (default: 8080)

For detailed documentation, see [go/example/signature-tool/README.md](go/example/signature-tool/README.md)

## Features

- **Secure Message Signing**: Sign messages using distributed cryptographic keys
- **AppID Service Integration**: Get public keys and sign messages using AppID
- **Multiple Protocols**: Support for ECDSA and Schnorr signature protocols
- **Multiple Curves**: Support for ED25519, SECP256K1, and SECP256R1 curves
- **TLS Security**: Secure communication with TEE nodes using mutual TLS authentication
- **Simple API**: Easy-to-use client interface with automatic configuration
- **Multi-language Support**: Go and TypeScript implementations with identical APIs

## Go Implementation

### Installation

```bash
go get github.com/TEENet-io/tee-dao-key-management-client/go
```

### Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    client "github.com/TEENet-io/tee-dao-key-management-client/go"
)

func main() {
    // Create client with config server address
    client := client.NewClient("localhost:50052")
    defer client.Close()

    // Initialize client (fetch config + establish TLS connection)
    if err := client.Init(); err != nil {
        log.Fatalf("Initialization failed: %v", err)
    }

    fmt.Printf("Client connected, Node ID: %d\n", client.GetNodeID())

    // Example 1: Get public key by app ID
    appID := "xxxxxxx"
    publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID)
    if err != nil {
        log.Printf("Failed to get public key by app ID: %v", err)
    } else {
        fmt.Printf("Public key for app ID %s:\n", appID)
        fmt.Printf("  - Protocol: %s\n", protocol)
        fmt.Printf("  - Curve: %s\n", curve)
        fmt.Printf("  - Public Key: %s\n", publicKey)
    }

    // Example 2: Sign message with app ID
    message := []byte("Hello from AppID Service!")
    signature, err := client.SignWithAppID(message, appID)
    if err != nil {
        log.Printf("Signing with app ID failed: %v", err)
    } else {
        fmt.Printf("Signing with app ID successful!\n")
        fmt.Printf("Message: %s\n", string(message))
        fmt.Printf("Signature: %x\n", signature)
    }

    // Example 3: Traditional signing with explicit protocol and curve
    publicKeyBytes := []byte("example-public-key-from-dkg-service") // From external DKG service
    message2 := []byte("Hello, TEE DAO!")
    
    signature2, err := client.Sign(message2, publicKeyBytes, 1, 1) // ECDSA, ED25519
    if err != nil {
        log.Fatalf("Signing failed: %v", err)
    }
    fmt.Printf("Traditional signing successful!\n")
    fmt.Printf("Message: %s\n", string(message2))
    fmt.Printf("Signature: %x\n", signature2)
}
```

### Running the Go Example

```bash
cd go
go run example/main.go
```

## TypeScript Implementation

### Installation

```bash
cd typescript
npm install
```

### Quick Start

```typescript
import { Client } from './src/client';

async function main() {
  // Create client with config server address
  const client = new Client('localhost:50052');

  try {
    // Initialize client (fetch config + establish TLS connection)
    await client.init();
    console.log(`Client connected, Node ID: ${client.getNodeId()}`);

    // Example 1: Get public key by app ID
    const appID = 'xxxxxxx';
    try {
      const { publickey, protocol, curve } = await client.getPublicKeyByAppID(appID);
      console.log(`Public key for app ID ${appID}:`);
      console.log(`  - Protocol: ${protocol}`);
      console.log(`  - Curve: ${curve}`);
      console.log(`  - Public Key: ${publickey}`);
    } catch (error) {
      console.error(`Failed to get public key by app ID: ${error}`);
    }

    // Example 2: Sign message with app ID
    const message = new TextEncoder().encode('Hello from AppID Service!');
    try {
      const signature = await client.signWithAppID(message, appID);
      console.log('Signing with app ID successful!');
      console.log(`Message: ${new TextDecoder().decode(message)}`);
      console.log(`Signature: ${Buffer.from(signature).toString('hex')}`);
    } catch (error) {
      console.error(`Signing with app ID failed: ${error}`);
    }

    // Example 3: Traditional signing with explicit protocol and curve
    const publicKey = new TextEncoder().encode('example-public-key-from-dkg-service'); // From external DKG service
    const message2 = new TextEncoder().encode('Hello, TEE DAO!');
    
    const signature2 = await client.sign(message2, publicKey, 1, 1); // ECDSA, ED25519
    console.log('Traditional signing successful!');
    console.log(`Message: ${new TextDecoder().decode(message2)}`);
    console.log(`Signature: ${Buffer.from(signature2).toString('hex')}`);

  } catch (error) {
    console.error('Error:', error);
  } finally {
    await client.close();
  }
}

main();
```

### Running the TypeScript Example

```bash
cd typescript
npm run example  # Build and run with no warnings
# or
npm run build && node dist/example.js
```

## API Reference

Both Go and TypeScript implementations provide identical functionality:

### Client Creation and Initialization

**Go:**
```go
client := client.NewClient("localhost:50052")
err := client.Init()
```

**TypeScript:**
```typescript
const client = new Client('localhost:50052');
await client.init();
```

### AppID Service Methods

**Get Public Key by AppID:**
```go
publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID)
```

**TypeScript:**
```typescript
const { publickey, protocol, curve } = await client.getPublicKeyByAppID(appId)
```

**Sign with AppID:**
```go
signature, err := client.SignWithAppID(message, appID)
```

**TypeScript:**
```typescript
const signature = await client.signWithAppID(message, appId)
```

### Traditional Message Signing

**Go:**
```go
signature, err := client.Sign(message, publicKey, protocol, curve)
```

**TypeScript:**
```typescript
const signature = await client.sign(message, publicKey, protocol, curve)
```

#### Protocol Constants

**Go:**
- `constants.ProtocolECDSA` (1)
- `constants.ProtocolSchnorr` (2)

**TypeScript:**
- `Protocol.ECDSA` (1)
- `Protocol.SCHNORR` (2)

#### Curve Constants

**Go:**
- `constants.CurveED25519` (1)
- `constants.CurveSECP256K1` (2)
- `constants.CurveSECP256R1` (3)

**TypeScript:**
- `Curve.ED25519` (1)
- `Curve.SECP256K1` (2)
- `Curve.SECP256R1` (3)

### Utility Methods

#### Get Node ID
**Go:** `nodeID := client.GetNodeID()`  
**TypeScript:** `const nodeId = client.getNodeId()`

#### Set Timeout
**Go:** `client.SetTimeout(30 * time.Second)`  
**TypeScript:** `client.setTimeout(30000)`

#### Close Connection
**Go:** `client.Close()`  
**TypeScript:** `await client.close()`

## Architecture

Both implementations consist of three main components:

- **Config Client**: Handles communication with the configuration server to retrieve node configuration
- **Task Client**: Manages secure communication with TEE nodes for cryptographic operations
- **AppID Client**: Manages communication with the AppID service for user management operations

The client workflow:
1. Initialize client with config server address
2. Fetch node configuration (certificates, keys, target address)
3. Establish secure TLS connection to TEE node
4. Establish secure TLS connection to AppID service
5. Perform signing operations with specified protocol and curve or using AppID

## Protocol Buffers

The clients use gRPC with Protocol Buffers for communication:
- `proto/node_management/`: Node management services for configuration
- `proto/key_management/user_task.proto`: User task definitions for signing operations
- `proto/appid/appid_service.proto`: AppID service definitions for user management

## Requirements

### Go
- Go 1.24.2 or later
- gRPC and Protocol Buffers support

### TypeScript
- Node.js 18.0.0 or later
- npm or yarn

## Project Structure

```
├── go/                     # Go client implementation
│   ├── client.go          # Main client
│   ├── pkg/               # Core packages
│   │   ├── config/        # Configuration client
│   │   ├── constants/     # Protocol and curve constants
│   │   ├── task/          # Task client for signing
│   │   └── usermgmt/      # User management client
│   ├── example/           # Go examples
│   │   ├── main.go        # Basic client example
│   │   └── signature-tool/ # Web-based signature tool
│   │       ├── main.go    # Signature tool web application
│   │       ├── README.md  # Signature tool documentation
│   │       └── Dockerfile # Docker configuration
│   └── proto/             # Generated Go protobuf files
├── typescript/            # TypeScript client implementation
│   ├── src/               # TypeScript source code
│   │   ├── client.ts      # Main client
│   │   ├── config-client.ts # Configuration client
│   │   ├── task-client.ts # Task client for signing
│   │   ├── appid-client.ts # AppID client for user management
│   │   ├── types.ts       # TypeScript types and constants
│   │   └── example.ts     # TypeScript example
│   ├── proto/             # Protobuf definitions
│   └── dist/              # Compiled JavaScript
└── mock-server/           # Complete mock server environment
    ├── dao-server.go      # Mock DAO server with real cryptography
    ├── mock-config-server.go # Mock configuration server
    ├── mock-app-node.go   # Mock app node/user management
    ├── proto/             # Protocol buffer definitions
    ├── certs/             # TLS certificates (auto-generated)
    ├── logs/              # Server logs
    ├── start-test-env.sh  # Start all services
    ├── stop-test-env.sh   # Stop all services
    └── README.md          # Detailed mock server documentation
```

## Examples

- **Go Client**: See [go/example/main.go](go/example/main.go)
- **TypeScript Client**: See [typescript/src/example.ts](typescript/src/example.ts)
- **TEENet Signature Tool Example**: See [go/example/signature-tool/](go/example/signature-tool/) for web-based signature operations
- **Mock Server**: See [mock-server/README.md](mock-server/README.md) for detailed documentation

### Complete Testing Workflow

1. **Start Mock Environment:**
   ```bash
   cd mock-server
   ./start-test-env.sh
   ```

2. **Run Client Examples:**
   ```bash
   # Go client
   cd go && go run example/main.go
   
   # TypeScript client  
   cd typescript && npm run example
   ```

3. **View Server Logs:**
   ```bash
   tail -f mock-server/logs/*.log
   ```

4. **Stop Environment:**
   ```bash
   cd mock-server
   ./stop-test-env.sh
   ```

## Security Notes

- TLS mutual authentication is enabled for all communications
- Hostname verification is maintained (never disabled)
- Certificate and key files are excluded from git via .gitignore
- No hardcoded credentials or secrets

## License

This project is part of the TEENet ecosystem for secure distributed key management.