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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	client "github.com/TEENet-io/tee-dao-key-management-client/go"
	"github.com/gin-gonic/gin"
)

var teeClient *client.Client
var defaultAppID string

func main() {
	// Get configuration from environment variables
	configAddr := os.Getenv("TEE_CONFIG_ADDR")
	if configAddr == "" {
		configAddr = "localhost:50052" // Default TEE configuration server address
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Get App ID from environment variable
	defaultAppID = os.Getenv("APP_ID")
	if defaultAppID == "" {
		log.Fatalf("APP_ID environment variable is required")
	}

	// Frontend path
	frontendPath := os.Getenv("FRONTEND_PATH")
	if frontendPath == "" {
		frontendPath = "./frontend" // Default frontend path
	}

	// Initialize TEE client with custom voting handler
	teeClient = client.NewClient(configAddr)
	votingHandler := createVotingHandler(defaultAppID)
	if err := teeClient.Init(votingHandler); err != nil {
		log.Fatalf("Failed to initialize TEE client: %v", err)
	}
	defer teeClient.Close()

	log.Printf("TEE client initialized successfully with custom voting handler for app ID: %s", defaultAppID)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Add static file handler for frontend
	router.Use(staticFileHandler(frontendPath))

	// API endpoints
	api := router.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "TEENet Signature Tool",
		})
	})

	// Configuration endpoint for frontend
	api.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app_id": defaultAppID,
		})
	})

	// Get public key by app ID
	api.POST("/get-public-key", func(c *gin.Context) {
		var req GetPublicKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, GetPublicKeyResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
			})
			return
		}

		publicKey, protocol, curve, err := teeClient.GetPublicKeyByAppID(req.AppID)
		if err != nil {
			log.Printf("Failed to get public key for app ID %s: %v", req.AppID, err)
			c.JSON(http.StatusInternalServerError, GetPublicKeyResponse{
				Success: false,
				AppID:   req.AppID,
				Error:   err.Error(),
			})
			return
		}

		log.Printf("Successfully retrieved public key for app ID %s", req.AppID)
		c.JSON(http.StatusOK, GetPublicKeyResponse{
			Success:   true,
			AppID:     req.AppID,
			PublicKey: publicKey,
			Protocol:  protocol,
			Curve:     curve,
		})
	})

	// Sign message with app ID
	api.POST("/sign-with-appid", func(c *gin.Context) {
		var req SignWithAppIDRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, SignWithAppIDResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
			})
			return
		}

		signature, err := teeClient.SignWithAppID([]byte(req.Message), req.AppID)
		if err != nil {
			log.Printf("Failed to sign message with app ID %s: %v", req.AppID, err)
			c.JSON(http.StatusInternalServerError, SignWithAppIDResponse{
				Success: false,
				Message: req.Message,
				AppID:   req.AppID,
				Error:   err.Error(),
			})
			return
		}

		signatureHex := hex.EncodeToString(signature)
		log.Printf("Successfully signed message with app ID %s", req.AppID)
		c.JSON(http.StatusOK, SignWithAppIDResponse{
			Success:   true,
			Message:   req.Message,
			AppID:     req.AppID,
			Signature: signatureHex,
		})
	})

	// Verify signature with App ID
	api.POST("/verify-with-appid", func(c *gin.Context) {
		var req VerifyWithAppIDRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, VerifyWithAppIDResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
			})
			return
		}

		// Get public key by app ID
		publicKey, protocol, curve, err := teeClient.GetPublicKeyByAppID(req.AppID)
		if err != nil {
			log.Printf("Failed to get public key for app ID %s: %v", req.AppID, err)
			c.JSON(http.StatusInternalServerError, VerifyWithAppIDResponse{
				Success: false,
				AppID:   req.AppID,
				Error:   err.Error(),
			})
			return
		}

		// Parse protocol and curve
		protocolNum, err := parseProtocol(protocol)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyWithAppIDResponse{
				Success: false,
				Error:   "Invalid protocol: " + err.Error(),
			})
			return
		}

		curveNum, err := parseCurve(curve)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyWithAppIDResponse{
				Success: false,
				Error:   "Invalid curve: " + err.Error(),
			})
			return
		}

		// Decode public key and signature from hex/base64
		publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyWithAppIDResponse{
				Success: false,
				Error:   "Invalid public key format: " + err.Error(),
			})
			return
		}

		signatureBytes, err := hex.DecodeString(req.Signature)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyWithAppIDResponse{
				Success: false,
				Error:   "Invalid signature format (must be hex): " + err.Error(),
			})
			return
		}

		// Verify the signature
		valid, err := verifySignature([]byte(req.Message), publicKeyBytes, signatureBytes, protocolNum, curveNum)
		if err != nil {
			log.Printf("Failed to verify signature: %v", err)
			c.JSON(http.StatusInternalServerError, VerifyWithAppIDResponse{
				Success: false,
				Message: req.Message,
				AppID:   req.AppID,
				Error:   err.Error(),
			})
			return
		}

		log.Printf("Signature verification completed for app ID %s: valid=%t", req.AppID, valid)
		c.JSON(http.StatusOK, VerifyWithAppIDResponse{
			Success:   true,
			Valid:     valid,
			Message:   req.Message,
			Signature: req.Signature,
			AppID:     req.AppID,
			PublicKey: publicKey,
			Protocol:  protocol,
			Curve:     curve,
		})
	})

	// Voting endpoint - make decision and run VotingSign
	api.POST("/vote", func(c *gin.Context) {
		// Read raw request body
		requestBody, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		
		var req IncomingVoteRequest
		if err := json.Unmarshal(requestBody, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		log.Printf("🗳️  [%s] Received vote request", defaultAppID)

		// Decode message
		messageBytes, err := base64.StdEncoding.DecodeString(req.Message)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message encoding"})
			return
		}

		// Make vote decision: approve if message contains "test"
		messageStr := string(messageBytes)
		localApproval := strings.Contains(strings.ToLower(messageStr), "test")
		
		log.Printf("📝 [%s] Local vote decision for message '%s': %t", defaultAppID, messageStr, localApproval)

		// Restore request body for VotingSign to read
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		
		// Pass HTTP request to VotingSign - it will extract headers and request body automatically
		// Use req.SignerAppID as the one requesting signature
		// Target App IDs and required votes are now fetched from server configuration
		votingResult, err := teeClient.VotingSign(c.Request, messageBytes, req.SignerAppID, localApproval)
		if err != nil {
			log.Printf("❌ [%s] VotingSign failed: %v", defaultAppID, err)
			
			// Check if we have partial voting results (e.g., voting passed but signature failed)
			if votingResult != nil {
				c.JSON(http.StatusOK, gin.H{
					"success":  true,
					"approved": false,
					"app_id":   defaultAppID,
					"message":  fmt.Sprintf("VotingSign failed: %v", err),
					"voting_results": gin.H{
						"voting_complete":  votingResult.VotingComplete,
						"successful_votes": votingResult.SuccessfulVotes,
						"required_votes":   votingResult.RequiredVotes,
						"total_targets":    votingResult.TotalTargets,
						"final_result":     votingResult.FinalResult,
						"vote_details":     votingResult.VoteDetails,
						"error":           err.Error(),
					},
					"signature": "",
					"timestamp": time.Now().Format(time.RFC3339),
				})
			} else {
				// No voting results at all (e.g., configuration error)
				c.JSON(http.StatusOK, gin.H{
					"success":  true,
					"approved": false,
					"app_id":   defaultAppID,
					"message":  fmt.Sprintf("VotingSign failed: %v", err),
					"voting_results": gin.H{
						"voting_complete":  false,
						"successful_votes": 0,
						"required_votes":   0,
						"total_targets":    0,
						"final_result":     "ERROR",
						"vote_details":     []interface{}{},
						"error":           err.Error(),
					},
					"signature": "",
					"timestamp": time.Now().Format(time.RFC3339),
				})
			}
			return
		}

		finalApproval := votingResult.VotingComplete && votingResult.FinalResult == "APPROVED"
		log.Printf("✅ [%s] VotingSign result: %t", defaultAppID, finalApproval)
		
		// Convert signature to hex string if available
		var signatureHex string
		if votingResult.Signature != nil && len(votingResult.Signature) > 0 {
			signatureHex = hex.EncodeToString(votingResult.Signature)
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"approved": finalApproval,
			"app_id":   defaultAppID,
			"message":  votingResult.FinalResult,
			"voting_results": gin.H{
				"voting_complete":  votingResult.VotingComplete,
				"successful_votes": votingResult.SuccessfulVotes,
				"required_votes":   votingResult.RequiredVotes,
				"total_targets":    votingResult.TotalTargets,
				"final_result":     votingResult.FinalResult,
				"vote_details":     votingResult.VoteDetails,
			},
			"signature": signatureHex,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	log.Printf("Starting TEENet Signature Tool on port %s...", port)
	log.Printf("TEE Configuration Server: %s", configAddr)
	log.Printf("Default App ID: %s", defaultAppID)
	log.Printf("Frontend Path: %s", frontendPath)
	log.Printf("Web interface available at: http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}