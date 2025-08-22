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

// VotingResultSummary contains the summary of voting results
type VotingResultSummary struct {
	TotalResponses  int          `json:"total_responses"`
	SuccessfulVotes int          `json:"successful_votes"`
	RequiredVotes   int          `json:"required_votes"`
	VotingComplete  bool         `json:"voting_complete"`
	FinalResult     string       `json:"final_result"`
	VoteDetails     []VoteDetail `json:"vote_details"`
}

// VoteDetail contains details of each vote
type VoteDetail struct {
	ClientID string `json:"client_id"`
	Success  bool   `json:"success"`
	Response bool   `json:"response"`
	Error    string `json:"error,omitempty"`
}

// VotingRequest for handling HTTP requests
type VotingRequest struct {
	Description        string   `json:"description"`
	TargetAppIDs       []string `json:"target_app_ids"`
	RequiredVotes      int      `json:"required_votes"`
	TotalParticipants  int      `json:"total_participants"`
}

// VotingResponse for handling HTTP responses
type VotingResponse struct {
	Success       bool                   `json:"success"`
	TaskID        string                 `json:"task_id"`
	Message       string                 `json:"message"`
	VotingResults *VotingResultSummary   `json:"voting_results,omitempty"`
	Signature     string                 `json:"signature,omitempty"`
	Timestamp     string                 `json:"timestamp,omitempty"`
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