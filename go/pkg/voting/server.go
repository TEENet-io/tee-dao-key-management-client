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

// Package voting provides gRPC voting service implementation
package voting

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/TEENet-io/tee-dao-key-management-client/go/proto/voting"
	"google.golang.org/grpc"
)

// Server wraps Client to implement VotingServiceServer with custom handler
type Server struct {
	pb.UnimplementedVotingServiceServer
	handler func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error)
}

// NewServer creates a new voting server with the provided handler
func NewServer(handler func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error)) *Server {
	return &Server{
		handler: handler,
	}
}

// Voting handles incoming voting requests (gRPC method implementation)
func (vs *Server) Voting(ctx context.Context, req *pb.VotingRequest) (*pb.VotingResponse, error) {
	log.Printf("🏛️  Received voting request: %s", req.TaskId)
	log.Printf("📄 Message: %s", string(req.Message))
	log.Printf("👥 Required votes: %d/%d", req.RequiredVotes, req.TotalParticipants)

	// Delegate to application-provided handler
	if vs.handler != nil {
		return vs.handler(ctx, req)
	}

	// Default fallback (should not be reached if handler is provided)
	log.Printf("⚠️  No voting handler provided, rejecting by default")
	return &pb.VotingResponse{
		Success: false,
		TaskId:  req.TaskId,
	}, nil
}

// StartVotingService starts the gRPC voting service to receive voting requests from other clients
func StartVotingService(votingHandler func(context.Context, *pb.VotingRequest) (*pb.VotingResponse, error), existingServer **grpc.Server) error {
	// Stop existing voting service if running
	if *existingServer != nil {
		(*existingServer).GracefulStop()
		*existingServer = nil
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on port 50051: %w", err)
	}

	*existingServer = grpc.NewServer()
	votingServer := NewServer(votingHandler)
	pb.RegisterVotingServiceServer(*existingServer, votingServer)

	log.Printf("🗳️  Voting service started on port 50051")

	go func() {
		if err := (*existingServer).Serve(lis); err != nil {
			log.Printf("❌ Voting service error: %v", err)
		}
	}()

	return nil
}