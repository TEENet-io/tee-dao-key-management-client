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

	client "github.com/TEENet-io/tee-dao-key-management-client/go"
)

func main() {
	// Configuration
	configServerAddr := "localhost:50052" // TEE config server address

	fmt.Println("=== TEE DAO Key Management Client with AppID Service Integration ===")

	// Create client
	client := client.NewClient(configServerAddr)
	defer client.Close()

	if err := client.Init(nil); err != nil {
		log.Fatalf("Client initialization failed: %v", err)
	}

	fmt.Println("Client initialized successfully")

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

	// Example: Multi-party voting signature
	fmt.Println("\n3. Multi-party voting signature")
	targetAppIDs := []string{"secure-messaging-app", "secure-messaging-app1", "secure-messaging-app2"}
	requiredVotes := 2
	votingMessage := []byte("test message for multi-party voting") // Contains "test" to trigger approval

	// Create request data similar to signature-tool
	messageBase64 := base64.StdEncoding.EncodeToString(votingMessage)
	requestData := map[string]interface{}{
		"message":        messageBase64,
		"signer_app_id":  appID,
		"target_app_ids": targetAppIDs,
		"required_votes": requiredVotes,
		"is_forwarded":   false,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("Failed to create request body: %v", err)
		return
	}

	// Create a mock HTTP request like signature-tool does
	req, err := http.NewRequest("POST", "/vote", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return
	}

	// Set headers like signature-tool
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TEE-DAO-Client/1.0")

	// Local approval decision (same logic as signature-tool)
	localApproval := strings.Contains(strings.ToLower(string(votingMessage)), "test")

	fmt.Printf("Voting request:\n")
	fmt.Printf("  - Message: %s\n", string(votingMessage))
	fmt.Printf("  - Signer App ID: %s\n", appID)
	fmt.Printf("  - Target App IDs: %v\n", targetAppIDs)
	fmt.Printf("  - Required Votes: %d/%d\n", requiredVotes, len(targetAppIDs))
	fmt.Printf("  - Local Approval: %t\n", localApproval)

	// Use VotingSign with the constructed HTTP request
	votingResult, err := client.VotingSign(req, votingMessage, appID, targetAppIDs, requiredVotes, localApproval)
	if err != nil {
		log.Printf("Voting signature failed: %v", err)
	} else {
		fmt.Printf("\nVoting signature completed!\n")
		fmt.Printf("Total Targets: %d\n", votingResult.TotalTargets)
		fmt.Printf("Successful Votes: %d/%d\n", votingResult.SuccessfulVotes, votingResult.RequiredVotes)
		fmt.Printf("Voting Complete: %t\n", votingResult.VotingComplete)
		fmt.Printf("Final Result: %s\n", votingResult.FinalResult)

		if votingResult.Signature != nil {
			fmt.Printf("Signature: %x\n", votingResult.Signature)
		} else {
			fmt.Printf("Signature: No signature (voting failed or incomplete)\n")
		}

		// Print detailed vote results
		fmt.Printf("\nVote Details:\n")
		for i, detail := range votingResult.VoteDetails {
			status := "FAILED"
			if detail.Success && detail.Response {
				status = "APPROVED"
			} else if detail.Success && !detail.Response {
				status = "REJECTED"
			}
			fmt.Printf("  %d. %s: %s", i+1, detail.ClientID, status)
			if detail.Error != "" {
				fmt.Printf(" (Error: %s)", detail.Error)
			}
			fmt.Printf("\n")
		}
	}

	fmt.Println("\n=== Example completed ===")
}
