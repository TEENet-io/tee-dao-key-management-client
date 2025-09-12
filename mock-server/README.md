# TEE DAO Mock Server

A local testing environment for TEE DAO distributed key management system that simulates complete DAO nodes, configuration servers, and user management systems, allowing developers to test their programs locally without connecting to real DAO networks.

## 🚀 Quick Start

### 1. Start Services

```bash
# Start all services with one command (config server, DAO server, app node)
./start-test-env.sh
```

After successful startup, you'll see:
```
=======================================
   Test Environment Ready!
=======================================

Service Status:
  Config Server: localhost:50052 (PID: xxxx)
  DAO Server:    localhost:50051 (PID: xxxx)
  App Node:      localhost:50053 (PID: xxxx)
```

### 2. View Available App ID List

After starting services, the App node will print all available App IDs to the console:

```
Available App IDs for testing:
  - secure-messaging-app (schnorr + ed25519) - Secure Messaging Application - Schnorr/ED25519
  - financial-trading-platform (ecdsa + secp256r1) - Financial Trading Platform - ECDSA/SECP256R1
  - digital-identity-service (schnorr + secp256k1) - Digital Identity Service - Schnorr/SECP256K1
  - bitcoin-wallet-app (ecdsa + secp256k1) - Bitcoin Wallet - ECDSA/SECP256K1

💡 Usage Tips:
   Copy any of the above App IDs to use in your client programs
   Each App ID corresponds to different signature protocol and curve combinations
```

Or check the App node logs:

```bash
tail -f logs/app-node.log
```

### 3. Run Example Program

```bash
# Run example program
./example-program
```

### 4. Stop Services

```bash
# Stop all services
./stop-test-env.sh
```

## 🔧 Core Features

### Config Server (localhost:50052)
- **Node Discovery**: Provides DAO node and App node address information
- **Certificate Distribution**: Provides TLS certificates required for client connections
- **Configuration Management**: Returns node configuration and network topology information

### DAO Server (localhost:50051) 
- **Real Cryptographic Signatures**: Supports multiple signature protocols and curves with real cryptography
  - ECDSA (secp256k1, secp256r1)
  - Schnorr (ed25519, secp256k1)
- **TLS Security**: Mutual certificate authentication
- **Consistent Key Generation**: Deterministic key generation for reproducible testing

### App Node (localhost:50053)
- **App ID Management**: Retrieve real public keys by App ID
- **User Management**: Simulates user management system functionality
- **Real Public Key Mapping**: Pre-configured semantic App IDs with real cryptographic key pairs
- **Protocol Support**: Supports different cryptographic protocol combinations

## 📝 Usage Examples

### Basic Signature Operations

```go
package main

import (
    "fmt"
    "log"
    client "github.com/TEENet-io/teenet-sdk/go"
)

func main() {
    // Connect to config server
    client := client.NewClient("localhost:50052")
    defer client.Close()
    
    // Initialize client
    if err := client.Init(); err != nil {
        log.Fatal(err)
    }
    
    // Perform signature operation
    message := []byte("Hello TEE DAO!")
    publicKey := []byte{...} // Your public key
    signature, err := client.Sign(message, publicKey, 2, 1) // Schnorr + ED25519
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Signature successful: %x\n", signature)
}
```

### App ID Business Flow

```go
// Get public key by App ID (use any App ID from above)
publicKey, protocol, curve, err := client.GetPublicKeyByAppID("financial-trading-platform")
if err != nil {
    log.Fatal(err)
}

// Sign directly with App ID (if client library supports it)
signature, err := client.SignWithAppID(message, "financial-trading-platform")
if err != nil {
    log.Fatal(err)
}
```

## 🧪 Test App IDs

The system provides semantic App IDs with real cryptographic keys. View available App IDs:

### Method 1: Check Startup Output
When starting services, the App node will directly display all available App IDs in the console

### Method 2: Check App Node Logs
```bash
tail -f logs/app-node.log
```

### Available App ID Types:

| App ID | Protocol | Curve | Description |
|--------|----------|-------|-------------|
| `secure-messaging-app` | Schnorr | ed25519 | Secure Messaging Application - Schnorr/ED25519 |
| `financial-trading-platform` | ECDSA | secp256r1 | Financial Trading Platform - ECDSA/SECP256R1 |
| `digital-identity-service` | Schnorr | secp256k1 | Digital Identity Service - Schnorr/SECP256K1 |
| `bitcoin-wallet-app` | ECDSA | secp256k1 | Bitcoin Wallet - ECDSA/SECP256K1 |

> **Note**: All App IDs use real cryptographic keys that are deterministically generated for consistent testing.
> 
> 💡 **Usage Suggestion**: Copy complete App IDs directly from console output to use in your client programs.

## 🔒 Security Features

- **Dynamic Certificate Generation**: TLS certificates are regenerated on each startup, not stored in version control
- **Mutual Authentication**: All services use mutual TLS certificate verification
- **CA Verification**: Both clients and servers verify each other's certificate chains
- **Encrypted Communication**: All gRPC communication is encrypted via TLS

## 📂 File Structure

```
tee-dao-mock-server/
├── dao-server.go               # DAO server main program
├── mock-config-server.go       # Config server
├── mock-app-node.go           # App node server
├── example-user-program.go    # User program example
├── proto/                     # Protocol Buffers definitions
│   ├── *.proto               # gRPC service definitions
│   └── *.pb.go               # Generated Go code
├── certs/                    # TLS certificate directory (dynamically generated)
├── logs/                     # Service logs directory
├── start-test-env.sh         # Startup script
├── stop-test-env.sh          # Stop script
├── generate-certs.sh         # Certificate generation script
├── Makefile                  # Build configuration
├── go.mod                    # Go module definition
└── README.md                # This documentation
```

## 🛠️ Development Commands

```bash
# Build all components
make build

# Quick start test environment
make start

# Run example program
make example

# Generate Protocol Buffers code
make proto

# Generate TLS certificates
make certs

# Clean build files
make clean

# View service logs
tail -f logs/*.log
```

## ⚠️ Important Notes

1. **Development Testing Only**: This is a mock environment, generated signatures are for testing purposes only
2. **Certificate Security**: TLS certificates are self-signed, suitable for local testing only
3. **Data Persistence**: All data is in memory, resets after restart
4. **Network Configuration**: Ensure ports 50051, 50052, 50053 are not occupied

## 🔧 Troubleshooting

### Service Startup Failure
```bash
# Check port usage
lsof -i :50051
lsof -i :50052  
lsof -i :50053

# Stop all services
./stop-test-env.sh
```

### Certificate Issues
```bash
# Regenerate certificates
./generate-certs.sh

# Check certificate validity
openssl x509 -in certs/dao-server.crt -text -noout
```

### Connection Issues
```bash
# Check service status
ps aux | grep -E "(dao-server|config-server|app-node)"

# View service logs
tail -f logs/dao-server.log
tail -f logs/config-server.log
tail -f logs/app-node.log
```

## 📞 Support

If you encounter issues, please check:
1. All dependencies are correctly installed (Go 1.19+, Protocol Buffers)
2. Ports are not occupied by other programs
3. Firewall settings are not blocking local connections
4. Error information in log files

---

**TEE DAO Mock Server** - Provides complete local development testing environment 🚀