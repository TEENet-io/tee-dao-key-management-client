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
	"fmt"
	"time"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/config"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/task"
)

// Client is a simplified key management client
type Client struct {
	configClient *config.Client
	taskClient   *task.Client
	config       *config.NodeConfig
	timeout      time.Duration
}

// NewClient creates a new client instance
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

	// 2. Create task client (with TLS)
	c.taskClient = task.NewClient(config)

	// 3. Connect to TEE server
	if err := c.taskClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to TEE server: %w", err)
	}

	fmt.Printf("Client initialized successfully, node ID: %d\n", config.NodeID)
	return nil
}

// Close closes client connection
func (c *Client) Close() error {
	if c.taskClient != nil {
		return c.taskClient.Close()
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
