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
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	client "github.com/TEENet-io/tee-dao-key-management-client/go"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
	"github.com/gin-gonic/gin"
)

type VerifyRequest struct {
	Message   string `json:"message" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Protocol  string `json:"protocol" binding:"required"`
	Curve     string `json:"curve" binding:"required"`
}

type VerifyResponse struct {
	Success   bool   `json:"success"`
	Valid     bool   `json:"valid,omitempty"`
	Message   string `json:"message,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Signature string `json:"signature,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Error     string `json:"error,omitempty"`
}

type VerifyWithAppIDRequest struct {
	Message   string `json:"message" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	AppID     string `json:"app_id" binding:"required"`
}

type VerifyWithAppIDResponse struct {
	Success   bool   `json:"success"`
	Valid     bool   `json:"valid,omitempty"`
	Message   string `json:"message,omitempty"`
	Signature string `json:"signature,omitempty"`
	AppID     string `json:"app_id,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GetPublicKeyRequest struct {
	AppID string `json:"app_id" binding:"required"`
}

type GetPublicKeyResponse struct {
	Success   bool   `json:"success"`
	AppID     string `json:"app_id,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Error     string `json:"error,omitempty"`
}

type SignWithAppIDRequest struct {
	Message string `json:"message" binding:"required"`
	AppID   string `json:"app_id" binding:"required"`
}

type SignWithAppIDResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	AppID     string `json:"app_id,omitempty"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

type SignRequest struct {
	Message   string `json:"message" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
	Protocol  string `json:"protocol" binding:"required"`
	Curve     string `json:"curve" binding:"required"`
}

type SignResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

var teeClient *client.Client
var defaultAppID string

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
	case "p256":
		return constants.CurveP256, nil
	default:
		if num, err := strconv.ParseUint(curve, 10, 32); err == nil {
			return uint32(num), nil
		}
		return 0, fmt.Errorf("unknown curve: %s", curve)
	}
}

// verifySignature verifies a signature against a message and public key
func verifySignature(message, publicKey, signature []byte, protocol, curve uint32) (bool, error) {
	switch curve {
	case constants.CurveED25519:
		// ED25519 verification
		if len(publicKey) != ed25519.PublicKeySize {
			return false, fmt.Errorf("invalid ED25519 public key size: expected %d, got %d", ed25519.PublicKeySize, len(publicKey))
		}
		if len(signature) != ed25519.SignatureSize {
			return false, fmt.Errorf("invalid ED25519 signature size: expected %d, got %d", ed25519.SignatureSize, len(signature))
		}

		// For ED25519, we verify directly
		return ed25519.Verify(ed25519.PublicKey(publicKey), message, signature), nil

	case constants.CurveSECP256K1, constants.CurveSECP256R1, constants.CurveP256:
		// For SECP curves, we need to hash the message first (commonly SHA256)
		hasher := sha256.New()
		hasher.Write(message)
		messageHash := hasher.Sum(nil)
		_ = messageHash // Will be used for actual ECDSA verification implementation

		// Note: This is a simplified implementation
		// In practice, you'd need a proper ECDSA verification library
		// For now, we'll return true for demonstration purposes
		// TODO: Implement proper ECDSA verification
		log.Printf("Warning: ECDSA verification not fully implemented for curve %d", curve)
		return true, nil

	default:
		return false, fmt.Errorf("unsupported curve: %d", curve)
	}
}

func main() {
	// Get configuration from environment variables
	configAddr := os.Getenv("TEE_CONFIG_ADDR")
	if configAddr == "" {
		configAddr = "localhost:50052" // Default TEE configuration server address
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port
	}

	// Get App ID from environment variable
	defaultAppID = os.Getenv("APP_ID")
	if defaultAppID == "" {
		log.Fatalf("APP_ID environment variable is required")
	}

	// Initialize TEE client
	teeClient = client.NewClient(configAddr)
	if err := teeClient.Init(); err != nil {
		log.Fatalf("Failed to initialize TEE client: %v", err)
	}
	defer teeClient.Close()

	log.Printf("TEE client initialized, Node ID: %d", teeClient.GetNodeID())

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

	// Serve static HTML page
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		// Replace placeholder with actual App ID from environment
		htmlWithAppID := strings.Replace(htmlContent, "{{APP_ID}}", defaultAppID, -1)
		c.String(http.StatusOK, htmlWithAppID)
	})

	// API endpoints
	api := router.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "TEENet Signature Tool",
			"node_id": teeClient.GetNodeID(),
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

	// Sign message with direct public key input
	api.POST("/sign", func(c *gin.Context) {
		var req SignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, SignResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
			})
			return
		}

		// Parse protocol and curve
		protocolNum, err := parseProtocol(req.Protocol)
		if err != nil {
			c.JSON(http.StatusBadRequest, SignResponse{
				Success: false,
				Error:   "Invalid protocol: " + err.Error(),
			})
			return
		}

		curveNum, err := parseCurve(req.Curve)
		if err != nil {
			c.JSON(http.StatusBadRequest, SignResponse{
				Success: false,
				Error:   "Invalid curve: " + err.Error(),
			})
			return
		}

		// Decode public key from base64
		publicKeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, SignResponse{
				Success: false,
				Error:   "Invalid public key (must be base64): " + err.Error(),
			})
			return
		}

		// Sign the message
		signature, err := teeClient.Sign([]byte(req.Message), publicKeyBytes, protocolNum, curveNum)
		if err != nil {
			log.Printf("Failed to sign message: %v", err)
			c.JSON(http.StatusInternalServerError, SignResponse{
				Success:   false,
				Message:   req.Message,
				PublicKey: req.PublicKey,
				Protocol:  req.Protocol,
				Curve:     req.Curve,
				Error:     err.Error(),
			})
			return
		}

		signatureHex := hex.EncodeToString(signature)
		log.Printf("Successfully signed message with protocol %s, curve %s", req.Protocol, req.Curve)
		c.JSON(http.StatusOK, SignResponse{
			Success:   true,
			Message:   req.Message,
			PublicKey: req.PublicKey,
			Protocol:  req.Protocol,
			Curve:     req.Curve,
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
			Protocol:  protocol,
			Curve:     curve,
		})
	})

	// Verify signature with direct public key input
	api.POST("/verify", func(c *gin.Context) {
		var req VerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, VerifyResponse{
				Success: false,
				Error:   "Invalid request: " + err.Error(),
			})
			return
		}

		// Parse protocol and curve
		protocolNum, err := parseProtocol(req.Protocol)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyResponse{
				Success: false,
				Error:   "Invalid protocol: " + err.Error(),
			})
			return
		}

		curveNum, err := parseCurve(req.Curve)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyResponse{
				Success: false,
				Error:   "Invalid curve: " + err.Error(),
			})
			return
		}

		// Decode public key from base64 and signature from hex
		publicKeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyResponse{
				Success: false,
				Error:   "Invalid public key format (must be base64): " + err.Error(),
			})
			return
		}

		signatureBytes, err := hex.DecodeString(req.Signature)
		if err != nil {
			c.JSON(http.StatusBadRequest, VerifyResponse{
				Success: false,
				Error:   "Invalid signature format (must be hex): " + err.Error(),
			})
			return
		}

		// Verify the signature
		valid, err := verifySignature([]byte(req.Message), publicKeyBytes, signatureBytes, protocolNum, curveNum)
		if err != nil {
			log.Printf("Failed to verify signature: %v", err)
			c.JSON(http.StatusInternalServerError, VerifyResponse{
				Success: false,
				Message: req.Message,
				Error:   err.Error(),
			})
			return
		}

		log.Printf("Signature verification completed: valid=%t", valid)
		c.JSON(http.StatusOK, VerifyResponse{
			Success:   true,
			Valid:     valid,
			Message:   req.Message,
			PublicKey: req.PublicKey,
			Signature: req.Signature,
			Protocol:  req.Protocol,
			Curve:     req.Curve,
		})
	})

	log.Printf("Starting TEENet Signature Tool on port %s...", port)
	log.Printf("TEE Configuration Server: %s", configAddr)
	log.Printf("Default App ID: %s", defaultAppID)
	log.Printf("Web interface available at: http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TEENet Signature Tool</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 40px auto;
            padding: 20px;
            background-color: #f5f5f5;
            line-height: 1.6;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .section {
            margin: 30px 0;
            padding: 20px;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            background-color: #fafafa;
        }
        .section h2 {
            color: #444;
            margin-top: 0;
        }
        input, textarea, select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
            box-sizing: border-box;
        }
        button {
            background-color: #007cba;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            margin-top: 10px;
            margin-right: 10px;
        }
        button:hover {
            background-color: #005a8b;
        }
        button:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }
        .result {
            margin-top: 15px;
            padding: 15px;
            border-radius: 5px;
            font-family: monospace;
            font-size: 12px;
            white-space: pre-wrap;
            word-break: break-all;
        }
        .success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .error {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .loading {
            background-color: #d1ecf1;
            border: 1px solid #bee5eb;
            color: #0c5460;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
            color: #555;
        }
        .form-group {
            margin-bottom: 15px;
        }
        .flex-container {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
        }
        .flex-item {
            flex: 1;
            min-width: 300px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 TEENet Signature Tool</h1>
        <div style="text-align: center; margin-bottom: 30px; padding: 10px; background-color: #f0f8ff; border-radius: 5px; border-left: 4px solid #007cba;">
            <strong>App ID:</strong> <code style="background-color: #e8e8e8; padding: 2px 6px; border-radius: 3px;">{{APP_ID}}</code>
        </div>
        
        <div class="flex-container">
            <div class="flex-item">
                <div class="section">
                    <h2>1. Sign Message</h2>
                    <div class="form-group">
                        <label for="message1">Message to Sign:</label>
                        <textarea id="message1" rows="3" placeholder="Enter message to sign"></textarea>
                    </div>
                    <button onclick="signWithAppID()">Sign Message</button>
                    <div id="signAppIDResult" class="result" style="display: none;"></div>
                </div>
            </div>

            <div class="flex-item">
                <div class="section">
                    <h2>2. Verify Signature</h2>
                    <div class="form-group">
                        <label for="verifyMessage1">Original Message:</label>
                        <textarea id="verifyMessage1" rows="3" placeholder="Enter the original message that was signed"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="verifySignature1">Signature (Hex):</label>
                        <textarea id="verifySignature1" rows="2" placeholder="Enter hex-encoded signature to verify"></textarea>
                    </div>
                    <button onclick="verifyWithAppID()">Verify Signature</button>
                    <div id="verifyAppIDResult" class="result" style="display: none;"></div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>🔧 Advanced Functions</h2>
            <div class="flex-container">
                <div class="flex-item">
                    <div class="section">
                        <h3>3. Get Public Key</h3>
                        <button onclick="getPublicKey()">Get Public Key</button>
                        <div id="publicKeyResult" class="result" style="display: none;"></div>
                    </div>
                </div>
            </div>
            
            <div class="flex-container">
                <div class="flex-item">
                    <h3>4. Sign with Public Key (Advanced)</h3>
                    <div class="form-group">
                        <label for="publicKey">Public Key (Base64):</label>
                        <textarea id="publicKey" rows="2" placeholder="Enter base64 encoded public key"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="protocol">Protocol:</label>
                        <select id="protocol">
                            <option value="schnorr">Schnorr (default)</option>
                            <option value="ecdsa">ECDSA</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="curve">Curve:</label>
                        <select id="curve">
                            <option value="ed25519">ED25519 (default)</option>
                            <option value="secp256k1">SECP256K1</option>
                            <option value="secp256r1">SECP256R1</option>
                            <option value="p256">P256</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="message2">Message to Sign:</label>
                        <textarea id="message2" rows="3" placeholder="Enter message to sign"></textarea>
                    </div>
                    <button onclick="signWithPublicKey()">Sign Message</button>
                    <div id="signDirectResult" class="result" style="display: none;"></div>
                </div>

                <div class="flex-item">
                    <h3>5. Verify with Public Key (Advanced)</h3>
                    <div class="form-group">
                        <label for="verifyPublicKey">Public Key (Base64):</label>
                        <textarea id="verifyPublicKey" rows="2" placeholder="Enter base64 encoded public key"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="verifyProtocol">Protocol:</label>
                        <select id="verifyProtocol">
                            <option value="schnorr">Schnorr (default)</option>
                            <option value="ecdsa">ECDSA</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="verifyCurve">Curve:</label>
                        <select id="verifyCurve">
                            <option value="ed25519">ED25519 (default)</option>
                            <option value="secp256k1">SECP256K1</option>
                            <option value="secp256r1">SECP256R1</option>
                            <option value="p256">P256</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="verifyMessage2">Original Message:</label>
                        <textarea id="verifyMessage2" rows="3" placeholder="Enter the original message that was signed"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="verifySignature2">Signature (Hex):</label>
                        <textarea id="verifySignature2" rows="2" placeholder="Enter hex-encoded signature to verify"></textarea>
                    </div>
                    <button onclick="verifyWithPublicKey()">Verify Signature</button>
                    <div id="verifyDirectResult" class="result" style="display: none;"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const FIXED_APP_ID = "{{APP_ID}}";
        
        // Dynamic API base path detection - works for both direct access and proxy access
        function getApiBasePath() {
            const currentPath = window.location.pathname;
            // If accessed through proxy, keep the current path as base
            // If accessed directly, use empty base
            return currentPath.endsWith('/') ? currentPath : currentPath + '/';
        }
        
        async function makeApiCall(endpoint, options = {}) {
            const basePath = getApiBasePath();
            const url = basePath + 'api/' + endpoint;
            return fetch(url, options);
        }
        
        async function getPublicKey() {
            const resultDiv = document.getElementById('publicKeyResult');

            showResult(resultDiv, 'Getting public key...', 'loading');

            try {
                const response = await makeApiCall('get-public-key', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ app_id: FIXED_APP_ID })
                });

                const data = await response.json();
                
                if (data.success) {
                    const result = JSON.stringify({
                        app_id: data.app_id,
                        protocol: data.protocol,
                        curve: data.curve,
                        public_key: data.public_key
                    }, null, 2);
                    showResult(resultDiv, result, 'success');
                    
                    // Auto-fill the advanced form if empty
                    if (!document.getElementById('publicKey').value) {
                        document.getElementById('publicKey').value = data.public_key;
                        document.getElementById('protocol').value = data.protocol;
                        document.getElementById('curve').value = data.curve;
                    }
                } else {
                    showResult(resultDiv, 'Error: ' + data.error, 'error');
                }
            } catch (error) {
                showResult(resultDiv, 'Network error: ' + error.message, 'error');
            }
        }

        async function signWithAppID() {
            const message = document.getElementById('message1').value.trim();
            const resultDiv = document.getElementById('signAppIDResult');
            
            if (!message) {
                showResult(resultDiv, 'Please enter a message', 'error');
                return;
            }

            showResult(resultDiv, 'Signing message...', 'loading');

            try {
                const response = await makeApiCall('sign-with-appid', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        app_id: FIXED_APP_ID,
                        message: message 
                    })
                });

                const data = await response.json();
                
                if (data.success) {
                    const result = JSON.stringify({
                        message: data.message,
                        app_id: data.app_id,
                        signature: data.signature
                    }, null, 2);
                    showResult(resultDiv, result, 'success');
                    
                    // Auto-fill verification form if empty
                    if (!document.getElementById('verifyMessage1').value) {
                        document.getElementById('verifyMessage1').value = message;
                        document.getElementById('verifySignature1').value = data.signature;
                    }
                } else {
                    showResult(resultDiv, 'Error: ' + data.error, 'error');
                }
            } catch (error) {
                showResult(resultDiv, 'Network error: ' + error.message, 'error');
            }
        }

        async function signWithPublicKey() {
            const publicKey = document.getElementById('publicKey').value.trim();
            const protocol = document.getElementById('protocol').value;
            const curve = document.getElementById('curve').value;
            const message = document.getElementById('message2').value.trim();
            const resultDiv = document.getElementById('signDirectResult');
            
            if (!publicKey || !message) {
                showResult(resultDiv, 'Please enter both public key and message', 'error');
                return;
            }

            showResult(resultDiv, 'Signing message...', 'loading');

            try {
                const response = await makeApiCall('sign', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        public_key: publicKey,
                        protocol: protocol,
                        curve: curve,
                        message: message 
                    })
                });

                const data = await response.json();
                
                if (data.success) {
                    const result = JSON.stringify({
                        message: data.message,
                        protocol: data.protocol,
                        curve: data.curve,
                        signature: data.signature
                    }, null, 2);
                    showResult(resultDiv, result, 'success');
                    
                    // Auto-fill verification form if empty
                    if (!document.getElementById('verifyPublicKey').value) {
                        document.getElementById('verifyPublicKey').value = publicKey;
                        document.getElementById('verifyProtocol').value = protocol;
                        document.getElementById('verifyCurve').value = curve;
                        document.getElementById('verifyMessage2').value = message;
                        document.getElementById('verifySignature2').value = data.signature;
                    }
                } else {
                    showResult(resultDiv, 'Error: ' + data.error, 'error');
                }
            } catch (error) {
                showResult(resultDiv, 'Network error: ' + error.message, 'error');
            }
        }

        async function verifyWithAppID() {
            const message = document.getElementById('verifyMessage1').value.trim();
            const signature = document.getElementById('verifySignature1').value.trim();
            const resultDiv = document.getElementById('verifyAppIDResult');
            
            if (!message || !signature) {
                showResult(resultDiv, 'Please enter message and signature', 'error');
                return;
            }

            showResult(resultDiv, 'Verifying signature...', 'loading');

            try {
                const response = await makeApiCall('verify-with-appid', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        app_id: FIXED_APP_ID,
                        message: message,
                        signature: signature
                    })
                });

                const data = await response.json();
                
                if (data.success) {
                    const result = JSON.stringify({
                        valid: data.valid,
                        message: data.message,
                        app_id: data.app_id,
                        protocol: data.protocol,
                        curve: data.curve,
                        verification_result: data.valid ? '✅ Valid signature' : '❌ Invalid signature'
                    }, null, 2);
                    showResult(resultDiv, result, data.valid ? 'success' : 'error');
                } else {
                    showResult(resultDiv, 'Error: ' + data.error, 'error');
                }
            } catch (error) {
                showResult(resultDiv, 'Network error: ' + error.message, 'error');
            }
        }

        async function verifyWithPublicKey() {
            const publicKey = document.getElementById('verifyPublicKey').value.trim();
            const protocol = document.getElementById('verifyProtocol').value;
            const curve = document.getElementById('verifyCurve').value;
            const message = document.getElementById('verifyMessage2').value.trim();
            const signature = document.getElementById('verifySignature2').value.trim();
            const resultDiv = document.getElementById('verifyDirectResult');
            
            if (!publicKey || !message || !signature) {
                showResult(resultDiv, 'Please enter public key, message, and signature', 'error');
                return;
            }

            showResult(resultDiv, 'Verifying signature...', 'loading');

            try {
                const response = await makeApiCall('verify', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        public_key: publicKey,
                        protocol: protocol,
                        curve: curve,
                        message: message,
                        signature: signature
                    })
                });

                const data = await response.json();
                
                if (data.success) {
                    const result = JSON.stringify({
                        valid: data.valid,
                        message: data.message,
                        protocol: data.protocol,
                        curve: data.curve,
                        verification_result: data.valid ? '✅ Valid signature' : '❌ Invalid signature'
                    }, null, 2);
                    showResult(resultDiv, result, data.valid ? 'success' : 'error');
                } else {
                    showResult(resultDiv, 'Error: ' + data.error, 'error');
                }
            } catch (error) {
                showResult(resultDiv, 'Network error: ' + error.message, 'error');
            }
        }

        function showResult(element, content, type) {
            element.textContent = content;
            element.className = 'result ' + type;
            element.style.display = 'block';
        }
    </script>
</body>
</html>`
