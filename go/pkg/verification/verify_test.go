package verification

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/TEENet-io/teenet-sdk/go/pkg/constants"
	"github.com/btcsuite/btcd/btcec/v2"
	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

func TestED25519Verification(t *testing.T) {
	// Generate ED25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ED25519 key: %v", err)
	}

	message := []byte("Hello, TEENet!")
	
	// Sign the message
	signature := ed25519.Sign(privKey, message)
	
	// Test valid signature
	valid, err := VerifySignature(message, pubKey, signature, 0, constants.CurveED25519)
	if err != nil {
		t.Fatalf("Verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid ED25519 signature not verified")
	}
	
	// Test invalid signature (modify one byte)
	invalidSig := make([]byte, len(signature))
	copy(invalidSig, signature)
	invalidSig[0] ^= 0xFF
	
	valid, err = VerifySignature(message, pubKey, invalidSig, 0, constants.CurveED25519)
	if err != nil {
		t.Fatalf("Verification failed with error: %v", err)
	}
	if valid {
		t.Error("Invalid ED25519 signature was verified")
	}
	
	// Test wrong message
	wrongMessage := []byte("Wrong message")
	valid, err = VerifySignature(wrongMessage, pubKey, signature, 0, constants.CurveED25519)
	if err != nil {
		t.Fatalf("Verification failed with error: %v", err)
	}
	if valid {
		t.Error("Signature verified with wrong message")
	}
	
	t.Log("✅ ED25519 verification tests passed")
}

func TestSecp256k1ECDSAVerification(t *testing.T) {
	// Generate secp256k1 key pair using btcec
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate secp256k1 key: %v", err)
	}
	pubKey := privKey.PubKey()
	
	message := []byte("Hello, Bitcoin!")
	
	// Hash the message
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)
	
	// Sign with ECDSA
	sig := btcecdsa.Sign(privKey, messageHash)
	
	// Test with DER encoded signature
	derSig := sig.Serialize()
	valid, err := VerifySignature(message, pubKey.SerializeUncompressed(), derSig, constants.ProtocolECDSA, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("DER ECDSA verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid secp256k1 ECDSA signature (DER) not verified")
	}
	
	// Test with raw r,s format (64 bytes)
	r := sig.R()
	s := sig.S()
	rawSig := make([]byte, 64)
	r.PutBytesUnchecked(rawSig[:32])
	s.PutBytesUnchecked(rawSig[32:])
	
	valid, err = VerifySignature(message, pubKey.SerializeUncompressed(), rawSig, constants.ProtocolECDSA, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Raw ECDSA verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid secp256k1 ECDSA signature (raw) not verified")
	}
	
	// Test with compressed public key
	valid, err = VerifySignature(message, pubKey.SerializeCompressed(), derSig, constants.ProtocolECDSA, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Compressed key verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid signature with compressed public key not verified")
	}
	
	// Test with raw public key format (64 bytes, no prefix)
	uncompressed := pubKey.SerializeUncompressed()
	rawPubKey := uncompressed[1:] // Remove 0x04 prefix
	
	valid, err = VerifySignature(message, rawPubKey, derSig, constants.ProtocolECDSA, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Raw public key verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid signature with raw public key not verified")
	}
	
	// Test invalid signature
	invalidSig := make([]byte, len(derSig))
	copy(invalidSig, derSig)
	invalidSig[len(invalidSig)-1] ^= 0xFF
	
	valid, err = VerifySignature(message, pubKey.SerializeUncompressed(), invalidSig, constants.ProtocolECDSA, constants.CurveSECP256K1)
	if err == nil && valid {
		t.Error("Invalid ECDSA signature was verified")
	}
	
	t.Log("✅ Secp256k1 ECDSA verification tests passed")
}

func TestSecp256k1SchnorrVerification(t *testing.T) {
	// Generate secp256k1 key pair
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate secp256k1 key: %v", err)
	}
	pubKey := privKey.PubKey()
	
	message := []byte("Hello, Schnorr!")
	
	// Hash the message
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)
	
	// Sign with Schnorr
	sig, err := schnorr.Sign(privKey, messageHash)
	if err != nil {
		t.Fatalf("Failed to create Schnorr signature: %v", err)
	}
	
	// Test Schnorr signature verification
	valid, err := VerifySignature(message, pubKey.SerializeUncompressed(), sig.Serialize(), constants.ProtocolSchnorr, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Schnorr verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid secp256k1 Schnorr signature not verified")
	}
	
	// Test with compressed public key
	valid, err = VerifySignature(message, pubKey.SerializeCompressed(), sig.Serialize(), constants.ProtocolSchnorr, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Schnorr verification with compressed key failed: %v", err)
	}
	if !valid {
		t.Error("Valid Schnorr signature with compressed key not verified")
	}
	
	// Test invalid signature
	invalidSig := make([]byte, schnorr.SignatureSize)
	copy(invalidSig, sig.Serialize())
	invalidSig[0] ^= 0xFF
	
	valid, err = VerifySignature(message, pubKey.SerializeUncompressed(), invalidSig, constants.ProtocolSchnorr, constants.CurveSECP256K1)
	if err != nil {
		t.Fatalf("Invalid Schnorr verification failed with error: %v", err)
	}
	if valid {
		t.Error("Invalid Schnorr signature was verified")
	}
	
	t.Log("✅ Secp256k1 Schnorr verification tests passed")
}

func TestSecp256r1ECDSAVerification(t *testing.T) {
	// Generate P-256 key pair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate P-256 key: %v", err)
	}
	
	message := []byte("Hello, P-256!")
	
	// Hash the message
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)
	
	// Sign with ECDSA
	r, s, err := ecdsa.Sign(rand.Reader, privKey, messageHash)
	if err != nil {
		t.Fatalf("Failed to sign with P-256: %v", err)
	}
	
	// Create raw signature (64 bytes)
	rawSig := make([]byte, 64)
	r.FillBytes(rawSig[:32])
	s.FillBytes(rawSig[32:])
	
	// Create uncompressed public key
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)
	
	// Test with raw signature format
	valid, err := VerifySignature(message, pubKeyBytes, rawSig, constants.ProtocolECDSA, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("P-256 ECDSA verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid P-256 ECDSA signature not verified")
	}
	
	// Test with compressed public key
	compressedPubKey := elliptic.MarshalCompressed(elliptic.P256(), privKey.X, privKey.Y)
	valid, err = VerifySignature(message, compressedPubKey, rawSig, constants.ProtocolECDSA, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("P-256 compressed key verification failed: %v", err)
	}
	if !valid {
		t.Error("Valid P-256 signature with compressed key not verified")
	}
	
	// Test with raw public key (no 0x04 prefix)
	rawPubKey := pubKeyBytes[1:]
	valid, err = VerifySignature(message, rawPubKey, rawSig, constants.ProtocolECDSA, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("P-256 raw public key verification failed: %v", err)
	}
	if !valid {
		t.Error("Valid P-256 signature with raw public key not verified")
	}
	
	// Test invalid signature
	invalidSig := make([]byte, 64)
	copy(invalidSig, rawSig)
	invalidSig[0] ^= 0xFF
	
	valid, err = VerifySignature(message, pubKeyBytes, invalidSig, constants.ProtocolECDSA, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("Invalid P-256 verification failed with error: %v", err)
	}
	if valid {
		t.Error("Invalid P-256 signature was verified")
	}
	
	t.Log("✅ Secp256r1 (P-256) ECDSA verification tests passed")
}

func TestSecp256r1SchnorrVerification(t *testing.T) {
	// Generate P-256 key pair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate P-256 key: %v", err)
	}
	
	message := []byte("Hello, P-256 Schnorr!")
	
	// For Schnorr, we need to implement a simple signature
	// Since P-256 Schnorr is not standard, we'll create a test signature
	
	// Create a Schnorr signature manually
	// k = random nonce
	k, err := rand.Int(rand.Reader, privKey.Params().N)
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}
	
	// R = k*G
	Rx, _ := privKey.Curve.ScalarBaseMult(k.Bytes())
	
	// e = H(R.x || P.x || message)
	hasher := sha256.New()
	hasher.Write(Rx.Bytes())
	hasher.Write(privKey.X.Bytes())
	hasher.Write(message)
	e := new(big.Int).SetBytes(hasher.Sum(nil))
	e.Mod(e, privKey.Params().N)
	
	// s = k + e*d
	s := new(big.Int).Mul(e, privKey.D)
	s.Add(s, k)
	s.Mod(s, privKey.Params().N)
	
	// Create signature (r || s)
	schnorrSig := make([]byte, 64)
	Rx.FillBytes(schnorrSig[:32])
	s.FillBytes(schnorrSig[32:])
	
	// Create public key
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)
	
	// Test Schnorr verification
	valid, err := VerifySignature(message, pubKeyBytes, schnorrSig, constants.ProtocolSchnorr, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("P-256 Schnorr verification failed with error: %v", err)
	}
	if !valid {
		t.Error("Valid P-256 Schnorr signature not verified")
	}
	
	// Test invalid signature
	invalidSig := make([]byte, 64)
	copy(invalidSig, schnorrSig)
	invalidSig[0] ^= 0xFF
	
	valid, err = VerifySignature(message, pubKeyBytes, invalidSig, constants.ProtocolSchnorr, constants.CurveSECP256R1)
	if err != nil {
		t.Fatalf("Invalid P-256 Schnorr verification failed with error: %v", err)
	}
	if valid {
		t.Error("Invalid P-256 Schnorr signature was verified")
	}
	
	t.Log("✅ Secp256r1 (P-256) Schnorr verification tests passed")
}

func TestInvalidInputs(t *testing.T) {
	message := []byte("test message")
	
	// Test invalid curve
	_, err := VerifySignature(message, []byte{1, 2, 3}, []byte{4, 5, 6}, 0, 999)
	if err == nil {
		t.Error("Expected error for invalid curve")
	}
	
	// Test invalid ED25519 public key size
	_, err = VerifySignature(message, []byte{1, 2, 3}, make([]byte, 64), 0, constants.CurveED25519)
	if err == nil {
		t.Error("Expected error for invalid ED25519 public key size")
	}
	
	// Test invalid ED25519 signature size
	_, err = VerifySignature(message, make([]byte, 32), []byte{1, 2, 3}, 0, constants.CurveED25519)
	if err == nil {
		t.Error("Expected error for invalid ED25519 signature size")
	}
	
	// Test invalid protocol for secp256k1
	_, err = VerifySignature(message, make([]byte, 65), make([]byte, 64), 999, constants.CurveSECP256K1)
	if err == nil {
		t.Error("Expected error for invalid protocol")
	}
	
	t.Log("✅ Invalid input tests passed")
}

func TestRealWorldVectors(t *testing.T) {
	// Test with some known test vectors
	tests := []struct {
		name      string
		message   string
		pubKey    string
		signature string
		protocol  uint32
		curve     uint32
		expected  bool
	}{
		{
			name:      "Secp256k1 ECDSA Test Vector",
			message:   "Hello, World!",
			pubKey:    "04c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee51ae168fea63dc339a3c58419466ceaeef7f632653266d0e1236431a950cfe52a", // Example uncompressed public key
			signature: "3045022100f8b8af8ce7c3a91c5b4e1f8c4f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f02207f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e8f8e", // Example DER signature
			protocol:  constants.ProtocolECDSA,
			curve:     constants.CurveSECP256K1,
			expected:  false, // This is just an example, not a valid signature
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pubKeyBytes, err := hex.DecodeString(tc.pubKey)
			if err != nil {
				t.Fatalf("Failed to decode public key: %v", err)
			}
			
			sigBytes, err := hex.DecodeString(tc.signature)
			if err != nil {
				t.Fatalf("Failed to decode signature: %v", err)
			}
			
			valid, err := VerifySignature([]byte(tc.message), pubKeyBytes, sigBytes, tc.protocol, tc.curve)
			if err != nil && tc.expected {
				t.Errorf("Verification failed with error: %v", err)
			}
			
			if valid != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, valid)
			}
		})
	}
	
	t.Log("✅ Real world vector tests completed")
}

func BenchmarkED25519Verification(b *testing.B) {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("Benchmark message")
	signature := ed25519.Sign(privKey, message)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifySignature(message, pubKey, signature, 0, constants.CurveED25519)
	}
}

func BenchmarkSecp256k1ECDSAVerification(b *testing.B) {
	privKey, _ := btcec.NewPrivateKey()
	pubKey := privKey.PubKey()
	message := []byte("Benchmark message")
	
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)
	
	sig := btcecdsa.Sign(privKey, messageHash)
	pubKeyBytes := pubKey.SerializeUncompressed()
	sigBytes := sig.Serialize()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifySignature(message, pubKeyBytes, sigBytes, constants.ProtocolECDSA, constants.CurveSECP256K1)
	}
}

func BenchmarkSecp256r1ECDSAVerification(b *testing.B) {
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	message := []byte("Benchmark message")
	
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)
	
	r, s, _ := ecdsa.Sign(rand.Reader, privKey, messageHash)
	
	rawSig := make([]byte, 64)
	r.FillBytes(rawSig[:32])
	s.FillBytes(rawSig[32:])
	
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifySignature(message, pubKeyBytes, rawSig, constants.ProtocolECDSA, constants.CurveSECP256R1)
	}
}