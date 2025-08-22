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

package constants

import "time"

// Timeout constants for client operations
const (
	// DefaultClientTimeout is the default timeout for main client operations (DKG, Sign, Resharing)
	DefaultClientTimeout = 10 * time.Second

	// DefaultConfigTimeout is the default timeout for configuration operations
	DefaultConfigTimeout = 10 * time.Second

	// DefaultTaskTimeout is the default timeout for task client operations
	DefaultTaskTimeout = 10 * time.Second
)

// Protocol constants
const (
	ProtocolECDSA   uint32 = 1
	ProtocolSchnorr uint32 = 2
)

// Curve constants
const (
	CurveED25519   uint32 = 1
	CurveSECP256K1 uint32 = 2
	CurveSECP256R1 uint32 = 3
)

// gRPC retry configuration constants
const (
	// GRPCRetryPolicy is the complete retry policy configuration for gRPC
	GRPCRetryPolicy = `{
		"methodConfig": [{
			"name": [{"service": "UserTask"}],
			"retryPolicy": {
				"maxAttempts": 3,
				"initialBackoff": "0.1s",
				"maxBackoff": "1s",
				"backoffMultiplier": 2,
				"retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
			}
		}]
	}`
)
