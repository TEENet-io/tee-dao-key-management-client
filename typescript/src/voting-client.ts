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

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import * as path from 'path';
import * as https from 'https';
import * as http from 'http';
import { VotingRequest, VotingResponse, DeploymentTarget, Constants } from './types';

// HTTP voting request payload (matches Go's HTTPVotingRequest)
interface HTTPVotingRequest {
  task_id: string;
  message: string;  // Base64 encoded message
  required_votes: number;
  total_participants: number;
  app_id: string;
}

// HTTP voting response payload (matches Go's HTTPVotingResponse)
interface HTTPVotingResponse {
  success: boolean;
  message?: string;
  error?: string;
}

export class VotingClient {
  
  // Send HTTP voting request directly to deployment-client (matches Go's SendHTTPVotingRequest)
  static async sendHTTPVotingRequest(
    target: DeploymentTarget,
    taskId: string,
    message: Uint8Array,
    requiredVotes: number,
    totalParticipants: number,
    timeout: number
  ): Promise<boolean> {
    // Check if VotingSign path is configured
    if (!target.votingSignPath) {
      throw new Error(`no VotingSign path configured for app ID ${target.appID}`);
    }

    // Prepare HTTP request payload
    const requestPayload: HTTPVotingRequest = {
      task_id: taskId,
      message: Buffer.from(message).toString('base64'),
      required_votes: requiredVotes,
      total_participants: totalParticipants,
      app_id: target.appID
    };

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
      const postData = JSON.stringify(requestPayload);
      
      const headers: { [key: string]: string } = {
        'Content-Type': 'application/json',
        'X-Task-ID': taskId,
        'X-App-ID': target.appID,
        'X-VotingSign-Request': 'true'
      };

      // Add authentication headers if provided
      if (target.authHeaders) {
        Object.assign(headers, target.authHeaders);
      }

      const options = {
        method: 'POST',
        headers,
        timeout: timeout
      };

      console.log(`DEBUG: Sending HTTP request to deployment-client: ${endpoint}`);
      
      const req = http.request(endpoint, options, (res) => {
        let data = '';
        
        res.on('data', (chunk) => {
          data += chunk;
        });

        res.on('end', () => {
          console.log(`DEBUG: HTTP Status: ${res.statusCode}`);
          console.log(`DEBUG: Response Body: ${data}`);

          try {
            const response: HTTPVotingResponse = JSON.parse(data);
            
            // Check HTTP status
            if (res.statusCode !== 200) {
              reject(new Error(`HTTP voting failed with status ${res.statusCode}: ${response.error}`));
              return;
            }

            if (!response.success) {
              resolve(false); // Voting rejected
            } else {
              resolve(true); // Voting approved
            }
          } catch (error) {
            reject(new Error(`Failed to parse HTTP response: ${error}`));
          }
        });
      });

      req.on('error', (error) => {
        reject(new Error(`HTTP voting request failed: ${error}`));
      });

      req.on('timeout', () => {
        req.destroy();
        reject(new Error('HTTP voting request timeout'));
      });

      // Send the request data
      req.write(postData);
      req.end();
    });
  }

  // Send smart voting request - automatically chooses between HTTP and gRPC, defaulting to HTTP
  // (matches Go's SendSmartVotingRequest)
  static async sendSmartVotingRequest(
    target: DeploymentTarget,
    taskId: string,
    message: Uint8Array,
    requiredVotes: number,
    totalParticipants: number,
    timeout: number
  ): Promise<boolean> {
    // Default to HTTP if VotingSign path is configured and HTTP base URL is available
    if (target.votingSignPath && target.httpBaseURL) {
      return this.sendHTTPVotingRequest(target, taskId, message, requiredVotes, totalParticipants, timeout);
    }

    // Fallback to gRPC if deployment client address is available
    if (target.deploymentClientAddress) {
      return this.sendVotingRequestToDeployment(target, target.appID, target.containerIP, taskId, message, requiredVotes, totalParticipants, timeout);
    }

    throw new Error(`no valid voting protocol configuration found for app ID ${target.appID}`);
  }

  // Legacy gRPC method - kept for fallback compatibility
  static async sendVotingRequestToDeployment(
    deploymentTarget: DeploymentTarget,
    appId: string,
    containerIP: string,
    taskId: string,
    message: Uint8Array,
    requiredVotes: number,
    totalTargets: number,
    timeout: number
  ): Promise<boolean> {
    return new Promise((resolve) => {
      try {
        // Load voting service proto
        const protoPath = path.resolve(__dirname, '../proto/voting/voting.proto');
        const packageDefinition = protoLoader.loadSync(protoPath, {
          keepCase: true,
          longs: String,
          enums: String,
          defaults: true,
          oneofs: true,
        });

        const votingProto = grpc.loadPackageDefinition(packageDefinition) as any;
        
        // Try to find VotingService - it might be at root level if no package is defined
        const VotingService = votingProto.VotingService || (votingProto.voting && votingProto.voting.VotingService);
        
        if (!VotingService) {
          throw new Error('VotingService not found in proto definition');
        }

        // Create connection (like Go version)
        const address = `${deploymentTarget.address}:${deploymentTarget.port}`;
        const connection = new grpc.Client(address, grpc.credentials.createInsecure());
        
        // Create client using the connection
        const client = new VotingService(
          address,
          grpc.credentials.createInsecure()
        ) as any;

        const request: VotingRequest = {
          task_id: taskId,
          message,
          required_votes: requiredVotes,
          total_participants: totalTargets,
          app_id: appId,
          target_container_ip: containerIP
        };
        

        let isResolved = false;
        let timeoutHandle: NodeJS.Timeout | null = null;

        const cleanup = () => {
          // Clear timeout if it exists
          if (timeoutHandle) {
            clearTimeout(timeoutHandle);
            timeoutHandle = null;
          }
          // Close the gRPC connection (like Go's defer conn.Close())
          connection.close();
        };

        const deadline = new Date();
        deadline.setMilliseconds(deadline.getMilliseconds() + timeout);

        client.Voting(request, (error: grpc.ServiceError | null, response?: VotingResponse) => {
          if (isResolved) return; // Prevent multiple resolutions
          isResolved = true;
          cleanup(); // Close connection and clear timeout

          if (error) {
            console.warn(`❌ Voting request failed for ${address}: ${error.message}`);
            resolve(false);
          } else if (response) {
            const approved = response.success;
            console.log(`${approved ? '✅' : '❌'} Vote ${approved ? 'approved' : 'rejected'} by ${address}`);
            resolve(approved);
          } else {
            console.warn(`❌ No response from ${address}`);
            resolve(false);
          }
        });

        // Handle timeout
        timeoutHandle = setTimeout(() => {
          if (isResolved) return; // Don't log timeout if already resolved
          isResolved = true;
          cleanup(); // Close connection and clear timeout
          console.warn(`⏰ Voting request timeout for ${address}`);
          resolve(false);
        }, timeout);

      } catch (error) {
        console.warn(`❌ Failed to create voting client for ${deploymentTarget.address}: ${error}`);
        resolve(false);
      }
    });
  }

  // Start voting service (simplified version - in production would use proper gRPC server)
  static startVotingService(
    votingHandler: (request: VotingRequest) => Promise<VotingResponse>,
    port: number = Constants.DEFAULT_VOTING_PORT
  ): Promise<grpc.Server> {
    return new Promise((resolve, reject) => {
      try {
        const protoPath = path.resolve(__dirname, '../proto/voting/voting.proto');
        const packageDefinition = protoLoader.loadSync(protoPath, {
          keepCase: true,
          longs: String,
          enums: String,
          defaults: true,
          oneofs: true,
        });

        const votingProto = grpc.loadPackageDefinition(packageDefinition) as any;
        
        const server = new grpc.Server();
        
        const VotingService = votingProto.VotingService || (votingProto.voting && votingProto.voting.VotingService);
        
        if (!VotingService) {
          throw new Error('VotingService not found in proto definition');
        }
        
        server.addService(VotingService.service, {
          Voting: async (
            call: grpc.ServerUnaryCall<VotingRequest, VotingResponse>,
            callback: grpc.sendUnaryData<VotingResponse>
          ) => {
            try {
              const response = await votingHandler(call.request);
              callback(null, response);
            } catch (error) {
              callback({
                code: grpc.status.INTERNAL,
                message: `Voting handler error: ${error}`
              } as grpc.ServiceError, null);
            }
          }
        });

        server.bindAsync(
          `0.0.0.0:${port}`,
          grpc.ServerCredentials.createInsecure(),
          (error, boundPort) => {
            if (error) {
              reject(error);
            } else {
              server.start();
              console.log(`🗳️  Voting service started on port ${boundPort}`);
              resolve(server);
            }
          }
        );

      } catch (error) {
        reject(error);
      }
    });
  }
}