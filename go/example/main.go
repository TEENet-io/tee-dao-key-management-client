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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	client "github.com/TEENet-io/teenet-sdk/go"
)

func main() {
	// Configuration
	configServerAddr := "localhost:50052" // TEE config server address

	fmt.Println("=== TEE DAO Key Management Client with AppID Service Integration ===")

	// Create client
	teeClient := client.NewClient(configServerAddr)
	defer teeClient.Close()

	if err := teeClient.Init(nil); err != nil {
		log.Fatalf("Client initialization failed: %v", err)
	}

	fmt.Println("Client initialized successfully")

	// Example: Get public key by app ID
	fmt.Println("\n1. Get public key by app ID")
	appID := "secure-messaging-app"
	publicKey, protocol, curve, err := teeClient.GetPublicKeyByAppID(appID)
	if err != nil {
		log.Printf("Failed to get public key by app ID: %v", err)
	} else {
		fmt.Printf("Public key for app ID %s:\n", appID)
		fmt.Printf("  - Protocol: %s\n", protocol)
		fmt.Printf("  - Curve: %s\n", curve)
		fmt.Printf("  - Public Key: %s\n", publicKey)
	}

	// Example: Sign message using Sign method
	fmt.Println("\n2. Sign message")
	message := []byte("Hello from AppID Service!")

	signReq := &client.SignRequest{
		Message: message,
		AppID:   appID,
	}
	signResult, err := teeClient.Sign(signReq)
	if err != nil {
		log.Printf("Signing failed: %v", err)
	} else {
		fmt.Printf("Signing successful!\n")
		fmt.Printf("Message: %s\n", string(message))
		fmt.Printf("Signature: %x\n", signResult.Signature)
		fmt.Printf("Success: %t\n", signResult.Success)
		if signResult.Error != "" {
			fmt.Printf("Error: %s\n", signResult.Error)
		}
	}

	// Example: Multi-party voting signature
	fmt.Println("\n3. Multi-party voting signature example")
	votingMessage := []byte("test message for multi-party voting") // Contains "test" to trigger approval

	fmt.Printf("Voting request:\n")
	fmt.Printf("  - Message: %s\n", string(votingMessage))
	fmt.Printf("  - Signer App ID: %s\n", appID)
	fmt.Printf("  - Voting Enabled: true\n")

	// Create HTTP request body similar to signature-tool
	requestData := map[string]interface{}{
		"message":       base64.StdEncoding.EncodeToString(votingMessage),
		"signer_app_id": appID,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("Failed to create request body: %v", err)
		return
	}

	// Create a mock HTTP request like signature-tool does
	httpReq, err := http.NewRequest("POST", "/vote", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Make vote decision: approve if message contains "test"
	localApproval := strings.Contains(strings.ToLower(string(votingMessage)), "test")
	fmt.Printf("  - Local Approval: %t\n", localApproval)

	// Sign with voting enabled
	votingSignReq := &client.SignRequest{
		Message:       votingMessage,
		AppID:         appID,
		EnableVoting:  true,
		LocalApproval: localApproval,
		HTTPRequest:   httpReq,
	}

	votingSignResult, err := teeClient.Sign(votingSignReq)
	if err != nil {
		log.Printf("Voting signature failed: %v", err)
	} else {
		fmt.Printf("\nVoting signature completed!\n")
		fmt.Printf("Success: %t\n", votingSignResult.Success)
		if votingSignResult.Signature != nil {
			fmt.Printf("Signature: %x\n", votingSignResult.Signature)
		}

		// Display voting information if available
		if votingSignResult.VotingInfo != nil {
			fmt.Printf("\nVoting Details:\n")
			fmt.Printf("  - Total Targets: %d\n", votingSignResult.VotingInfo.TotalTargets)
			fmt.Printf("  - Successful Votes: %d\n", votingSignResult.VotingInfo.SuccessfulVotes)
			fmt.Printf("  - Required Votes: %d\n", votingSignResult.VotingInfo.RequiredVotes)

			fmt.Printf("\nIndividual Votes:\n")
			for i, vote := range votingSignResult.VotingInfo.VoteDetails {
				fmt.Printf("  %d. Client %s: Success=%t\n", i+1, vote.ClientID, vote.Success)
			}
		}

		if votingSignResult.Error != "" {
			fmt.Printf("Error: %s\n", votingSignResult.Error)
		}
	}

	fmt.Println("\n=== Example completed ===")
}
