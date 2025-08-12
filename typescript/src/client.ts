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

import { NodeConfig, ClientOptions, Constants, Protocol, Curve } from './types';
import { ConfigClient } from './config-client';
import { TaskClient } from './task-client';
import { AppIDClient } from './appid-client';
import * as tls from 'tls';

export class Client {
  private configClient: ConfigClient;
  private taskClient: TaskClient | null = null;
  private appIDClient: AppIDClient | null = null;
  private config: NodeConfig | null = null;
  private timeout: number;

  constructor(configServerAddress: string, options?: Partial<ClientOptions>) {
    this.configClient = new ConfigClient(configServerAddress);
    this.timeout = options?.timeout || Constants.DEFAULT_CLIENT_TIMEOUT;
  }

  async init(): Promise<void> {
    // 1. Fetch configuration
    const config = await this.configClient.getConfig(this.timeout);
    this.config = config;

    // 2. Create task client
    this.taskClient = new TaskClient(config);
    
    // 3. Create TLS configuration for TEE server
    const teeTLSConfig = this.createTEETLSConfig();
    
    // 4. Connect to TEE server
    await this.taskClient.connect(this.timeout);

    // 5. Create AppID client
    this.appIDClient = new AppIDClient(config.appNodeAddr);
    
    // 6. Create TLS configuration for App node
    const appTLSConfig = this.createAppTLSConfig();
    
    // 7. Connect to user management system
    await this.appIDClient.connect(appTLSConfig);

    console.log(`Client initialized successfully, node ID: ${config.nodeId}`);
  }

  async close(): Promise<void> {
    if (this.taskClient) {
      await this.taskClient.close();
      this.taskClient = null;
    }
    if (this.appIDClient) {
      this.appIDClient.close();
      this.appIDClient = null;
    }
  }

  async sign(message: Uint8Array, publicKey: Uint8Array, protocol: number, curve: number): Promise<Uint8Array> {
    if (!this.taskClient) {
      throw new Error('client not initialized');
    }
    return this.taskClient.sign(message, publicKey, protocol, curve, this.timeout);
  }

  getNodeId(): number {
    if (!this.config) {
      return 0;
    }
    return this.config.nodeId;
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
    if (!this.config) {
      throw new Error('config not loaded');
    }
    return {
      cert: this.config.cert,
      key: this.config.key,
      ca: this.config.targetCert,
    };
  }

  // Create TLS configuration for App node (user management system)
  private createAppTLSConfig(): tls.SecureContextOptions {
    if (!this.config) {
      throw new Error('config not loaded');
    }
    return {
      cert: this.config.cert,
      key: this.config.key,
      ca: this.config.appNodeCert,
    };
  }

  // Get public key by app ID from user management system
  async getPublicKeyByAppID(appId: string): Promise<{publickey: string, protocol: string, curve: string}> {
    if (!this.appIDClient) {
      throw new Error('AppID client not initialized');
    }
    return this.appIDClient.getPublicKeyByAppID(appId);
  }

  // Sign with app ID (combines getPublicKeyByAppID and sign)
  async signWithAppID(message: Uint8Array, appId: string): Promise<Uint8Array> {
    // Get public key from user management system
    const {publickey, protocol, curve} = await this.getPublicKeyByAppID(appId);
    
    // Parse protocol and curve
    const protocolNum = this.parseProtocol(protocol);
    const curveNum = this.parseCurve(curve);
    
    // Decode public key from base64
    const publicKeyBuffer = Buffer.from(publickey, 'base64');
    
    // Sign the message
    return this.sign(message, new Uint8Array(publicKeyBuffer), protocolNum, curveNum);
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
}