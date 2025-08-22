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
import * as tls from 'tls';
import { DeploymentTarget } from './types';

export interface GetPublicKeyByAppIDRequest {
  app_id: string;
}

export interface GetPublicKeyByAppIDResponse {
  publickey: string;
  protocol: string;
  curve: string;
}

export interface GetDeploymentAddressesRequest {
  app_ids: string[];
}

export interface DeploymentInfo {
  app_id: string;
  project_name: string;
  deployment_host: string;
  container_ip: string;
  service_port: number;
  deployment_client_address: string;
  deployed_at: number;
  deployment_type: string;
}

export interface GetDeploymentAddressesResponse {
  deployments: { [appId: string]: DeploymentInfo };
  not_found: string[];
}

interface AppIDServiceClient {
  GetPublicKeyByAppID(
    request: GetPublicKeyByAppIDRequest,
    callback: (error: grpc.ServiceError | null, response?: GetPublicKeyByAppIDResponse) => void
  ): grpc.ClientUnaryCall;
  
  GetDeploymentAddresses(
    request: GetDeploymentAddressesRequest,
    callback: (error: grpc.ServiceError | null, response?: GetDeploymentAddressesResponse) => void
  ): grpc.ClientUnaryCall;
}

export class AppIDClient {
  private serverAddr: string;
  private client: AppIDServiceClient | null = null;
  private grpcConnection: grpc.Client | null = null;

  constructor(serverAddr: string) {
    this.serverAddr = serverAddr;
  }

  async connect(tlsConfig: tls.SecureContextOptions): Promise<void> {
    try {
      // Close existing connection if any
      if (this.grpcConnection) {
        this.grpcConnection.close();
      }

      const protoPath = path.resolve(__dirname, '../proto/appid/appid_service.proto');
      const packageDefinition = protoLoader.loadSync(protoPath, {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
      });

      const appidProto = grpc.loadPackageDefinition(packageDefinition) as any;
      const AppIDServiceClient = appidProto.appid.AppIDService;

      // Create TLS credentials
      const credentials = grpc.credentials.createSsl(
        tlsConfig.ca as Buffer,
        tlsConfig.key as Buffer,
        tlsConfig.cert as Buffer
      );

      this.grpcConnection = new AppIDServiceClient(
        this.serverAddr,
        credentials
      ) as grpc.Client;
      
      this.client = this.grpcConnection as unknown as AppIDServiceClient;
    } catch (error) {
      throw new Error(`failed to connect to AppID service: ${error}`);
    }
  }

  async getPublicKeyByAppID(appId: string): Promise<{publickey: string, protocol: string, curve: string}> {
    if (!this.client) {
      throw new Error('client not connected');
    }

    return new Promise((resolve, reject) => {
      const request: GetPublicKeyByAppIDRequest = { app_id: appId };
      
      this.client!.GetPublicKeyByAppID(request, (error: grpc.ServiceError | null, response?: GetPublicKeyByAppIDResponse) => {
        if (error) {
          reject(new Error(`failed to get public key: ${error.message}`));
        } else if (response) {
          resolve({
            publickey: response.publickey,
            protocol: response.protocol,
            curve: response.curve
          });
        } else {
          reject(new Error('no response received'));
        }
      });
    });
  }

  async getDeploymentAddresses(appIds: string[], timeout: number): Promise<{ deployments: { [appId: string]: DeploymentInfo }, notFound: string[] }> {
    if (!this.client) {
      throw new Error('client not connected');
    }

    return new Promise((resolve, reject) => {
      const request: GetDeploymentAddressesRequest = { app_ids: appIds };
      
      this.client!.GetDeploymentAddresses(request, (error: grpc.ServiceError | null, response?: GetDeploymentAddressesResponse) => {
        if (error) {
          reject(new Error(`failed to get deployment addresses: ${error.message}`));
        } else if (response) {
          resolve({
            deployments: response.deployments,
            notFound: response.not_found
          });
        } else {
          reject(new Error('no response received'));
        }
      });
    });
  }

  async getDeploymentTargetsForAppIDs(appIds: string[], timeout: number): Promise<{ [appId: string]: DeploymentTarget }> {
    const { deployments, notFound } = await this.getDeploymentAddresses(appIds, timeout);
    
    const result: { [appId: string]: DeploymentTarget } = {};
    
    // Process successful deployments
    for (const [appId, deployment] of Object.entries(deployments)) {
      if (!deployment.container_ip || !deployment.deployment_client_address) {
        console.warn(`⚠️  App ID ${appId} missing container IP or deployment client address`);
        continue;
      }
      
      // Parse the deployment client address to extract host and port
      const addressParts = deployment.deployment_client_address.split(':');
      const address = addressParts[0];
      const port = addressParts.length > 1 ? parseInt(addressParts[1], 10) : 50053; // Default voting port
      
      result[appId] = {
        appID: appId,
        address,
        port,
        containerIP: deployment.container_ip
      };
    }
    
    // Log not found app IDs
    for (const appId of notFound) {
      console.warn(`⚠️  App ID ${appId} not found or not deployed`);
    }
    
    return result;
  }

  close(): void {
    if (this.grpcConnection) {
      this.grpcConnection.close();
      this.grpcConnection = null;
      this.client = null;
    }
  }
}