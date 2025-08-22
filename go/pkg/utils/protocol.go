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

// Package utils provides utility functions for TEE client operations
package utils

import (
	"strconv"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
)

// ParseProtocol converts protocol string to uint32
func ParseProtocol(protocol string) (uint32, error) {
	switch protocol {
	case "schnorr":
		return constants.ProtocolSchnorr, nil
	case "ecdsa":
		return constants.ProtocolECDSA, nil
	default:
		if num, err := strconv.ParseUint(protocol, 10, 32); err == nil {
			return uint32(num), nil
		}
		return constants.ProtocolSchnorr, nil // Default to schnorr
	}
}

// ParseCurve converts curve string to uint32
func ParseCurve(curve string) (uint32, error) {
	switch curve {
	case "ed25519":
		return constants.CurveED25519, nil
	case "secp256k1":
		return constants.CurveSECP256K1, nil
	case "secp256r1":
		return constants.CurveSECP256R1, nil
	default:
		if num, err := strconv.ParseUint(curve, 10, 32); err == nil {
			return uint32(num), nil
		}
		return constants.CurveED25519, nil // Default to ed25519
	}
}