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

// Package client provides simplified TEE DAO key management client
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/config"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/task"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/usermgmt"
)

// Client is a simplified key management client
type Client struct {
	configClient   *config.Client
	taskClient     *task.Client
	userMgmtClient *usermgmt.Client
	config         *config.NodeConfig
	timeout        time.Duration
}

// NewClient creates a new client instance with user management integration
func NewClient(configServerAddr string) *Client {
	return &Client{
		configClient: config.NewClient(configServerAddr),
		timeout:      constants.DefaultClientTimeout,
	}
}

// Init initializes client, fetches config and establishes TLS connection
func (c *Client) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 1. Fetch configuration (without TLS)
	config, err := c.configClient.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	c.config = config

	// 2. Create task client
	c.taskClient = task.NewClient(config)

	// 3. Create TLS configuration for TEE server
	teeTLSConfig, err := c.createTEETLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TEE TLS config: %w", err)
	}

	// 3. Connect to TEE server (with TLS)
	if err := c.taskClient.Connect(ctx, teeTLSConfig); err != nil {
		return fmt.Errorf("failed to connect to TEE server: %w", err)
	}

	// 4. Create user management client
	c.userMgmtClient = usermgmt.NewClient(config.AppNodeAddr)

	// 5. Create TLS configuration for App node
	appTLSConfig, err := c.createAppTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create App TLS config: %w", err)
	}

	// 5. Connect to user management system (with App TLS)
	if err := c.userMgmtClient.Connect(ctx, appTLSConfig); err != nil {
		return fmt.Errorf("failed to connect to user management system: %w", err)
	}

	fmt.Printf("Client initialized successfully, node ID: %d\n", config.NodeID)
	return nil
}

// createTEETLSConfig creates TLS configuration for TEE server
func (c *Client) createTEETLSConfig() (*tls.Config, error) {
	// Parse client certificate and key
	cert, err := tls.X509KeyPair(c.config.Cert, c.config.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Parse TEE server certificate
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(c.config.TargetCert) {
		return nil, fmt.Errorf("failed to parse TEE server certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil
}

// createAppTLSConfig creates TLS configuration for App node (user management system)
func (c *Client) createAppTLSConfig() (*tls.Config, error) {
	// Parse client certificate and key
	cert, err := tls.X509KeyPair(c.config.Cert, c.config.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Parse App node certificate
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(c.config.AppNodeCert) {
		return nil, fmt.Errorf("failed to parse App node certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil
}

// Close closes client connection
func (c *Client) Close() error {
	var errs []error

	if c.taskClient != nil {
		if err := c.taskClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.userMgmtClient != nil {
		if err := c.userMgmtClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}

// Sign executes signing operation
func (c *Client) Sign(message, publicKey []byte, protocol, curve uint32) ([]byte, error) {
	if c.taskClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	return c.taskClient.Sign(ctx, message, publicKey, protocol, curve)
}

// GetNodeID returns the node ID
func (c *Client) GetNodeID() uint32 {
	if c.config == nil {
		return 0
	}
	return c.config.NodeID
}

// SetTimeout sets the default timeout for all operations
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	if c.taskClient != nil {
		c.taskClient.SetTimeout(timeout)
	}
}

// SetTaskTimeout sets task timeout duration (deprecated, use SetTimeout)
func (c *Client) SetTaskTimeout(timeout time.Duration) {
	c.SetTimeout(timeout)
}

// GetPublicKeyByAppID retrieves public key from user management system by app ID
func (c *Client) GetPublicKeyByAppID(appID string) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.userMgmtClient.GetPublicKeyByAppID(ctx, appID)
}

// SignWithAppID signs a message using a public key from user management system by app ID
func (c *Client) SignWithAppID(message []byte, appID string) ([]byte, error) {
	// Get public key from user management system
	publicKeyStr, protocolStr, curveStr, err := c.GetPublicKeyByAppID(appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse protocol and curve strings to uint32
	protocol, err := parseProtocol(protocolStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protocol: %w", err)
	}

	curve, err := parseCurve(curveStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse curve: %w", err)
	}

	// Decode the public key from base64
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Sign the message
	return c.Sign(message, publicKey, protocol, curve)
}

// parseProtocol converts protocol string to uint32
func parseProtocol(protocol string) (uint32, error) {
	switch protocol {
	case "schnorr":
		return constants.ProtocolSchnorr, nil
	case "ecdsa":
		return constants.ProtocolECDSA, nil
	default:
		// Try to parse as number
		if num, err := strconv.ParseUint(protocol, 10, 32); err == nil {
			return uint32(num), nil
		}
		return constants.ProtocolSchnorr, nil // Default to schnorr
	}
}

// parseCurve converts curve string to uint32
func parseCurve(curve string) (uint32, error) {
	switch curve {
	case "ed25519":
		return constants.CurveED25519, nil
	case "secp256k1":
		return constants.CurveSECP256K1, nil
	case "secp256r1":
		return constants.CurveSECP256R1, nil
	default:
		// Try to parse as number
		if num, err := strconv.ParseUint(curve, 10, 32); err == nil {
			return uint32(num), nil
		}
		return constants.CurveED25519, nil // Default to ed25519
	}
}
