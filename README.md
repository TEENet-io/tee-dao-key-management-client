# TEE DAO Key Management Client

A comprehensive TEENet distributed key management client library with multi-language support and distributed voting signature mechanism, including a complete local testing environment.

## 🚀 Core Components

### 1. Client Libraries
- **Go** - Production-ready implementation with distributed voting signatures
- **TypeScript** - Node.js compatible implementation

### 2. Example Applications
- **TEENet Signature Tool** - Unified web application supporting digital signatures and distributed voting
- **Distributed Voting Signatures** - M-of-N threshold voting mechanism
- **Multi-Protocol Support** - ECDSA and Schnorr protocols
- **Multi-Curve Support** - ED25519, SECP256K1, SECP256R1 curves
- **Docker Ready** - Containerized deployment

### 3. Mock Server Environment
- **Mock DAO Server** - Simulates distributed key management with real cryptographic operations
- **Mock Config Server** - Provides node discovery and configuration
- **Mock App Node** - Simulates user management system

## ✨ Key Features

### Distributed Voting Signatures
- **M-of-N Threshold Voting**: Configurable voting requirements (e.g., 2-of-3, 3-of-5)
- **Concurrent Processing**: Simultaneous voting requests to all target nodes
- **Complete Collection**: Waits for all voting responses with detailed status
- **Automatic Signing**: Generates cryptographic signatures upon voting approval
- **Loop Prevention**: Uses `is_forwarded` flag to prevent infinite loops

### Key Management
- **Secure Message Signing**: Sign messages using distributed cryptographic keys
- **AppID Service Integration**: Get public keys and sign messages using AppID
- **Multi-Protocol Support**: ECDSA and Schnorr signature protocols
- **Multi-Curve Support**: ED25519, SECP256K1, SECP256R1 curves
- **TLS Security**: Secure communication using mutual TLS authentication

### Mock Server Features
- **Semantic App IDs**: 
  - `secure-messaging-app` (Schnorr + ED25519)
  - `financial-trading-platform` (ECDSA + SECP256R1)
  - `digital-identity-service` (Schnorr + SECP256K1)
  - `bitcoin-wallet-app` (ECDSA + SECP256K1)
- **Deterministic Testing**: Reproducible key generation for testing
- **Complete Environment**: Config server, DAO server, app node

## 🏁 Quick Start

### Start Mock Server Environment

```bash
cd mock-server
./start-test-env.sh
```

This starts:
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

### TEENet Signature Tool

**Start Signature Tool:**
```bash
cd go/example/signature-tool
APP_ID=secure-messaging-app TEE_CONFIG_ADDR=localhost:50052 go run .
```

Web interface available at: `http://localhost:8080`

**Docker Deployment:**
```bash
cd go/example/signature-tool
docker build -t teenet-signature-tool .
docker run -p 8080:8080 \
  -e APP_ID=secure-messaging-app \
  -e TEE_CONFIG_ADDR=host.docker.internal:50052 \
  teenet-signature-tool
```

### Stop Mock Server

```bash
cd mock-server
./stop-test-env.sh
```

## 🗳️ Distributed Voting Signature Workflow

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │   Application   │    │ TEE DAO Client  │    │ TEE DAO Network │
│                 │    │                 │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │                      │
          │ 1. POST /api/vote    │                      │                      │
          ├─────────────────────►│                      │                      │
          │ {                    │                      │                      │
          │   message,           │                      │                      │
          │   signer_app_id,     │                      │                      │
          │   target_app_ids,    │                      │                      │
          │   required_votes     │                      │                      │
          │ }                    │                      │                      │
          │                      │                      │                      │
          │                      │ 2. VotingSign()      │                      │
          │                      ├─────────────────────►│                      │
          │                      │                      │                      │
          │                      │                      │ 3. Concurrent voting requests   │
          │                      │                      │ ┌─────────────────┐             │
          │                      │                      │ │                 │             │
          │                      │                      ├─┤ Target App ID 1 │             │
          │                      │                      │ │ (Local decision)│             │
          │                      │                      │ └─────────────────┘             │
          │                      │                      │ ┌─────────────────┐             │
          │                      │                      │ │                 │             │
          │                      │                      ├─┤ Target App ID 2 │             │
          │                      │                      │ │ (Local decision)│             │
          │                      │                      │ └─────────────────┘             │
          │                      │                      │ ┌─────────────────┐             │
          │                      │                      │ │                 │             │
          │                      │                      ├─┤ Target App ID N │             │
          │                      │                      │ │ (Local decision)│             │
          │                      │                      │ └─────────────────┘             │
          │                      │                      │                                  │
          │                      │                      │ 4. Collect all voting results   │
          │                      │                      │ (Wait for all responses)        │
          │                      │                      │                                  │
          │                      │                      │ 5. Internal processing:         │
          │                      │                      │ - Count approvals               │
          │                      │                      │ - Check threshold               │
          │                      │                      │                                  │
          │                      │                      │ 6. Generate signature           │
          │                      │                      │ (if voting passes)              │
          │                      │                      ├─────────────────────────────────►│
          │                      │                      │                                  │
          │                      │                      │ 7. Return signature             │
          │                      │                      │◄─────────────────────────────────┤
          │                      │                      │                                  │
          │                      │ 8. Return results    │                                  │
          │                      │◄─────────────────────┤                                  │
          │                      │                      │                                  │
          │ 9. Complete response │                      │                                  │
          │ {                    │                      │                                  │
          │   success: true,     │                      │                                  │
          │   approved: true,    │                      │                                  │
          │   voting_results: {  │                      │                                  │
          │     vote_details,    │                      │                                  │
          │     final_result     │                      │                                  │
          │   },                 │                      │                                  │
          │   signature          │                      │                                  │
          │ }                    │                      │                                  │
          │◄─────────────────────┤                      │                                  │
          │                      │                      │                                  │
```

### Key Features
- **M-of-N Threshold**: Configurable voting requirements (e.g., 2-of-3, 3-of-5)
- **Concurrent Processing**: Parallel voting requests to all target nodes
- **Complete Collection**: Waits for all responses before making decisions
- **Detailed Tracking**: Records each node's voting status and errors
- **Automatic Signing**: Generates cryptographic signature upon voting approval
- **Real-time UI**: Dynamic display of voting progress and results

### Voting Decision Logic
Current voting decision implementation:
- Approves if message content contains "test" (case-insensitive)
- Can be customized by modifying the application code

## Go Implementation

### Installation

```bash
go get github.com/TEENet-io/tee-dao-key-management-client/go
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    client "github.com/TEENet-io/tee-dao-key-management-client/go"
)

func main() {
    // Create client
    client := client.NewClient("localhost:50052")
    defer client.Close()

    // Initialize client (fetch config + establish TLS connection)
    if err := client.Init(nil); err != nil { // nil uses default auto-approve voting handler
        log.Fatalf("Initialization failed: %v", err)
    }

    fmt.Printf("Client connected, Node ID: %d\n", client.GetNodeID())

    // Example 1: Get public key by App ID
    appID := "secure-messaging-app"
    publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID)
    if err != nil {
        log.Printf("Failed to get public key: %v", err)
    } else {
        fmt.Printf("Public key for App ID %s:\n", appID)
        fmt.Printf("  - Protocol: %s\n", protocol)
        fmt.Printf("  - Curve: %s\n", curve)
        fmt.Printf("  - Public Key: %s\n", publicKey)
    }

    // Example 2: Sign message with App ID
    message := []byte("Hello from AppID Service!")
    signature, err := client.SignWithAppID(message, appID)
    if err != nil {
        log.Printf("Signing with App ID failed: %v", err)
    } else {
        fmt.Printf("Signing with App ID successful!\n")
        fmt.Printf("Message: %s\n", string(message))
        fmt.Printf("Signature: %x\n", signature)
    }

    // Example 3: Distributed voting signature
    targetAppIDs := []string{"secure-messaging-app", "financial-trading-platform", "digital-identity-service"}
    requiredVotes := 2
    votingMessage := []byte("Multi-party signature request")
    localApproval := true
    
    // Create request data (simplified example)
    requestBody := []byte(`{"message":"dGVzdA==","signer_app_id":"secure-messaging-app","target_app_ids":["app-1","app-2","app-3"],"required_votes":2}`)
    
    votingResult, err := client.VotingSign(votingMessage, appID, targetAppIDs, requiredVotes, localApproval, requestBody)
    if err != nil {
        log.Printf("Voting signature failed: %v", err)
    } else {
        fmt.Printf("Voting signature successful!\n")
        fmt.Printf("Votes received: %d/%d\n", votingResult.SuccessfulVotes, votingResult.RequiredVotes)
        fmt.Printf("Final result: %s\n", votingResult.FinalResult)
        if votingResult.Signature != nil {
            fmt.Printf("Signature: %x\n", votingResult.Signature)
        }
        
        // Print detailed vote results
        fmt.Printf("Vote details:\n")
        for i, detail := range votingResult.VoteDetails {
            status := "FAILED"
            if detail.Success && detail.Response {
                status = "APPROVED"
            } else if detail.Success && !detail.Response {
                status = "REJECTED"
            }
            fmt.Printf("  %d. %s: %s\n", i+1, detail.ClientID, status)
            if detail.Error != "" {
                fmt.Printf("     Error: %s\n", detail.Error)
            }
        }
    }
}
```

## TypeScript Implementation

### Installation

```bash
cd typescript
npm install
```

### Basic Usage

```typescript
import { Client } from './src/client';

async function main() {
  // Create client
  const client = new Client('localhost:50052');

  try {
    // Initialize client
    await client.init();
    console.log(`Client connected, Node ID: ${client.getNodeId()}`);

    // Example 1: Get public key by App ID
    const appID = 'secure-messaging-app';
    try {
      const { publickey, protocol, curve } = await client.getPublicKeyByAppID(appID);
      console.log(`Public key for App ID ${appID}:`);
      console.log(`  - Protocol: ${protocol}`);
      console.log(`  - Curve: ${curve}`);
      console.log(`  - Public Key: ${publickey}`);
    } catch (error) {
      console.error(`Failed to get public key: ${error}`);
    }

    // Example 2: Sign message with App ID
    const message = new TextEncoder().encode('Hello from AppID Service!');
    try {
      const signature = await client.signWithAppID(message, appID);
      console.log('Signing with App ID successful!');
      console.log(`Message: ${new TextDecoder().decode(message)}`);
      console.log(`Signature: ${Buffer.from(signature).toString('hex')}`);
    } catch (error) {
      console.error(`Signing with App ID failed: ${error}`);
    }

    // Example 3: Distributed voting signature
    const targetAppIDs = ['secure-messaging-app', 'financial-trading-platform', 'digital-identity-service'];
    const requiredVotes = 2;
    const votingMessage = new TextEncoder().encode('Multi-party signature request');
    
    try {
      const votingResult = await client.votingSign(votingMessage, appID, targetAppIDs, requiredVotes);
      console.log('Voting signature successful!');
      console.log(`Votes received: ${votingResult.successfulVotes}/${votingResult.requiredVotes}`);
      console.log(`Final result: ${votingResult.finalResult}`);
      if (votingResult.signature) {
        console.log(`Signature: ${Buffer.from(votingResult.signature).toString('hex')}`);
      }
      
      // Print detailed vote results
      console.log(`Vote details:`);
      votingResult.voteDetails.forEach((detail, index) => {
        const status = detail.success ? (detail.response ? 'APPROVED' : 'REJECTED') : 'FAILED';
        console.log(`  ${index + 1}. ${detail.clientId}: ${status}`);
        if (detail.error) {
          console.log(`     Error: ${detail.error}`);
        }
      });
    } catch (error) {
      console.error(`Voting signature failed: ${error}`);
    }

  } catch (error) {
    console.error('Error:', error);
  } finally {
    await client.close();
  }
}

main();
```

## API Reference

### Client Creation and Initialization

**Go:**
```go
client := client.NewClient("localhost:50052")
err := client.Init(nil) // nil uses default voting handler
```

**TypeScript:**
```typescript
const client = new Client('localhost:50052');
await client.init();
```

### Distributed Voting Signature

**Go:**
```go
votingResult, err := client.VotingSign(message, signerAppID, targetAppIDs, requiredVotes, localApproval, requestBody)
```

**TypeScript:**
```typescript
const votingResult = await client.votingSign(message, signerAppId, targetAppIds, requiredVotes)
```

### AppID Service Methods

**Get Public Key by AppID:**
```go
publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID)
```

**Sign with AppID:**
```go
signature, err := client.SignWithAppID(message, appID)
```

### Protocol and Curve Constants

**Protocols:**
- `constants.ProtocolECDSA` (1)
- `constants.ProtocolSchnorr` (2)

**Curves:**
- `constants.CurveED25519` (1)
- `constants.CurveSECP256K1` (2)
- `constants.CurveSECP256R1` (3)

## Project Structure

```
├── go/                     # Go client implementation
│   ├── client.go          # Main client (with distributed voting)
│   ├── pkg/               # Core packages
│   │   ├── config/        # Configuration client
│   │   ├── constants/     # Protocol and curve constants
│   │   ├── task/          # Task client for signing
│   │   ├── usermgmt/      # User management client
│   │   ├── utils/         # Utility functions
│   │   └── voting/        # Voting service
│   ├── example/           # Go examples
│   │   ├── main.go        # Basic client example
│   │   └── signature-tool/ # Signature tool web application
│   │       ├── main.go    # Web application main program
│   │       ├── types.go   # Data structures (simplified)
│   │       ├── crypto.go  # Cryptographic operations
│   │       ├── server.go  # Static file service (no-cache)
│   │       ├── voting.go  # Voting processing logic
│   │       ├── frontend/  # Frontend files
│   │       ├── README.md  # Detailed documentation
│   │       └── Dockerfile      # Docker build configuration
│   └── proto/             # Generated Go protobuf files
├── typescript/            # TypeScript client implementation
│   ├── src/               # TypeScript source code
│   │   ├── client.ts      # Main client
│   │   ├── config-client.ts # Configuration client
│   │   ├── task-client.ts # Task client
│   │   ├── appid-client.ts # AppID client
│   │   ├── types.ts       # Types and constants
│   │   └── example.ts     # TypeScript example
│   ├── proto/             # Protobuf definitions
│   └── dist/              # Compiled JavaScript
├── mock-server/           # Complete Mock server environment
│   ├── dao-server.go      # Mock DAO server
│   ├── mock-config-server.go # Mock config server
│   ├── mock-app-node.go   # Mock app node
│   ├── proto/             # Protocol buffer definitions
│   ├── certs/             # TLS certificates (auto-generated)
│   ├── logs/              # Server logs
│   ├── start-test-env.sh  # Start all services
│   ├── stop-test-env.sh   # Stop all services
│   └── README.md          # Detailed documentation
```

## Examples and Documentation

- **Go Client**: See [go/example/main.go](go/example/main.go)
- **TypeScript Client**: See [typescript/src/example.ts](typescript/src/example.ts)
- **TEENet Signature Tool**: See [go/example/signature-tool/](go/example/signature-tool/) for detailed documentation
- **Mock Server**: See [mock-server/README.md](mock-server/README.md) for detailed documentation

## Latest Architecture Optimizations

### Distributed Voting System Improvements
1. **Simplified API**: Removed redundant parameters, `VotingSign` now automatically parses request data
2. **Unified Voting Method**: Only one `VotingSign` method, removed duplicate methods
3. **Correct Signer**: Uses `signer_app_id` as signature generator, not receiver
4. **Cleaned Data Structures**: Removed unnecessary fields like `TaskID` and `TotalParticipants`
5. **Cache-Free Deployment**: Web application supports zero-cache deployment, no need to manually clear browser cache

### Technical Features
- **Loop Prevention**: Uses `is_forwarded` flag to prevent infinite voting request loops
- **Concurrent Processing**: Uses goroutines to handle multiple voting requests concurrently
- **Complete Collection**: Waits for all voting responses, provides detailed voting status
- **Automatic Signing**: Automatically generates signatures using key management system upon voting approval
- **Modular Design**: Clean code structure for easy maintenance and extension

## Complete Testing Workflow

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
   
   # Signature tool web application
   cd go/example/signature-tool
   APP_ID=secure-messaging-app go run .
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

- All communications use mutual TLS authentication
- Hostname verification is maintained (never disabled)
- Certificate and key files are excluded via .gitignore
- No hardcoded credentials or secrets
- Voting requests include loop prevention mechanism

## License

This project is part of the TEENet ecosystem for secure distributed key management.