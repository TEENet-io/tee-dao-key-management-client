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

import "math/big"

// ECDSASignature represents an ECDSA signature with r and s values
type ECDSASignature struct {
	R, S *big.Int
}


// IncomingVoteRequest for handling vote requests from other apps
type IncomingVoteRequest struct {
	Message           string   `json:"message" binding:"required"`           // Base64 encoded message
	SignerAppID       string   `json:"signer_app_id" binding:"required"`     // The app requesting the signature
	RequiredVotes     int      `json:"required_votes" binding:"required"`
	TargetAppIDs      []string `json:"target_app_ids,omitempty"`             // Target apps for further voting
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
	PublicKey string `json:"public_key,omitempty"`
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