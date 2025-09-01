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

import { NodeConfig, ClientOptions, Constants, Protocol, Curve, VotingResult, VoteDetail, VotingHandler, VotingRequest, VotingResponse, DeploymentTarget } from './types';
import { ConfigClient } from './config-client';
import { TaskClient } from './task-client';
import { AppIDClient } from './appid-client';
import { VotingClient } from './voting-client';
import * as tls from 'tls';
import * as grpc from '@grpc/grpc-js';
import { IncomingMessage } from 'http';

export class Client {
  private configClient: ConfigClient;
  private taskClient: TaskClient | null = null;
  private appIDClient: AppIDClient | null = null;
  private nodeConfig: NodeConfig | null = null;
  private timeout: number;
  private votingHandler: VotingHandler;
  private votingServer: grpc.Server | null = null;

  constructor(configServerAddress: string, options?: Partial<ClientOptions>) {
    this.configClient = new ConfigClient(configServerAddress);
    this.timeout = options?.timeout || Constants.DEFAULT_CLIENT_TIMEOUT;
    
    // Set default voting handler (auto-approve all votes)
    this.votingHandler = this.createDefaultVotingHandler();
  }

  // Create default voting handler that auto-approves all voting requests
  private createDefaultVotingHandler(): VotingHandler {
    return async (request: VotingRequest): Promise<VotingResponse> => {
      // Simulate processing delay
      await new Promise(resolve => setTimeout(resolve, 200));

      // Auto-approve all voting requests by default
      console.log(`✅ [DEFAULT] Auto-approving voting request for task: ${request.task_id}`);

      return {
        success: true,
        task_id: request.task_id,
      };
    };
  }

  // Set custom voting handler and restart voting service if running
  setVotingHandler(handler: VotingHandler): void {
    this.votingHandler = handler;

    // If voting service is already running, restart it with the new handler
    if (this.votingServer) {
      console.log('🔄 Restarting voting service with new handler...');
      this.restartVotingService();
    }
  }

  async init(votingHandler?: VotingHandler): Promise<void> {
    // 1. Fetch configuration
    const nodeConfig = await this.configClient.getConfig(this.timeout);
    this.nodeConfig = nodeConfig;

    // 2. Create task client
    this.taskClient = new TaskClient(nodeConfig);
    
    // 3. Create TLS configuration for TEE server
    const teeTLSConfig = this.createTEETLSConfig();
    
    // 4. Connect to TEE server
    await this.taskClient.connect(this.timeout);

    // 5. Create AppID client
    this.appIDClient = new AppIDClient(nodeConfig.appNodeAddr);
    
    // 6. Create TLS configuration for App node
    const appTLSConfig = this.createAppTLSConfig();
    
    // 7. Connect to user management system
    await this.appIDClient.connect(appTLSConfig);

    // 8. Set voting handler and auto-start voting service
    if (votingHandler) {
      this.votingHandler = votingHandler;
      console.log('🗳️  Using custom voting handler provided in init()');
    } else {
      console.log('🗳️  Using default auto-approve voting handler');
    }

    try {
      this.votingServer = await VotingClient.startVotingService(this.votingHandler);
      console.log('🗳️  Voting service auto-started during initialization');
    } catch (error) {
      console.warn(`⚠️  Warning: Failed to start voting service: ${error}`);
      // Don't fail initialization if voting service fails to start
    }

    console.log(`✅ Client initialized successfully, node ID: ${nodeConfig.nodeId}`);
  }

  private async restartVotingService(): Promise<void> {
    if (this.votingServer) {
      // Quick shutdown without waiting
      setImmediate(() => this.votingServer?.forceShutdown());
      this.votingServer = null;
    }

    try {
      this.votingServer = await VotingClient.startVotingService(this.votingHandler);
    } catch (error) {
      console.warn(`⚠️  Warning: Failed to restart voting service: ${error}`);
    }
  }

  async close(): Promise<void> {
    console.log('🛑 Stopping voting service...');
    
    // Gracefully stop voting service
    if (this.votingServer) {
      return new Promise<void>((resolve) => {
        console.log('🔄 Attempting graceful shutdown...');
        this.votingServer!.tryShutdown(() => {
          console.log('✅ Voting service stopped gracefully');
          this.votingServer = null;
          
          // Close other clients
          console.log('🔄 Closing task client...');
          if (this.taskClient) {
            this.taskClient.close();
            this.taskClient = null;
          }
          console.log('🔄 Closing appID client...');
          if (this.appIDClient) {
            this.appIDClient.close();
            this.appIDClient = null;
          }
          console.log('✅ All clients closed');
          
          resolve();
        });
      });
    }

    // If no voting server, just close other clients
    if (this.taskClient) {
      await this.taskClient.close();
      this.taskClient = null;
    }
    if (this.appIDClient) {
      this.appIDClient.close();
      this.appIDClient = null;
    }
  }


  getNodeId(): number {
    if (!this.nodeConfig) {
      return 0;
    }
    return this.nodeConfig.nodeId;
  }

  setTimeout(timeout: number): void {
    this.timeout = timeout;
    if (this.taskClient) {
      this.taskClient.setTimeout(timeout);
    }
  }

  setTaskTimeout(timeout: number): void {
    this.setTimeout(timeout);
  }

  // Create TLS configuration for TEE server
  private createTEETLSConfig(): tls.SecureContextOptions {
    if (!this.nodeConfig) {
      throw new Error('config not loaded');
    }
    return {
      cert: this.nodeConfig.cert,
      key: this.nodeConfig.key,
      ca: this.nodeConfig.targetCert,
    };
  }

  // Create TLS configuration for App node (user management system)
  private createAppTLSConfig(): tls.SecureContextOptions {
    if (!this.nodeConfig) {
      throw new Error('config not loaded');
    }
    return {
      cert: this.nodeConfig.cert,
      key: this.nodeConfig.key,
      ca: this.nodeConfig.appNodeCert,
    };
  }

  // Get public key by app ID from user management system
  async getPublicKeyByAppID(appId: string): Promise<{publickey: string, protocol: string, curve: string}> {
    if (!this.appIDClient) {
      throw new Error('AppID client not initialized');
    }
    return this.appIDClient.getPublicKeyByAppID(appId);
  }


  // Parse protocol string to number
  private parseProtocol(protocol: string): number {
    switch (protocol) {
      case 'schnorr':
        return Protocol.SCHNORR;
      case 'ecdsa':
        return Protocol.ECDSA;
      default:
        const num = parseInt(protocol, 10);
        return isNaN(num) ? Protocol.SCHNORR : num; // Default to schnorr
    }
  }

  // Parse curve string to number
  private parseCurve(curve: string): number {
    switch (curve) {
      case 'ed25519':
        return Curve.ED25519;
      case 'secp256k1':
        return Curve.SECP256K1;
      case 'secp256r1':
        return Curve.SECP256R1;
      default:
        const num = parseInt(curve, 10);
        return isNaN(num) ? Curve.ED25519 : num; // Default to ed25519
    }
  }

  // Sign with AppID (combines getPublicKeyByAppID and taskClient.sign)
  async signWithAppID(message: Uint8Array, appId: string): Promise<Uint8Array> {
    if (!this.taskClient) {
      throw new Error('client not initialized');
    }

    // Get public key from user management system
    const { publickey, protocol, curve } = await this.getPublicKeyByAppID(appId);

    // Parse protocol and curve
    const protocolNum = this.parseProtocol(protocol);
    const curveNum = this.parseCurve(curve);

    // Decode public key from base64
    const publicKeyBuffer = Buffer.from(publickey, 'base64');

    // Sign the message directly through taskClient
    return this.taskClient.sign(message, new Uint8Array(publicKeyBuffer), protocolNum, curveNum, this.timeout);
  }

  // VotingSign performs a voting process among specified app IDs using HTTP requests and returns detailed results with signature if approved
  // Method signature matches Go version: VotingSign(req *http.Request, message []byte, signerAppID string, targetAppIDs []string, requiredVotes int, localApproval bool)
  async votingSign(
    req: IncomingMessage | null,
    message: Uint8Array,
    signerAppId: string,
    targetAppIds: string[],
    requiredVotes: number,
    localApproval: boolean
  ): Promise<VotingResult>;

  async votingSign(
    req: IncomingMessage | null,
    message: Uint8Array,
    signerAppId: string,
    targetAppIds: string[],
    requiredVotes: number,
    localApproval: boolean
  ): Promise<VotingResult> {
    let finalRequestBody: Uint8Array | undefined;
    let headers: { [key: string]: string } | undefined;

    // Extract request body and headers from HTTP request if available
    if (req) {
      if ((req as any).body) {
        const bodyStr = typeof (req as any).body === 'string' ? (req as any).body : JSON.stringify((req as any).body);
        finalRequestBody = new TextEncoder().encode(bodyStr);
      }
      headers = this.extractHeadersFromRequest(req);
    }
    // Parse isForwarded from the request data
    let isForwarded = false;
    if (finalRequestBody) {
      try {
        const requestMap = JSON.parse(new TextDecoder().decode(finalRequestBody));
        isForwarded = requestMap.is_forwarded || false;
      } catch (error) {
        // Ignore JSON parse errors
      }
    }

    // If this is a forwarded request, just return the local decision without further forwarding
    if (isForwarded) {
      console.log(`🔄 Forwarded request - returning local decision: ${localApproval} for app ${signerAppId}`);

      const result: VotingResult = {
        totalTargets: 1,
        successfulVotes: localApproval ? 1 : 0,
        requiredVotes,
        votingComplete: localApproval,
        finalResult: localApproval ? 'APPROVED' : 'REJECTED',
        voteDetails: [{ clientId: signerAppId, success: true, response: localApproval }]
      };

      return result;
    }

    if (targetAppIds.length === 0) {
      throw new Error('no target app IDs provided');
    }

    if (requiredVotes <= 0 || requiredVotes > targetAppIds.length) {
      throw new Error(`invalid required votes: ${requiredVotes} (should be 1-${targetAppIds.length})`);
    }

    if (!this.appIDClient) {
      throw new Error('AppID client not initialized');
    }

    console.log(`🗳️  Starting HTTP voting process for ${signerAppId}`);
    console.log(`👥 Targets: ${JSON.stringify(targetAppIds)}, required votes: ${requiredVotes}/${targetAppIds.length}`);

    // Initialize vote details and approval count
    const voteDetails: VoteDetail[] = [];
    let approvalCount = 0;
    
    // Add local vote only if signerAppId is in targetAppIds
    const signerInTargets = targetAppIds.includes(signerAppId);
    if (signerInTargets) {
      voteDetails.push({ clientId: signerAppId, success: true, response: localApproval });
      if (localApproval) {
        approvalCount = 1;
      }
    }

    // Batch get deployment targets for remote app IDs (excluding self)
    const remoteTargetAppIds = targetAppIds.filter(targetAppId => targetAppId !== signerAppId);

    // If there are remote targets, send voting requests
    if (remoteTargetAppIds.length > 0) {
      console.log(`🔍 Getting deployment targets for remote apps: ${JSON.stringify(remoteTargetAppIds)}`);
      
      try {
        const deploymentTargets = await this.appIDClient.getDeploymentTargetsForAppIDs(remoteTargetAppIds, this.timeout);
        console.log(`✅ Found ${Object.keys(deploymentTargets).length} deployment targets: ${JSON.stringify(Object.keys(deploymentTargets))}`);

        // Send HTTP voting requests to remote targets concurrently
        const votePromises = remoteTargetAppIds.map(async (targetAppId): Promise<VoteDetail> => {
          const target = deploymentTargets[targetAppId];
          if (!target) {
            console.log(`❌ No deployment target found for ${targetAppId}, skipping`);
            return {
              clientId: targetAppId,
              success: false,
              response: false,
              error: 'No deployment target found'
            };
          }

          try {
            // Modify request body to mark as forwarded
            let modifiedRequestData = finalRequestBody;
            if (finalRequestBody) {
              const requestMap = JSON.parse(new TextDecoder().decode(finalRequestBody));
              requestMap.is_forwarded = true;
              modifiedRequestData = new TextEncoder().encode(JSON.stringify(requestMap));
            }

            const approved = await this.sendHTTPVoteRequest(target, modifiedRequestData || new Uint8Array(), headers);

            return {
              clientId: targetAppId,
              success: true,
              response: approved,
            };
          } catch (error) {
            console.log(`❌ Failed to get vote from ${targetAppId}: ${error}`);
            return {
              clientId: targetAppId,
              success: false,
              response: false,
              error: error instanceof Error ? error.message : String(error)
            };
          }
        });

        // Collect remote voting results
        const remoteVoteDetails = await Promise.all(votePromises);
        voteDetails.push(...remoteVoteDetails);

        // Count additional approvals from remote votes
        for (const detail of remoteVoteDetails) {
          if (detail.success && detail.response) {
            approvalCount++;
            console.log(`✅ Vote approved by ${detail.clientId} (${approvalCount}/${requiredVotes})`);
          } else if (detail.success && !detail.response) {
            console.log(`❌ Vote rejected by ${detail.clientId}`);
          } else {
            console.log(`❌ Failed to get vote from ${detail.clientId}: ${detail.error}`);
          }
        }
      } catch (error) {
        console.log(`❌ Failed to get deployment targets: ${error}`);
        throw new Error(`failed to get deployment targets: ${error}`);
      }
    }

    // Create final voting result
    const votingResult: VotingResult = {
      totalTargets: targetAppIds.length,
      successfulVotes: approvalCount,
      requiredVotes,
      votingComplete: true,
      finalResult: '',
      voteDetails
    };

    // Check if voting passed
    if (approvalCount < requiredVotes) {
      votingResult.finalResult = 'REJECTED';
      console.log(`❌ Voting failed: only ${approvalCount}/${requiredVotes} approvals received`);
      return votingResult; // Don't throw error, just return result
    }

    // Generate signature
    console.log(`🔐 Generating signature for approved message (${approvalCount}/${requiredVotes} votes received)`);
    try {
      const signature = await this.signWithAppID(message, signerAppId);
      votingResult.finalResult = 'APPROVED';
      votingResult.signature = signature;
    } catch (error) {
      votingResult.finalResult = 'SIGNATURE_FAILED';
      throw new Error(`failed to generate signature: ${error}`);
    }

    console.log('✅ Voting and signing completed successfully');
    return votingResult;
  }

  // Helper method to send HTTP vote request to a target app
  private async sendHTTPVoteRequest(target: DeploymentTarget, requestData: Uint8Array, additionalHeaders?: { [key: string]: string }): Promise<boolean> {
    // Build endpoint URL - send to deployment-client on port 8090 for HTTP forwarding
    // Format: http://deployment-host:8090/proxy/{app_id}{voting_sign_path}
    const deploymentHost = target.deploymentClientAddress.split(':')[0]; // Extract host from gRPC address
    let votingSignPath = target.votingSignPath;
    if (!votingSignPath.startsWith('/')) {
      votingSignPath = '/' + votingSignPath;
    }

    const proxyPath = `/proxy/${target.appID}${votingSignPath}`;
    const endpoint = `http://${deploymentHost}:8090${proxyPath}`;

    return new Promise((resolve, reject) => {
      const http = require('http');
      
      const headers: { [key: string]: string } = {
        'Content-Type': 'application/json'
      };

      // Add authentication headers if provided (for proxy forwarding)
      if (target.authHeaders) {
        Object.assign(headers, target.authHeaders);
      }

      // Add additional headers extracted from HTTP request
      if (additionalHeaders) {
        Object.assign(headers, additionalHeaders);
      }

      const options = {
        method: 'POST',
        headers,
        timeout: this.timeout
      };

      console.log(`📤 Sending vote request to ${target.appID} via deployment-client: ${endpoint}`);
      
      const req = http.request(endpoint, options, (res: any) => {
        let data = '';
        
        res.on('data', (chunk: any) => {
          data += chunk;
        });

        res.on('end', () => {
          // Check HTTP status
          if (res.statusCode !== 200) {
            reject(new Error(`HTTP vote request failed with status ${res.statusCode}: ${data}`));
            return;
          }

          try {
            // Parse response - only check for approved field
            const response = JSON.parse(data);
            const approved = response.approved || false;
            
            console.log(`📥 Received vote response from ${target.appID}: approved=${approved}`);
            resolve(approved);
          } catch (error) {
            reject(new Error(`failed to parse vote response: ${error}`));
          }
        });
      });

      req.on('error', (error: any) => {
        reject(new Error(`HTTP vote request failed: ${error}`));
      });

      req.on('timeout', () => {
        req.destroy();
        reject(new Error('HTTP vote request timeout'));
      });

      // Send the request data
      req.write(requestData);
      req.end();
    });
  }

  // Helper method to extract headers from HTTP request
  private extractHeadersFromRequest(req: IncomingMessage): { [key: string]: string } {
    const headers: { [key: string]: string } = {};
    
    for (const [name, value] of Object.entries(req.headers)) {
      if (value) {
        // Take first value if it's an array, otherwise use the string value
        headers[name] = Array.isArray(value) ? value[0] : value;
      }
    }
    
    return headers;
  }
}