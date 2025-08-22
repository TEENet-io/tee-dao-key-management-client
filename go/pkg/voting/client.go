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

// Package voting provides voting service client and server implementations
package voting

import (
	"context"
	"fmt"
	"time"

	"github.com/TEENet-io/tee-dao-key-management-client/go/pkg/usermgmt"
	pb "github.com/TEENet-io/tee-dao-key-management-client/go/proto/voting"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SendVotingRequestToDeployment sends a voting request to deployment-client which forwards to container
func SendVotingRequestToDeployment(target *usermgmt.DeploymentTarget, taskID string, message []byte, requiredVotes, totalParticipants int, timeout time.Duration) (bool, error) {
	// Connect to deployment-client's gRPC service
	conn, err := grpc.NewClient(target.DeploymentClientAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, fmt.Errorf("failed to connect to deployment-client %s: %w", target.DeploymentClientAddress, err)
	}
	defer conn.Close()

	grpcClient := pb.NewVotingServiceClient(conn)

	// Send voting request with container IP for deployment-client to forward
	request := &pb.VotingRequest{
		TaskId:            taskID,
		Message:           message,
		RequiredVotes:     uint32(requiredVotes),
		TotalParticipants: uint32(totalParticipants),
		AppId:             target.AppID,
		TargetContainerIp: target.ContainerIP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := grpcClient.Voting(ctx, request)
	if err != nil {
		return false, fmt.Errorf("voting request failed: %w", err)
	}

	if !response.Success {
		return false, nil // Voting rejected
	}

	return true, nil // Voting approved
}
