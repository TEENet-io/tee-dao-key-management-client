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
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// CreateTLSConfig creates TLS configuration for TEE server
func CreateTLSConfig(cert, key, targetCert []byte) (*tls.Config, error) {
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(targetCert) {
		return nil, fmt.Errorf("failed to parse TEE server certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      caPool,
	}, nil
}
