# TEENet SDK

A comprehensive TEENet sdk library with multi-language support, distributed voting signature mechanism, and signature verification, including a complete local testing environment.

> **🎉 New in v2.1**: Added signature verification support with `Verify()` method for both Go and TypeScript SDKs. See [Latest Updates](#-latest-updates-v21) for details.
> 
> **⚠️ Breaking Change in v2.0**: New unified `Sign()` API replaces `SignWithAppID` and `VotingSign` methods. Target app IDs and required votes are now automatically fetched from server configuration.

## 🚀 Core Components

### 1. Client Libraries
- **Go** - Production-ready implementation with distributed voting signatures and signature verification
- **TypeScript** - Node.js compatible implementation with full feature parity

### 2. Example Applications
- **TEENet Signature Tool** - Unified web application supporting digital signatures, verification, and distributed voting
- **Distributed Voting Signatures** - M-of-N threshold voting mechanism
- **Signature Verification** - Verify signatures across all supported protocols and curves
- **Multi-Protocol Support** - ECDSA and Schnorr protocols
- **Multi-Curve Support** - ED25519, SECP256K1, SECP256R1 curves
- **Docker Ready** - Containerized deployment

### 3. Mock Server Environment
- **Mock DAO Server** - Simulates distributed key management with real cryptographic operations
- **Mock Config Server** - Provides node discovery and configuration
- **Mock App Node** - Simulates user management system

## ✨ Key Features

### Distributed Voting Signatures
- **Server-Configured Voting**: Target nodes and required votes automatically fetched from server
- **M-of-N Threshold Voting**: Server-configured voting requirements based on project settings
- **Concurrent Processing**: Simultaneous voting requests to all target nodes
- **Complete Collection**: Waits for all voting responses with detailed status
- **Automatic Signing**: Generates cryptographic signatures upon voting approval
- **Loop Prevention**: Uses `is_forwarded` flag to prevent infinite loops

### Key Management
- **Secure Message Signing**: Sign messages using distributed cryptographic keys
- **Signature Verification**: Verify signatures with automatic protocol and curve detection
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

## API Reference

### Core Methods

#### Sign (Unified API)
```go
// Go
result, err := client.Sign(request *SignRequest) (*SignResult, error)

// TypeScript
result = await client.sign(request: SignRequest): Promise<SignResult>
```

#### GetPublicKeyByAppID
```go
// Go
publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID string)

// TypeScript
const { publicKey, protocol, curve } = await client.getPublicKeyByAppID(appID: string)
```

#### Verify
```go
// Go
valid, err := client.Verify(message []byte, signature []byte, appID string) (bool, error)

// TypeScript
valid = await client.verify(message: Buffer, signature: Buffer, appID: string): Promise<boolean>
```

### Core Types

#### SignRequest
```go
// Go
type SignRequest struct {
    Message       []byte        // Message to sign
    AppID         string        // Application identifier
    EnableVoting  bool          // Enable multi-party voting
    LocalApproval bool          // Local voting decision (for voting)
    HTTPRequest   *http.Request // HTTP request context (for voting)
}

// TypeScript
interface SignRequest {
    message: Uint8Array;       // Message to sign
    appID: string;             // Application identifier
    enableVoting?: boolean;    // Enable multi-party voting
    localApproval?: boolean;   // Local voting decision
    httpRequest?: any;         // HTTP request object
}
```

#### SignResult
```go
// Go
type SignResult struct {
    Success    bool        // Operation success
    Signature  []byte      // Generated signature
    Error      string      // Error message if failed
    VotingInfo *VotingInfo // Voting details (when voting enabled)
}

// TypeScript
interface SignResult {
    success: boolean;          // Operation success
    signature?: Uint8Array;    // Generated signature
    error?: string;            // Error message
    votingInfo?: VotingInfo;   // Voting details
}
```

#### VotingInfo
```go
// Go
type VotingInfo struct {
    TotalTargets    int          // Total voting nodes
    SuccessfulVotes int          // Number of approvals
    RequiredVotes   int          // Threshold for approval
    VoteDetails     []VoteDetail // Individual vote information
}

// TypeScript
interface VotingInfo {
    totalTargets: number;      // Total voting nodes
    successfulVotes: number;    // Number of approvals
    requiredVotes: number;      // Threshold for approval
    voteDetails: VoteDetail[];  // Individual vote information
}
```

### Protocol and Curve Constants

**Protocols:**
- `ProtocolECDSA` (1)
- `ProtocolSchnorr` (2)

**Curves:**
- `CurveED25519` (1)
- `CurveSECP256K1` (2)
- `CurveSECP256R1` (3)

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
          │   signer_app_id      │                      │                      │
          │ }                    │                      │                      │
          │                      │                      │                      │
          │ (target_app_ids and │                      │                      │
          │ required_votes are  │                      │                      │
          │ fetched from server)│                      │                      │
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
- **Server-Driven Configuration**: Target nodes and voting threshold from server settings
- **M-of-N Threshold**: Server-configured voting requirements
- **Concurrent Processing**: Parallel voting requests to all target nodes
- **Complete Collection**: Waits for all responses before making decisions
- **Detailed Tracking**: Records each node's voting status and errors
- **Automatic Signing**: Generates cryptographic signature upon voting approval
- **Real-time UI**: Dynamic display of voting progress and results

### Voting Decision Logic
Current voting decision implementation:
- **Auto-Approval**: Messages containing "test" (case-insensitive) are automatically approved
- **Auto-Rejection**: Messages without "test" are automatically rejected
- **Customizable**: Can be modified in the application code to implement custom approval logic
- **Consistent**: Same logic applied across all voting nodes for predictable testing

## Go Implementation

### Installation

```bash
go get github.com/TEENet-io/teenet-sdk/go
```

### Basic Usage

```go
package main

import (
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    
    client "github.com/TEENet-io/teenet-sdk/go"
)

func main() {
    // Create client
    teeClient := client.NewClient("localhost:50052")
    defer teeClient.Close()

    // Initialize client with custom voting handler (optional)
    if err := teeClient.Init(nil); err != nil {
        log.Fatalf("Initialization failed: %v", err)
    }

    fmt.Printf("Client connected, Node ID: %d\n", teeClient.GetNodeID())

    // Example 1: Simple signature using new Sign API
    appID := "secure-messaging-app"
    message := []byte("Hello from AppID Service!")
    
    result, err := teeClient.Sign(&client.SignRequest{
        Message: message,
        AppID:   appID,
        EnableVoting: false,
    })
    if err != nil {
        log.Printf("Signing failed: %v", err)
    } else if result.Success {
        fmt.Printf("Signature: %x\n", result.Signature)
    }

    // Example 2: Get public key by App ID
    publicKey, protocol, curve, err := teeClient.GetPublicKeyByAppID(appID)
    if err != nil {
        log.Printf("Failed to get public key: %v", err)
    } else {
        fmt.Printf("Public key for App ID %s:\n", appID)
        fmt.Printf("  - Protocol: %s\n", protocol)
        fmt.Printf("  - Curve: %s\n", curve)
        fmt.Printf("  - Public Key: %s\n", publicKey)
    }

    // Example 3: Verify signature
    if result.Success && result.Signature != nil {
        valid, err := teeClient.Verify(message, result.Signature, appID)
        if err != nil {
            log.Printf("Verification failed: %v", err)
        } else {
            fmt.Printf("Signature valid: %v\n", valid)
        }
    }
}

// Example 4: Voting signature in HTTP handler
func handleVotingRequest(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Message     string `json:"message"`
        SignerAppID string `json:"signer_app_id"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    // Decode message
    messageBytes, _ := base64.StdEncoding.DecodeString(req.Message)
    
    // Make local voting decision
    localApproval := strings.Contains(string(messageBytes), "test")
    
    // Use Sign API with voting enabled
    result, err := teeClient.Sign(&client.SignRequest{
        Message:       messageBytes,
        AppID:         req.SignerAppID,
        EnableVoting:  true,
        LocalApproval: localApproval,
        HTTPRequest:   r,  // Pass the incoming HTTP request
    })
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return results
    response := map[string]interface{}{
        "success": result.Success,
        "signature": hex.EncodeToString(result.Signature),
    }
    
    if result.VotingInfo != nil {
        response["voting_info"] = result.VotingInfo
    }
    
    json.NewEncoder(w).Encode(response)
}
```

## TypeScript Implementation

### Installation

```bash
npm install @teenet/teenet-sdk
```

### Basic Usage

```typescript
import { Client, SignRequest } from '@teenet/teenet-sdk';

async function main() {
  // Create and initialize client
  const client = new Client('localhost:50052');
  await client.init();
  
  console.log(`Client connected, Node ID: ${client.getNodeId()}`);

  // Example 1: Simple signature using new sign API
  const appID = 'secure-messaging-app';
  const message = new TextEncoder().encode('Hello from AppID Service!');
  
  const result = await client.sign({
    message: message,
    appID: appID,
    enableVoting: false,
  });
  
  if (result.success) {
    console.log(`Signature: ${Buffer.from(result.signature).toString('hex')}`);
  }

  // Example 2: Get public key by App ID
  const { publicKey, protocol, curve } = await client.getPublicKeyByAppID(appID);
  console.log(`Public key for ${appID}:`);
  console.log(`  - Protocol: ${protocol}`);
  console.log(`  - Curve: ${curve}`);
  console.log(`  - Public Key: ${publicKey}`);

  // Example 3: Verify signature
  if (result.success && result.signature) {
    const valid = await client.verify(message, result.signature, appID);
    console.log(`Signature valid: ${valid}`);
  }
  
  await client.close();
}

// Example 4: Voting signature in Express handler
app.post('/vote', async (req, res) => {
  // Extract message from incoming request
  const message = Buffer.from(req.body.message, 'base64');
  const signerAppID = req.body.signer_app_id;
  
  // Make local voting decision
  const messageStr = message.toString();
  const localApproval = messageStr.includes('test');
  
  // Use sign API with voting enabled
  const result = await client.sign({
    message: message,
    appID: signerAppID,
    enableVoting: true,
    localApproval: localApproval,
    httpRequest: req,  // Pass the incoming Express request
  });
  
  // Return results
  res.json({
    success: result.success,
    signature: result.signature ? 
      Buffer.from(result.signature).toString('hex') : null,
    votingInfo: result.votingInfo
  });
});

main().catch(console.error);
```

## Project Structure

```
├── go/                     # Go client implementation
│   ├── client.go          # Main client (with distributed voting and verification)
│   ├── pkg/               # Core packages
│   │   ├── config/        # Configuration client
│   │   ├── constants/     # Protocol and curve constants
│   │   ├── task/          # Task client for signing
│   │   ├── usermgmt/      # User management client
│   │   ├── utils/         # Utility functions
│   │   ├── verification/  # Signature verification
│   │   └── voting/        # Voting service
│   ├── example/           # Go examples
│   │   ├── main.go        # Basic client example with verification
│   │   └── signature-tool/ # Signature tool web application
│   │       ├── main.go    # Web application main program
│   │       ├── types.go   # Data structures (simplified)
│   │       ├── server.go  # Static file service (no-cache)
│   │       ├── voting.go  # Voting processing logic
│   │       ├── frontend/  # Frontend files
│   │       ├── README.md  # Detailed documentation
│   │       └── Dockerfile      # Docker build configuration
│   └── proto/             # Generated Go protobuf files
├── typescript/            # TypeScript client implementation
│   ├── src/               # TypeScript source code
│   │   ├── client.ts      # Main client with verification
│   │   ├── config-client.ts # Configuration client
│   │   ├── task-client.ts # Task client
│   │   ├── appid-client.ts # AppID client
│   │   ├── types.ts       # Types and constants
│   │   ├── verification/  # Signature verification
│   │   │   └── verify.ts  # Verification implementation
│   │   └── example.ts     # TypeScript example with verification
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

## 🆕 Latest Updates (v2.1)

### ⭐ New Features (v2.1)
1. **Signature Verification**: Added `Verify()` method to both Go and TypeScript SDKs
   - Automatic protocol and curve detection based on AppID
   - Support for all curves: ED25519, SECP256K1, SECP256R1
   - Support for all protocols: ECDSA, Schnorr, EdDSA
   - Multiple key formats supported (compressed, uncompressed, raw)
   - Production-ready implementation using established libraries (btcec for Go, elliptic for TypeScript)

2. **Updated Signature Tool**: Now uses SDK's built-in verification instead of custom implementation
   - Cleaner codebase with removed redundant verification code
   - Consistent verification across all SDK consumers

## 🆕 Previous Updates (v2.0)

### ⭐ Major API Changes
1. **Unified Sign API**: New `Sign()` method replaces separate `SignWithAppID` and `VotingSign` methods
   - **Before**: 
     ```go
     signature, err := client.SignWithAppID(message, appID)
     votingResult, err := client.VotingSign(req, message, appID, localApproval)
     ```
   - **After**: 
     ```go
     result, err := client.Sign(&SignRequest{
         Message: message,
         AppID: appID,
         EnableVoting: false, // or true for voting
         LocalApproval: localApproval,
         HTTPRequest: req,
     })
     ```

2. **Automatic Server Configuration**: Target nodes and voting threshold fetched from server
   - No need to hardcode target App IDs in client code
   - Voting threshold automatically determined by server settings
   - More flexible and easier to maintain

### Distributed Voting System Improvements
1. **Server-Driven Configuration**: Target nodes and voting requirements from server settings
2. **HTTP Request Integration**: `VotingSign` accepts HTTP request objects for better header and body handling
3. **Unified API Signature**: Both Go and TypeScript versions have identical method signatures
4. **Smart Vote Filtering**: Only shows votes from target App IDs, excludes local vote when not in target list
5. **Correct Signer**: Uses `signer_app_id` as signature generator, not receiver
6. **Cache-Free Deployment**: Web application supports zero-cache deployment
7. **Improved Success Conditions**: Clear indication that messages containing "test" will succeed, others will fail

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