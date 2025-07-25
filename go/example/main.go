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
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
)

func main() {
	// 1. Create client
	client := client.NewClient("localhost:50052") // Config server address
	defer client.Close()

	// 2. Initialize (fetch config + establish TLS connection)
	if err := client.Init(); err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	fmt.Printf("Client connected, Node ID: %d\n", client.GetNodeID())

	// 3. Execute signing (client now only supports signing)
	// Note: In a real scenario, you would get the public key from elsewhere (e.g., DKG service)
	publicKey := []byte("example-public-key-from-dkg-service") // Placeholder
	message := []byte("Hello, TEE DAO!")
	
	signature, err := client.Sign(message, publicKey, constants.ProtocolECDSA, constants.CurveED25519)
	if err != nil {
		log.Fatalf("Signing failed: %v", err)
	}
	fmt.Printf("Signing successful!\n")
	fmt.Printf("Message: %s\n", string(message))
	fmt.Printf("Signature: %x\n", signature)
}
