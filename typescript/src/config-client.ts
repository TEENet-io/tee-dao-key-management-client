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
import { NodeConfig, NodeType, Constants } from './types';

interface CLIRPCServiceClient {
  GetNodeInfo(request: any, callback: (error: grpc.ServiceError | null, response?: any) => void): grpc.ClientUnaryCall;
  GetPeerNode(request: any, callback: (error: grpc.ServiceError | null, response?: any) => void): grpc.ClientUnaryCall;
}

export class ConfigClient {
  private serverAddress: string;
  private timeout: number;

  constructor(serverAddress: string) {
    this.serverAddress = serverAddress;
    this.timeout = Constants.DEFAULT_CONFIG_TIMEOUT;
  }

  async getConfig(timeout?: number): Promise<NodeConfig> {
    const actualTimeout = timeout || this.timeout;
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error('config request timeout'));
      }, actualTimeout);

      this.fetchFromServer()
        .then(config => {
          clearTimeout(timer);
          resolve(config);
        })
        .catch(err => {
          clearTimeout(timer);
          reject(err);
        });
    });
  }

  private async fetchFromServer(): Promise<NodeConfig> {
    let client: CLIRPCServiceClient | null = null;
    
    try {
      const protoPath = path.resolve(__dirname, '../proto/node_management/node_management.proto');
      const packageDefinition = protoLoader.loadSync(protoPath, {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
      });

      const nodeManagementProto = grpc.loadPackageDefinition(packageDefinition) as any;
      const CLIRPCServiceClient = nodeManagementProto.tee_node_management.CLIRPCService;

      client = new CLIRPCServiceClient(
        this.serverAddress,
        grpc.credentials.createInsecure()
      );

      const nodeInfo = await this.getNodeInfo(client!);
      const peers = await this.getPeerNodes(client!);

      // Find TEE node and App node
      const teeNode = peers.peers?.find((peer: any) => peer.type === NodeType.TEE_NODE);
      const appNode = peers.peers?.find((peer: any) => peer.type === NodeType.APP_NODE);
      
      if (!teeNode && !appNode) {
        throw new Error('no TEE or App node found');
      }

      const config: NodeConfig = {
        nodeId: nodeInfo.node_id,
        rpcAddress: teeNode.rpc_address,
        cert: nodeInfo.cert.data ? Buffer.from(nodeInfo.cert.data) : Buffer.from(nodeInfo.cert),
        key: nodeInfo.key.data ? Buffer.from(nodeInfo.key.data) : Buffer.from(nodeInfo.key),
        targetCert: teeNode.cert.data ? Buffer.from(teeNode.cert.data) : Buffer.from(teeNode.cert),
        appNodeAddr: appNode.rpc_address,
        appNodeCert: appNode.cert.data ? Buffer.from(appNode.cert.data) : Buffer.from(appNode.cert),
      };

      console.log(`Retrieved config from server, node ID: ${config.nodeId}`);
      return config;
    } catch (error) {
      throw new Error(`failed to fetch config from server: ${error}`);
    } finally {
      if (client) {
        (client as any).close();
      }
    }
  }

  private async getNodeInfo(client: CLIRPCServiceClient): Promise<any> {
    return new Promise((resolve, reject) => {
      client.GetNodeInfo({}, (error: grpc.ServiceError | null, response?: any) => {
        if (error) {
          reject(new Error(`failed to get node info: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  private async getPeerNodes(client: CLIRPCServiceClient): Promise<any> {
    return new Promise((resolve, reject) => {
      client.GetPeerNode({}, (error: grpc.ServiceError | null, response?: any) => {
        if (error) {
          reject(new Error(`failed to get peer nodes: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  setTimeout(timeout: number): void {
    this.timeout = timeout;
  }
}