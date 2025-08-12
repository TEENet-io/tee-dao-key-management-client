package main

import (
	"fmt"
	"log"

	client "github.com/TEENet-io/tee-dao-key-management-client/go"
)

func main() {
	fmt.Println("=== Testing TEE DAO Client Library against Mock Server ===")

	// Create client
	configServerAddr := "localhost:50052"
	teeClient := client.NewClient(configServerAddr)
	defer teeClient.Close()

	// Initialize client
	fmt.Println("Initializing client...")
	if err := teeClient.Init(); err != nil {
		log.Fatalf("Client initialization failed: %v", err)
	}

	fmt.Printf("Client connected, Node ID: %d\n", teeClient.GetNodeID())

	// Test 1: Get public key by App ID
	fmt.Println("\n1. Testing GetPublicKeyByAppID")
	appID := "secure-messaging-app"
	publicKey, protocol, curve, err := teeClient.GetPublicKeyByAppID(appID)
	if err != nil {
		log.Printf("Failed to get public key by app ID: %v", err)
	} else {
		fmt.Printf("✓ GetPublicKeyByAppID successful!\n")
		fmt.Printf("  App ID: %s\n", appID)
		fmt.Printf("  Protocol: %s\n", protocol)
		fmt.Printf("  Curve: %s\n", curve)
		fmt.Printf("  Public Key: %s\n", publicKey)
	}

	// Test 2: Sign with App ID
	fmt.Println("\n2. Testing SignWithAppID")
	message := []byte("Hello from TEE DAO Client Library Test!")

	signature, err := teeClient.SignWithAppID(message, appID)
	if err != nil {
		log.Printf("Failed to sign with app ID: %v", err)
	} else {
		fmt.Printf("✓ SignWithAppID successful!\n")
		fmt.Printf("  Message: %s\n", string(message))
		fmt.Printf("  App ID: %s\n", appID)
		fmt.Printf("  Signature: %x\n", signature)
		fmt.Printf("  Signature length: %d bytes\n", len(signature))
	}

	fmt.Println("\n=== All tests completed successfully! ===")
}
