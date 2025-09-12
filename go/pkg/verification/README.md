# TEENet Signature Verification Package

## Overview

Production-ready cryptographic signature verification package supporting multiple curves and protocols.

## Features

### Supported Algorithms

| Curve | Protocol | Public Key Formats | Signature Formats | Library Used |
|-------|----------|-------------------|-------------------|--------------|
| **ED25519** | EdDSA | 32 bytes | 64 bytes | Go stdlib |
| **SECP256K1** | ECDSA | Compressed (33), Uncompressed (65), Raw (64) | DER, Raw (64) | btcec/v2 |
| **SECP256K1** | Schnorr | Compressed (33), Uncompressed (65), Raw (64) | 64 bytes | btcec/v2 |
| **SECP256R1** | ECDSA | Compressed (33), Uncompressed (65), Raw (64) | DER, Raw (64) | Go stdlib |
| **SECP256R1** | Schnorr | Compressed (33), Uncompressed (65), Raw (64) | 64 bytes | Custom impl |

### Key Features

- ✅ **Production Ready**: Uses battle-tested libraries (btcec for Bitcoin, Go stdlib for P-256)
- ✅ **Multiple Formats**: Supports compressed, uncompressed, and raw public key formats
- ✅ **Flexible Signatures**: Handles both DER-encoded and raw signature formats
- ✅ **Comprehensive Testing**: Full test coverage with unit, integration, and benchmark tests
- ✅ **Error Handling**: Detailed error messages for debugging
- ✅ **Performance**: Optimized with benchmarks showing excellent performance

## Usage

### Basic Example

```go
import (
    "github.com/TEENet-io/teenet-sdk/go/pkg/verification"
    "github.com/TEENet-io/teenet-sdk/go/pkg/constants"
)

// Verify ED25519 signature
valid, err := verification.VerifySignature(
    message,
    publicKey,
    signature,
    0, // protocol (ignored for ED25519)
    constants.CurveED25519,
)

// Verify SECP256K1 ECDSA signature
valid, err := verification.VerifySignature(
    message,
    publicKey,
    signature,
    constants.ProtocolECDSA,
    constants.CurveSECP256K1,
)

// Verify SECP256K1 Schnorr signature
valid, err := verification.VerifySignature(
    message,
    publicKey,
    signature,
    constants.ProtocolSchnorr,
    constants.CurveSECP256K1,
)
```

### Client Integration

```go
import client "github.com/TEENet-io/teenet-sdk/go"

// Initialize client
c := client.NewClient(configServerAddr)
err := c.Init(nil)

// Verify signature using app ID
valid, err := c.Verify(message, signature, appID)
```

## Performance Benchmarks

```
BenchmarkED25519Verification-4          22755     50992 ns/op       0 B/op       0 allocs/op
BenchmarkSecp256k1ECDSAVerification-4    7956    153737 ns/op     864 B/op      19 allocs/op
BenchmarkSecp256r1ECDSAVerification-4   14378     82033 ns/op    1712 B/op      35 allocs/op
```

## Public Key Formats

### Uncompressed (65 bytes)
- Format: `0x04 || X (32 bytes) || Y (32 bytes)`
- Supported by all curves

### Compressed (33 bytes)
- Format: `0x02/0x03 || X (32 bytes)`
- Prefix 0x02 for even Y, 0x03 for odd Y
- Supported by SECP256K1 and SECP256R1

### Raw (64 bytes)
- Format: `X (32 bytes) || Y (32 bytes)`
- No prefix byte
- Supported by SECP256K1 and SECP256R1

## Signature Formats

### DER Encoded
- ASN.1 DER format for ECDSA signatures
- Variable length (typically 70-72 bytes)
- Standard format for Bitcoin and many other systems

### Raw Format (64 bytes)
- Format: `R (32 bytes) || S (32 bytes)`
- Fixed length, simpler to handle
- Common in Ethereum and other systems

### Schnorr (64 bytes)
- Format: `R (32 bytes) || S (32 bytes)`
- Used for Schnorr signatures on SECP256K1
- BIP340 compatible for Bitcoin

## Testing

Run all tests:
```bash
go test ./pkg/verification -v
```

Run benchmarks:
```bash
go test ./pkg/verification -bench=. -benchmem
```

Run integration tests:
```bash
go test ./example -v -run TestClientVerifyIntegration
```

## Dependencies

- `github.com/btcsuite/btcd/btcec/v2` - Bitcoin secp256k1 implementation
- Go standard library - ED25519 and P-256 support

## Security Considerations

1. **Message Hashing**: The package automatically hashes messages with SHA-256 for ECDSA/Schnorr
2. **Point Validation**: Validates that public keys are valid points on the curve
3. **Range Checking**: Validates signature components are within valid ranges
4. **No Side Channels**: Uses constant-time operations where possible

## License

Copyright (c) 2025 TEENet Technology (Hong Kong) Limited. All Rights Reserved.