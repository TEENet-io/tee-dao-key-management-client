package verification_test

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/TEENet-io/teenet-sdk/go/pkg/constants"
	"github.com/TEENet-io/teenet-sdk/go/pkg/verification"
	"github.com/btcsuite/btcd/btcec/v2"
	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// Mock implementation for testing client.Verify
type mockUserMgmtClient struct {
	publicKey string
	protocol  string
	curve     string
}

func TestClientVerifyIntegration(t *testing.T) {
	fmt.Println("\n=== Client Verify Integration Tests ===")
	
	t.Run("ED25519 Client Integration", func(t *testing.T) {
		// Generate ED25519 key pair
		pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
		message := []byte("Test message for ED25519")
		signature := ed25519.Sign(privKey, message)
		
		// Test direct verification function
		valid, err := verification.VerifySignature(message, pubKey, signature, 0, constants.CurveED25519)
		if err != nil {
			t.Fatalf("ED25519 verification failed: %v", err)
		}
		if !valid {
			t.Error("ED25519 signature should be valid")
		}
		
		fmt.Printf("✅ ED25519 integration test passed\n")
		fmt.Printf("   Public key: %x\n", pubKey)
		fmt.Printf("   Signature: %x\n", signature)
		fmt.Printf("   Verified: %v\n", valid)
	})
	
	t.Run("SECP256K1 ECDSA Client Integration", func(t *testing.T) {
		// Generate secp256k1 key pair
		privKey, _ := btcec.NewPrivateKey()
		pubKey := privKey.PubKey()
		message := []byte("Test message for secp256k1 ECDSA")
		
		// Hash and sign
		hasher := sha256.New()
		hasher.Write(message)
		messageHash := hasher.Sum(nil)
		sig := btcecdsa.Sign(privKey, messageHash)
		
		// Test different public key formats
		formats := []struct {
			name   string
			pubKey []byte
		}{
			{"Uncompressed", pubKey.SerializeUncompressed()},
			{"Compressed", pubKey.SerializeCompressed()},
			{"Raw (64 bytes)", pubKey.SerializeUncompressed()[1:]}, // Remove 0x04 prefix
		}
		
		for _, format := range formats {
			// Test DER signature format
			valid, err := verification.VerifySignature(message, format.pubKey, sig.Serialize(), 
				constants.ProtocolECDSA, constants.CurveSECP256K1)
			if err != nil {
				t.Errorf("%s: ECDSA verification failed: %v", format.name, err)
			}
			if !valid {
				t.Errorf("%s: ECDSA signature should be valid", format.name)
			}
			
			// Test raw signature format
			r := sig.R()
			s := sig.S()
			rawSig := make([]byte, 64)
			r.PutBytesUnchecked(rawSig[:32])
			s.PutBytesUnchecked(rawSig[32:])
			
			valid, err = verification.VerifySignature(message, format.pubKey, rawSig,
				constants.ProtocolECDSA, constants.CurveSECP256K1)
			if err != nil {
				t.Errorf("%s: Raw ECDSA verification failed: %v", format.name, err)
			}
			if !valid {
				t.Errorf("%s: Raw ECDSA signature should be valid", format.name)
			}
			
			fmt.Printf("✅ SECP256K1 ECDSA %s format test passed\n", format.name)
		}
	})
	
	t.Run("SECP256K1 Schnorr Client Integration", func(t *testing.T) {
		// Generate secp256k1 key pair
		privKey, _ := btcec.NewPrivateKey()
		pubKey := privKey.PubKey()
		message := []byte("Test message for secp256k1 Schnorr")
		
		// Hash and sign with Schnorr
		hasher := sha256.New()
		hasher.Write(message)
		messageHash := hasher.Sum(nil)
		sig, _ := schnorr.Sign(privKey, messageHash)
		
		// Test verification
		valid, err := verification.VerifySignature(message, pubKey.SerializeUncompressed(), sig.Serialize(),
			constants.ProtocolSchnorr, constants.CurveSECP256K1)
		if err != nil {
			t.Fatalf("Schnorr verification failed: %v", err)
		}
		if !valid {
			t.Error("Schnorr signature should be valid")
		}
		
		fmt.Printf("✅ SECP256K1 Schnorr integration test passed\n")
		fmt.Printf("   Signature: %x\n", sig.Serialize())
	})
	
	t.Run("SECP256R1 ECDSA Client Integration", func(t *testing.T) {
		// Generate P-256 key pair
		privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		message := []byte("Test message for P-256")
		
		// Hash and sign
		hasher := sha256.New()
		hasher.Write(message)
		messageHash := hasher.Sum(nil)
		r, s, _ := ecdsa.Sign(rand.Reader, privKey, messageHash)
		
		// Create raw signature
		rawSig := make([]byte, 64)
		r.FillBytes(rawSig[:32])
		s.FillBytes(rawSig[32:])
		
		// Test different public key formats
		formats := []struct {
			name   string
			pubKey []byte
		}{
			{"Uncompressed", elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)},
			{"Compressed", elliptic.MarshalCompressed(elliptic.P256(), privKey.X, privKey.Y)},
			{"Raw (64 bytes)", elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)[1:]},
		}
		
		for _, format := range formats {
			valid, err := verification.VerifySignature(message, format.pubKey, rawSig,
				constants.ProtocolECDSA, constants.CurveSECP256R1)
			if err != nil {
				t.Errorf("%s: P-256 verification failed: %v", format.name, err)
			}
			if !valid {
				t.Errorf("%s: P-256 signature should be valid", format.name)
			}
			
			fmt.Printf("✅ SECP256R1 ECDSA %s format test passed\n", format.name)
		}
	})
}

func TestSignatureFormatCompatibility(t *testing.T) {
	fmt.Println("\n=== Signature Format Compatibility Tests ===")
	
	// Test that we can verify signatures from external sources
	t.Run("External Signature Formats", func(t *testing.T) {
		// Test cases with known signatures (you would replace these with real test vectors)
		testCases := []struct {
			name      string
			message   []byte
			pubKey    []byte
			signature []byte
			protocol  uint32
			curve     uint32
			expected  bool
		}{
			// Add real test vectors here from known implementations
			// For example, from Bitcoin test vectors, Ethereum test vectors, etc.
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				valid, err := verification.VerifySignature(tc.message, tc.pubKey, tc.signature, tc.protocol, tc.curve)
				if err != nil && tc.expected {
					t.Errorf("Verification failed: %v", err)
				}
				if valid != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, valid)
				}
			})
		}
		
		fmt.Println("✅ Format compatibility tests completed")
	})
}

func TestErrorHandling(t *testing.T) {
	fmt.Println("\n=== Error Handling Tests ===")
	
	tests := []struct {
		name      string
		message   []byte
		pubKey    []byte
		signature []byte
		protocol  uint32
		curve     uint32
		expectErr bool
	}{
		{
			name:      "Invalid curve ID",
			message:   []byte("test"),
			pubKey:    make([]byte, 32),
			signature: make([]byte, 64),
			protocol:  0,
			curve:     9999,
			expectErr: true,
		},
		{
			name:      "Invalid ED25519 key size",
			message:   []byte("test"),
			pubKey:    make([]byte, 31), // Should be 32
			signature: make([]byte, 64),
			protocol:  0,
			curve:     constants.CurveED25519,
			expectErr: true,
		},
		{
			name:      "Invalid ED25519 signature size",
			message:   []byte("test"),
			pubKey:    make([]byte, 32),
			signature: make([]byte, 63), // Should be 64
			protocol:  0,
			curve:     constants.CurveED25519,
			expectErr: true,
		},
		{
			name:      "Invalid protocol for secp256k1",
			message:   []byte("test"),
			pubKey:    make([]byte, 65),
			signature: make([]byte, 64),
			protocol:  9999,
			curve:     constants.CurveSECP256K1,
			expectErr: true,
		},
		{
			name:      "Invalid public key format for secp256k1",
			message:   []byte("test"),
			pubKey:    make([]byte, 50), // Invalid size
			signature: make([]byte, 64),
			protocol:  constants.ProtocolECDSA,
			curve:     constants.CurveSECP256K1,
			expectErr: true,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := verification.VerifySignature(tc.message, tc.pubKey, tc.signature, tc.protocol, tc.curve)
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectErr {
				fmt.Printf("✅ %s: Correctly returned error: %v\n", tc.name, err)
			}
		})
	}
	
	fmt.Println("✅ Error handling tests passed")
}

func ShowExamples() {
	fmt.Println("\n=== Example Usage ===")
	
	// Example 1: ED25519
	fmt.Println("\n1. ED25519 Example:")
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("Hello, TEENet!")
	signature := ed25519.Sign(privKey, message)
	
	valid, _ := verification.VerifySignature(message, pubKey, signature, 0, constants.CurveED25519)
	fmt.Printf("   Message: %s\n", message)
	fmt.Printf("   Public Key (hex): %x\n", pubKey)
	fmt.Printf("   Signature (hex): %x\n", signature)
	fmt.Printf("   Valid: %v\n", valid)
	
	// Example 2: SECP256K1 ECDSA
	fmt.Println("\n2. SECP256K1 ECDSA Example:")
	btcPrivKey, _ := btcec.NewPrivateKey()
	btcPubKey := btcPrivKey.PubKey()
	btcMessage := []byte("Bitcoin transaction")
	
	hasher := sha256.New()
	hasher.Write(btcMessage)
	messageHash := hasher.Sum(nil)
	btcSig := btcecdsa.Sign(btcPrivKey, messageHash)
	
	valid, _ = verification.VerifySignature(btcMessage, btcPubKey.SerializeCompressed(), 
		btcSig.Serialize(), constants.ProtocolECDSA, constants.CurveSECP256K1)
	fmt.Printf("   Message: %s\n", btcMessage)
	fmt.Printf("   Public Key (compressed hex): %x\n", btcPubKey.SerializeCompressed())
	fmt.Printf("   Signature (DER hex): %x\n", btcSig.Serialize())
	fmt.Printf("   Valid: %v\n", valid)
	
	// Example 3: SECP256R1 ECDSA
	fmt.Println("\n3. SECP256R1 (P-256) ECDSA Example:")
	p256PrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	p256Message := []byte("Web authentication")
	
	hasher = sha256.New()
	hasher.Write(p256Message)
	p256Hash := hasher.Sum(nil)
	r, s, _ := ecdsa.Sign(rand.Reader, p256PrivKey, p256Hash)
	
	rawSig := make([]byte, 64)
	r.FillBytes(rawSig[:32])
	s.FillBytes(rawSig[32:])
	
	p256PubKey := elliptic.MarshalCompressed(elliptic.P256(), p256PrivKey.X, p256PrivKey.Y)
	valid, _ = verification.VerifySignature(p256Message, p256PubKey, rawSig,
		constants.ProtocolECDSA, constants.CurveSECP256R1)
	fmt.Printf("   Message: %s\n", p256Message)
	fmt.Printf("   Public Key (compressed hex): %x\n", p256PubKey)
	fmt.Printf("   Signature (raw hex): %x\n", rawSig)
	fmt.Printf("   Valid: %v\n", valid)
}

func TestMain(m *testing.M) {
	// Run example first to show usage
	ShowExamples()
	
	// Then run tests
	m.Run()
}

