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

// Package usermgmt provides gRPC client for user management system integration
package usermgmt

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
	"github.com/TEENet-io/tee-dao-key-management-client/go/proto/appid"
)

// Client handles gRPC communication with the user management system
type Client struct {
	conn       *grpc.ClientConn
	client     appid.AppIDServiceClient
	serverAddr string
}

// NewClient creates a new user management gRPC client
func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

// Connect establishes gRPC connection to user management service
func (c *Client) Connect(ctx context.Context, tlsConfig *tls.Config) error {
	// gRPC connection options with TLS and retry configuration
	if c.conn != nil {
		c.conn.Close()
	}

	// gRPC connection options with TLS and retry configuration
	creds := credentials.NewTLS(tlsConfig)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(constants.GRPCRetryPolicy),
	}

	conn, err := grpc.NewClient(c.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to user management service: %w", err)
	}

	c.conn = conn
	c.client = appid.NewAppIDServiceClient(conn)
	return nil
}

// GetPublicKeyByAppID retrieves public key by app ID via gRPC
func (c *Client) GetPublicKeyByAppID(ctx context.Context, appID string) (string, string, string, error) {
	if c.client == nil {
		return "", "", "", fmt.Errorf("client not connected")
	}

	req := &appid.GetPublicKeyByAppIDRequest{
		AppId: appID,
	}

	resp, err := c.client.GetPublicKeyByAppID(ctx, req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get public key: %w", err)
	}

	return resp.Publickey, resp.Protocol, resp.Curve, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
