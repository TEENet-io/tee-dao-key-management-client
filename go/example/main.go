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
	"fmt"
	"log"

	client "github.com/TEENet-io/tee-dao-key-management-client/go"
)

func main() {
	// Configuration
	configServerAddr := "localhost:50052" // TEE config server address

	fmt.Println("=== TEE DAO Key Management Client with AppID Service Integration ===")

	// Create client
	client := client.NewClient(configServerAddr)
	defer client.Close()

	if err := client.Init(); err != nil {
		log.Fatalf("Client initialization failed: %v", err)
	}

	fmt.Printf("Client connected, Node ID: %d\n", client.GetNodeID())

	// Example: Get public key by app ID
	fmt.Println("\n1. Get public key by app ID")
	appID := "secure-messaging-app"
	publicKey, protocol, curve, err := client.GetPublicKeyByAppID(appID)
	if err != nil {
		log.Printf("Failed to get public key by app ID: %v", err)
	} else {
		fmt.Printf("Public key for app ID %s:\n", appID)
		fmt.Printf("  - Protocol: %s\n", protocol)
		fmt.Printf("  - Curve: %s\n", curve)
		fmt.Printf("  - Public Key: %s\n", publicKey)
	}

	// Example: Sign with app ID
	fmt.Println("\n2. Sign message with app ID")
	message := []byte("Hello from AppID Service!")

	signature, err := client.SignWithAppID(message, appID)
	if err != nil {
		log.Printf("Signing with app ID failed: %v", err)
	} else {
		fmt.Printf("Signing with app ID successful!\n")
		fmt.Printf("Message: %s\n", string(message))
		fmt.Printf("Signature: %x\n", signature)
	}

	fmt.Println("\n=== Example completed ===")
}
