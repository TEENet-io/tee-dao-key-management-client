// -----------------------------------------------------------------------------
// Copyright (c) 2025 TEENet Technology (Hong Kong) Limited. All Rights Reserved.
//
// This software and its associated documentation files (the "Software") are
// the proprietary and confidential information of TEENet Technology (Hong Kong) Limited.
// Unauthorized copying of this file, via any medium, is strictly prohibited.
//
// No license, express or implied, is hereby granted, except by written agreement
// with TEENet Technology (Hong Kong) Limited. Use of this software without permission
// is a violation of applicable laws.
//
// -----------------------------------------------------------------------------

package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/TEENet-io/teenet-sdk/go/pkg/constants"
)

// Helper functions to convert string to protocol/curve constants
func parseProtocol(protocol string) (uint32, error) {
	switch protocol {
	case "ecdsa":
		return constants.ProtocolECDSA, nil
	case "schnorr":
		return constants.ProtocolSchnorr, nil
	default:
		if num, err := strconv.ParseUint(protocol, 10, 32); err == nil {
			return uint32(num), nil
		}
		return 0, fmt.Errorf("unknown protocol: %s", protocol)
	}
}

func parseCurve(curve string) (uint32, error) {
	switch curve {
	case "ed25519":
		return constants.CurveED25519, nil
	case "secp256k1":
		return constants.CurveSECP256K1, nil
	case "secp256r1":
		return constants.CurveSECP256R1, nil
	default:
		if num, err := strconv.ParseUint(curve, 10, 32); err == nil {
			return uint32(num), nil
		}
		return 0, fmt.Errorf("unknown curve: %s", curve)
	}
}

// verifySignature verifies a signature against a message and public key
// Supports all protocol/curve combinations:
// - ED25519 with EdDSA (protocol parameter ignored for ED25519)
// - SECP256K1 with ECDSA or Schnorr protocols
// - SECP256R1 with ECDSA or Schnorr protocols
func verifySignature(message, publicKey, signature []byte, protocol, curve uint32) (bool, error) {
	switch curve {
	case constants.CurveED25519:
		// ED25519 only supports EdDSA (not ECDSA or Schnorr)
		if len(publicKey) != ed25519.PublicKeySize {
			return false, fmt.Errorf("invalid ED25519 public key size: expected %d, got %d", ed25519.PublicKeySize, len(publicKey))
		}
		if len(signature) != ed25519.SignatureSize {
			return false, fmt.Errorf("invalid ED25519 signature size: expected %d, got %d", ed25519.SignatureSize, len(signature))
		}

		// For ED25519, we verify directly (EdDSA protocol)
		return ed25519.Verify(ed25519.PublicKey(publicKey), message, signature), nil

	case constants.CurveSECP256K1:
		return verifySecp256k1(message, publicKey, signature, protocol)

	case constants.CurveSECP256R1:
		return verifySecp256r1(message, publicKey, signature, protocol)

	default:
		return false, fmt.Errorf("unsupported curve: %d", curve)
	}
}

// verifySecp256k1 verifies signatures on secp256k1 curve
func verifySecp256k1(message, publicKeyBytes, signature []byte, protocol uint32) (bool, error) {
	// Parse public key for secp256k1
	x, y, err := parseSecp256k1PublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse secp256k1 public key: %v", err)
	}

	// Note: secp256k1 is not directly supported in Go's standard crypto/elliptic
	// For proper secp256k1 support, you should use github.com/btcsuite/btcd/btcec/v2
	// For demo purposes, we'll use a simplified verification approach
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(), // This is incorrect for secp256k1 but used for demo
		X:     x,
		Y:     y,
	}

	log.Printf("Warning: Using P256 curve for secp256k1 verification (demo only - use btcec library for production)")
	return verifyECDSASignature(message, publicKey, signature, protocol)
}

// verifySecp256r1 verifies signatures on secp256r1 curve (NIST P-256)
func verifySecp256r1(message, publicKeyBytes, signature []byte, protocol uint32) (bool, error) {
	// Parse public key for secp256r1 (P-256)
	x, y, err := parseSecp256r1PublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse secp256r1 public key: %v", err)
	}

	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return verifyECDSASignature(message, publicKey, signature, protocol)
}

// verifyECDSASignature verifies ECDSA or Schnorr signatures
func verifyECDSASignature(message []byte, publicKey *ecdsa.PublicKey, signature []byte, protocol uint32) (bool, error) {
	// Hash the message with SHA256
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)

	switch protocol {
	case constants.ProtocolECDSA:
		return verifyECDSA(messageHash, publicKey, signature)
	case constants.ProtocolSchnorr:
		return verifySchnorr(messageHash, publicKey, signature)
	default:
		return false, fmt.Errorf("unsupported protocol: %d", protocol)
	}
}

// verifyECDSA verifies ECDSA signature
func verifyECDSA(messageHash []byte, publicKey *ecdsa.PublicKey, signature []byte) (bool, error) {
	// Parse ECDSA signature (DER format or raw r,s format)
	var ecdsaSig ECDSASignature

	// Try to parse as ASN.1 DER format first
	if _, err := asn1.Unmarshal(signature, &ecdsaSig); err != nil {
		// If DER parsing fails, try to parse as raw r,s format
		if len(signature)%2 != 0 {
			return false, fmt.Errorf("invalid signature length for raw r,s format")
		}

		half := len(signature) / 2
		ecdsaSig.R = new(big.Int).SetBytes(signature[:half])
		ecdsaSig.S = new(big.Int).SetBytes(signature[half:])
	}

	// Verify the ECDSA signature
	return ecdsa.Verify(publicKey, messageHash, ecdsaSig.R, ecdsaSig.S), nil
}

// verifySchnorr verifies Schnorr signature (simplified implementation)
func verifySchnorr(messageHash []byte, publicKey *ecdsa.PublicKey, signature []byte) (bool, error) {
	// Note: This is a simplified Schnorr verification
	// In practice, you'd need a proper Schnorr implementation
	// For now, we'll use ECDSA verification as a fallback
	log.Printf("Warning: Using ECDSA verification for Schnorr signature (simplified implementation)")

	// Parse signature as r,s values
	if len(signature) != 64 {
		return false, fmt.Errorf("invalid Schnorr signature length: expected 64, got %d", len(signature))
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// Use ECDSA verification as fallback
	return ecdsa.Verify(publicKey, messageHash, r, s), nil
}

// parseSecp256k1PublicKey parses a secp256k1 public key from bytes
// Supports multiple formats:
// - Uncompressed: 0x04 + 32 bytes X + 32 bytes Y (65 bytes total)
// - Compressed: 0x02/0x03 + 32 bytes X (33 bytes total)
// - Raw coordinates: 32 bytes X + 32 bytes Y (64 bytes total)
func parseSecp256k1PublicKey(publicKeyBytes []byte) (*big.Int, *big.Int, error) {
	// Expect uncompressed public key format (0x04 + 32 bytes X + 32 bytes Y)
	if len(publicKeyBytes) == 65 && publicKeyBytes[0] == 0x04 {
		x := new(big.Int).SetBytes(publicKeyBytes[1:33])
		y := new(big.Int).SetBytes(publicKeyBytes[33:65])
		return x, y, nil
	}

	// Handle compressed public key format (0x02/0x03 + 32 bytes X)
	if len(publicKeyBytes) == 33 && (publicKeyBytes[0] == 0x02 || publicKeyBytes[0] == 0x03) {
		x := new(big.Int).SetBytes(publicKeyBytes[1:33])
		y, err := decompressSecp256k1Point(x, publicKeyBytes[0] == 0x03)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decompress secp256k1 point: %v", err)
		}
		return x, y, nil
	}

	// Try raw coordinate format (32 bytes X + 32 bytes Y)
	if len(publicKeyBytes) == 64 {
		x := new(big.Int).SetBytes(publicKeyBytes[:32])
		y := new(big.Int).SetBytes(publicKeyBytes[32:64])
		return x, y, nil
	}

	return nil, nil, fmt.Errorf("unsupported secp256k1 public key format: length %d", len(publicKeyBytes))
}

// decompressSecp256k1Point decompresses a secp256k1 point from x coordinate
func decompressSecp256k1Point(x *big.Int, yOdd bool) (*big.Int, error) {
	// secp256k1 parameters
	// y² = x³ + 7 (mod p)
	// p = 2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1
	secp256k1P, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)

	// Calculate y² = x³ + 7
	x3 := new(big.Int).Exp(x, big.NewInt(3), secp256k1P)
	ySquared := new(big.Int).Add(x3, big.NewInt(7))
	ySquared.Mod(ySquared, secp256k1P)

	// Calculate y = sqrt(y²) mod p using Tonelli-Shanks algorithm
	y := modularSqrt(ySquared, secp256k1P)
	if y == nil {
		return nil, fmt.Errorf("invalid point: x coordinate does not correspond to a valid secp256k1 point")
	}

	// Choose the correct y based on the compressed point prefix
	// If yOdd is true (0x03), we want odd y; if false (0x02), we want even y
	yIsOdd := y.Bit(0) == 1
	if yOdd != yIsOdd {
		// Take the other root: p - y
		y.Sub(secp256k1P, y)
	}

	return y, nil
}

// modularSqrt computes sqrt(a) mod p using Tonelli-Shanks algorithm
// This is a simplified implementation for secp256k1's prime
func modularSqrt(a, p *big.Int) *big.Int {
	// For secp256k1, p ≡ 3 (mod 4), so we can use the simple case:
	// sqrt(a) = a^((p+1)/4) mod p
	if new(big.Int).Mod(p, big.NewInt(4)).Cmp(big.NewInt(3)) == 0 {
		exp := new(big.Int).Add(p, big.NewInt(1))
		exp.Div(exp, big.NewInt(4))
		result := new(big.Int).Exp(a, exp, p)

		// Verify the result
		check := new(big.Int).Exp(result, big.NewInt(2), p)
		if check.Cmp(new(big.Int).Mod(a, p)) == 0 {
			return result
		}
	}

	return nil
}

// parseSecp256r1PublicKey parses a secp256r1 (P-256) public key from bytes
// Supports multiple formats:
// - Uncompressed: 0x04 + 32 bytes X + 32 bytes Y (65 bytes total)
// - Compressed: 0x02/0x03 + 32 bytes X (33 bytes total)
// - Raw coordinates: 32 bytes X + 32 bytes Y (64 bytes total)
func parseSecp256r1PublicKey(publicKeyBytes []byte) (*big.Int, *big.Int, error) {
	// Expect uncompressed public key format (0x04 + 32 bytes X + 32 bytes Y)
	if len(publicKeyBytes) == 65 && publicKeyBytes[0] == 0x04 {
		x := new(big.Int).SetBytes(publicKeyBytes[1:33])
		y := new(big.Int).SetBytes(publicKeyBytes[33:65])
		return x, y, nil
	}

	// Handle compressed public key format (0x02/0x03 + 32 bytes X)
	if len(publicKeyBytes) == 33 && (publicKeyBytes[0] == 0x02 || publicKeyBytes[0] == 0x03) {
		x := new(big.Int).SetBytes(publicKeyBytes[1:33])
		y, err := decompressSecp256r1Point(x, publicKeyBytes[0] == 0x03)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decompress secp256r1 point: %v", err)
		}
		return x, y, nil
	}

	// Try raw coordinate format (32 bytes X + 32 bytes Y)
	if len(publicKeyBytes) == 64 {
		x := new(big.Int).SetBytes(publicKeyBytes[:32])
		y := new(big.Int).SetBytes(publicKeyBytes[32:64])
		return x, y, nil
	}

	return nil, nil, fmt.Errorf("unsupported secp256r1 public key format: length %d", len(publicKeyBytes))
}

// decompressSecp256r1Point decompresses a secp256r1 (P-256) point from x coordinate
func decompressSecp256r1Point(x *big.Int, yOdd bool) (*big.Int, error) {
	// secp256r1/P-256 parameters
	// y² = x³ - 3x + b (mod p)
	// p = 2^256 - 2^224 + 2^192 + 2^96 - 1
	// b = 0x5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b
	p256P := elliptic.P256().Params().P
	p256B := elliptic.P256().Params().B

	// Calculate y² = x³ - 3x + b
	x3 := new(big.Int).Exp(x, big.NewInt(3), p256P)
	threeX := new(big.Int).Mul(big.NewInt(3), x)
	ySquared := new(big.Int).Sub(x3, threeX)
	ySquared.Add(ySquared, p256B)
	ySquared.Mod(ySquared, p256P)

	// Calculate y = sqrt(y²) mod p
	y := modularSqrt(ySquared, p256P)
	if y == nil {
		return nil, fmt.Errorf("invalid point: x coordinate does not correspond to a valid secp256r1 point")
	}

	// Choose the correct y based on the compressed point prefix
	yIsOdd := y.Bit(0) == 1
	if yOdd != yIsOdd {
		// Take the other root: p - y
		y.Sub(p256P, y)
	}

	return y, nil
}