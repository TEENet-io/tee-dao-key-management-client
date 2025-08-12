package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "tee-dao-mock-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MockConfigServer implements the CLIRPCService
type MockConfigServer struct {
	pb.UnimplementedCLIRPCServiceServer
	clientCert []byte
	clientKey  []byte
	teeCert    []byte
	appCert    []byte
}

// NewMockConfigServer creates a new mock config server
func NewMockConfigServer() *MockConfigServer {
	// Load certificates
	clientCert, err := os.ReadFile("certs/client.crt")
	if err != nil {
		log.Fatalf("Failed to load client certificate: %v", err)
	}

	clientKey, err := os.ReadFile("certs/client.key")
	if err != nil {
		log.Fatalf("Failed to load client key: %v", err)
	}

	serverCert, err := os.ReadFile("certs/dao-server.crt")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	appCert, err := os.ReadFile("certs/app-node.crt")
	if err != nil {
		log.Fatalf("Failed to load app certificate: %v", err)
	}

	return &MockConfigServer{
		clientCert: clientCert,
		clientKey:  clientKey,
		teeCert:    serverCert,
		appCert:    appCert,
	}
}

// GetNodeInfo returns node information for the client
func (s *MockConfigServer) GetNodeInfo(ctx context.Context, req *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	log.Printf("Config server: GetNodeInfo called")

	return &pb.GetNodeInfoResponse{
		NodeId:     1001, // Mock client node ID
		RpcAddress: "localhost:50051",
		Cert:       s.clientCert,
		Key:        s.clientKey,
	}, nil
}

// GetPeerNode returns peer node information
func (s *MockConfigServer) GetPeerNode(ctx context.Context, req *pb.GetPeerNodeRequest) (*pb.GetPeerNodeResponse, error) {
	log.Printf("Config server: GetPeerNode called with type: %s", req.NodeType)

	peers := []*pb.Peer{
		{
			Id:         2001,
			RpcAddress: "localhost:50051", // Mock DAO server address
			Cert:       s.teeCert,
			Type:       1, // TEE node type
		},
		{
			Id:         3001,
			RpcAddress: "localhost:50053", // Mock App node address 
			Cert:       s.appCert,
			Type:       3, // App node type
		},
	}

	return &pb.GetPeerNodeResponse{
		Peers: peers,
	}, nil
}

func main() {
	port := ":50052"
	if p := os.Getenv("CONFIG_SERVER_PORT"); p != "" {
		port = ":" + p
	}

	log.Printf("Starting Mock Config Server on port %s", port)

	// Create listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server without TLS (as per original design)
	s := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	// Register service
	configServer := NewMockConfigServer()
	pb.RegisterCLIRPCServiceServer(s, configServer)

	fmt.Printf("Mock Config Server listening on %s (no TLS)\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}