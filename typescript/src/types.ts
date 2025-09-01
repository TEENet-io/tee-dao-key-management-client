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

export interface NodeConfig {
  nodeId: number;
  rpcAddress: string;
  cert: Buffer;
  key: Buffer;
  targetCert: Buffer;
  appNodeAddr: string;
  appNodeCert: Buffer;
}

export interface ClientOptions {
  configServerAddress: string;
  timeout?: number;
}

export interface SignRequest {
  from: number;
  msg: Uint8Array;
  publicKeyInfo: Uint8Array;
  protocol: number;
  curve: number;
}

export interface SignResponse {
  success: boolean;
  error?: string;
  signature?: Uint8Array;
}

export const Protocol = {
  ECDSA: 1,
  SCHNORR: 2,
} as const;

export const Curve = {
  ED25519: 1,
  SECP256K1: 2,
  SECP256R1: 3,
} as const;

export const NodeType = {
  INVALID_NODE: 0,
  TEE_NODE: 1,
  MESH_NODE: 2,
  APP_NODE: 3,
} as const;

// VoteDetail contains details of each vote
export interface VoteDetail {
  clientId: string;
  success: boolean;
  response: boolean;
  error?: string;
}

// VotingResult contains the result of a voting process
export interface VotingResult {
  totalTargets: number;
  successfulVotes: number;
  requiredVotes: number;
  votingComplete: boolean;
  finalResult: string;
  voteDetails: VoteDetail[];
  signature?: Uint8Array;
}

// VotingRequest for voting system
export interface VotingRequest {
  task_id: string;
  message: Uint8Array;
  required_votes: number;
  total_participants: number;
  app_id?: string;
  target_container_ip?: string;
}

// VotingResponse for voting system
export interface VotingResponse {
  success: boolean;
  task_id: string;
  error?: string;
}

// Voting handler function type
export type VotingHandler = (request: VotingRequest) => Promise<VotingResponse>;

// Deployment target information
export interface DeploymentTarget {
  appID: string;
  address: string;
  port: number;
  containerIP: string;
  deploymentClientAddress: string;  // gRPC deployment client address
  votingSignPath: string;           // HTTP API path for VotingSign requests
  httpBaseURL: string;              // HTTP base URL for API forwarding
  authHeaders?: { [key: string]: string }; // Optional authentication headers
}

export const Constants = {
  DEFAULT_CLIENT_TIMEOUT: 30000,
  DEFAULT_CONFIG_TIMEOUT: 10000,
  DEFAULT_TASK_TIMEOUT: 30000,
  DEFAULT_VOTING_PORT: 50053,
} as const;