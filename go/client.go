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
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/config"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/constants"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/task"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/usermgmt"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/utils"
	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/voting"
	pb "github.com/TEENet-io/tee-dao-key-management-client/go/proto/voting"
	"google.golang.org/grpc"
)

// VoteDetail contains details of each vote
type VoteDetail struct {
	ClientID string `json:"client_id"`
	Success  bool   `json:"success"`
	Response bool   `json:"response"`
	Error    string `json:"error,omitempty"`
}

// VotingResult contains the result of a voting process
type VotingResult struct {
	TaskID          string       `json:"task_id"`
	TotalTargets    int          `json:"total_targets"`
	SuccessfulVotes int          `json:"successful_votes"`
	RequiredVotes   int          `json:"required_votes"`
	VotingComplete  bool         `json:"voting_complete"`
	FinalResult     string       `json:"final_result"`
	VoteDetails     []VoteDetail `json:"vote_details"`
	Signature       []byte       `json:"signature,omitempty"`
}

// Client is a simplified key management client with voting capabilities
type Client struct {
	configClient   *config.Client
	taskClient     *task.Client
	userMgmtClient *usermgmt.Client
	nodeConfig     *config.NodeConfig
	timeout        time.Duration
	votingHandler  func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error)
	votingServer   *grpc.Server
}

// NewClient creates a new client instance
func NewClient(configServerAddr string) *Client {
	client := &Client{
		configClient: config.NewClient(configServerAddr),
		timeout:      constants.DefaultClientTimeout,
	}

	// Set default voting handler (auto-approve all votes)
	client.SetVotingHandler(client.createDefaultVotingHandler())

	return client
}

// createDefaultVotingHandler creates a default voting handler that auto-approves all voting requests
func (c *Client) createDefaultVotingHandler() func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error) {
	return func(ctx context.Context, req *pb.VotingRequest) (*pb.VotingResponse, error) {
		// Simulate processing delay
		time.Sleep(200 * time.Millisecond)

		// Auto-approve all voting requests by default
		log.Printf("✅ [DEFAULT] Auto-approving voting request for task: %s", req.TaskId)

		return &pb.VotingResponse{
			Success: true,
			TaskId:  req.TaskId,
		}, nil
	}
}

// SetVotingHandler allows users to set a custom voting handler and restarts the voting service
func (c *Client) SetVotingHandler(handler func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error)) {
	c.votingHandler = handler

	// If voting service is already running, restart it with the new handler
	if c.votingServer != nil {
		log.Printf("🔄 Restarting voting service with new handler...")
		if err := voting.StartVotingService(handler, &c.votingServer); err != nil {
			log.Printf("⚠️  Warning: Failed to restart voting service: %v", err)
		}
	}
}

// Init initializes client, fetches config and establishes TLS connection
// If votingHandler is nil, uses the default auto-approve handler
func (c *Client) Init(votingHandler func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 1. Fetch configuration
	nodeConfig, err := c.configClient.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	c.nodeConfig = nodeConfig

	// 2. Create task client
	c.taskClient = task.NewClient(nodeConfig)

	// 3. Create TLS configuration for TEE server
	teeTLSConfig, err := utils.CreateTLSConfig(nodeConfig.Cert, nodeConfig.Key, nodeConfig.TargetCert)
	if err != nil {
		return fmt.Errorf("failed to create TEE TLS config: %w", err)
	}

	// 4. Connect to TEE server
	if err := c.taskClient.Connect(ctx, teeTLSConfig); err != nil {
		return fmt.Errorf("failed to connect to TEE server: %w", err)
	}

	// 5. Create user management client
	c.userMgmtClient = usermgmt.NewClient(nodeConfig.AppNodeAddr)

	// 6. Create TLS configuration for App node
	appTLSConfig, err := utils.CreateTLSConfig(nodeConfig.Cert, nodeConfig.Key, nodeConfig.AppNodeCert)
	if err != nil {
		return fmt.Errorf("failed to create App TLS config: %w", err)
	}

	// 7. Connect to user management system
	if err := c.userMgmtClient.Connect(ctx, appTLSConfig); err != nil {
		return fmt.Errorf("failed to connect to user management system: %w", err)
	}

	// 8. Set voting handler and auto-start voting service
	if votingHandler != nil {
		c.votingHandler = votingHandler
		log.Printf("🗳️  Using custom voting handler provided in Init()")
	} else {
		log.Printf("🗳️  Using default auto-approve voting handler")
	}

	if err := voting.StartVotingService(c.votingHandler, &c.votingServer); err != nil {
		log.Printf("⚠️  Warning: Failed to start voting service: %v", err)
		// Don't fail initialization if voting service fails to start
	} else {
		log.Printf("🗳️  Voting service auto-started during initialization")
	}

	log.Printf("✅ Client initialized successfully, node ID: %d", nodeConfig.NodeID)
	return nil
}

// SignWithAppID signs a message using a public key from user management system by app ID
func (c *Client) SignWithAppID(message []byte, appID string) ([]byte, error) {
	if c.taskClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Get public key from user management system
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	publicKeyStr, protocolStr, curveStr, err := c.userMgmtClient.GetPublicKeyByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse protocol and curve strings to uint32
	protocol, err := utils.ParseProtocol(protocolStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse protocol: %w", err)
	}

	curve, err := utils.ParseCurve(curveStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse curve: %w", err)
	}

	// Decode the public key from base64
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Sign the message
	ctx2, cancel2 := context.WithTimeout(context.Background(), c.timeout)
	defer cancel2()

	return c.taskClient.Sign(ctx2, message, publicKey, protocol, curve)
}

// GetPublicKeyByAppID gets public key information for a specific app ID
func (c *Client) GetPublicKeyByAppID(appID string) (publicKey, protocol, curve string, err error) {
	if c.userMgmtClient == nil {
		return "", "", "", fmt.Errorf("client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.userMgmtClient.GetPublicKeyByAppID(ctx, appID)
}

// VotingSign performs a voting process among specified app IDs and returns detailed results with signature if approved
func (c *Client) VotingSign(message []byte, signerAppID string, targetAppIDs []string, requiredVotes int) (*VotingResult, error) {
	if len(targetAppIDs) == 0 {
		return nil, fmt.Errorf("no target app IDs provided")
	}

	if requiredVotes <= 0 || requiredVotes > len(targetAppIDs) {
		return nil, fmt.Errorf("invalid required votes: %d (should be 1-%d)", requiredVotes, len(targetAppIDs))
	}

	taskID := fmt.Sprintf("vote_%s_%d", signerAppID, time.Now().UnixNano())
	log.Printf("🗳️  Starting voting process: %s", taskID)
	log.Printf("👥 Targets: %v, required votes: %d/%d", targetAppIDs, requiredVotes, len(targetAppIDs))

	// Batch get deployment targets for all target app IDs
	deploymentTargets, err := c.userMgmtClient.GetDeploymentTargetsForAppIDs(targetAppIDs, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment targets: %w", err)
	}

	// Send voting requests to all target app IDs concurrently
	type voteResult struct {
		appID    string
		approved bool
		err      error
	}

	resultChan := make(chan voteResult, len(targetAppIDs))
	activeRequests := 0

	// Start concurrent voting requests
	for _, targetAppID := range targetAppIDs {
		target, exists := deploymentTargets[targetAppID]
		if !exists {
			log.Printf("❌ No deployment target found for %s, skipping", targetAppID)
			continue
		}

		activeRequests++
		go func(appID string, deployTarget *usermgmt.DeploymentTarget) {
			approved, err := voting.SendVotingRequestToDeployment(deployTarget, taskID, message, requiredVotes, len(targetAppIDs), c.timeout)
			resultChan <- voteResult{appID: appID, approved: approved, err: err}
		}(targetAppID, target)
	}

	// Collect results
	voteDetails := make([]VoteDetail, 0, len(targetAppIDs))
	approvalCount := 0

	for i := 0; i < activeRequests; i++ {
		result := <-resultChan

		voteDetail := VoteDetail{
			ClientID: result.appID,
			Success:  result.err == nil,
			Response: result.approved,
		}

		if result.err != nil {
			voteDetail.Error = result.err.Error()
			log.Printf("❌ Failed to get vote from %s: %v", result.appID, result.err)
		} else if result.approved {
			approvalCount++
			log.Printf("✅ Vote approved by %s (%d/%d)", result.appID, approvalCount, requiredVotes)
		} else {
			log.Printf("❌ Vote rejected by %s", result.appID)
		}

		voteDetails = append(voteDetails, voteDetail)
	}

	// Create voting result
	votingResult := &VotingResult{
		TaskID:          taskID,
		TotalTargets:    len(targetAppIDs),
		SuccessfulVotes: approvalCount,
		RequiredVotes:   requiredVotes,
		VotingComplete:  approvalCount >= requiredVotes,
		VoteDetails:     voteDetails,
	}

	// Check if voting passed
	if approvalCount < requiredVotes {
		votingResult.FinalResult = "REJECTED"
		log.Printf("❌ Voting failed: only %d/%d approvals received", approvalCount, requiredVotes)
		return votingResult, fmt.Errorf("voting failed: only %d/%d approvals received", approvalCount, requiredVotes)
	}

	// Generate signature
	log.Printf("🔐 Generating signature for approved message (%d/%d votes received)", approvalCount, requiredVotes)
	signature, err := c.SignWithAppID(message, signerAppID)
	if err != nil {
		votingResult.FinalResult = "SIGNATURE_FAILED"
		return votingResult, fmt.Errorf("failed to generate signature: %w", err)
	}

	votingResult.FinalResult = "APPROVED"
	votingResult.Signature = signature

	log.Printf("✅ Voting and signing completed successfully")
	return votingResult, nil
}

// Close closes client connections
func (c *Client) Close() error {
	var errs []error

	// Stop voting service gracefully
	if c.votingServer != nil {
		log.Printf("🛑 Stopping voting service...")
		c.votingServer.GracefulStop()
		c.votingServer = nil
	}

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
