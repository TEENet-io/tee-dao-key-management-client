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
import { VotingRequest, VotingResponse, DeploymentTarget, Constants } from './types';

// Voting service proto interface
interface VotingServiceClient {
  Voting(
    request: VotingRequest,
    callback: (error: grpc.ServiceError | null, response?: VotingResponse) => void
  ): grpc.ClientUnaryCall;
}

export class VotingClient {
  
  // Send voting request to a deployment target
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
        ) as VotingServiceClient;

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