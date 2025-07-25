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

import { NodeConfig, ClientOptions, Constants } from './types';
import { ConfigClient } from './config-client';
import { TaskClient } from './task-client';

export class Client {
  private configClient: ConfigClient;
  private taskClient: TaskClient | null = null;
  private config: NodeConfig | null = null;
  private timeout: number;

  constructor(configServerAddress: string, options?: Partial<ClientOptions>) {
    this.configClient = new ConfigClient(configServerAddress);
    this.timeout = options?.timeout || Constants.DEFAULT_CLIENT_TIMEOUT;
  }

  async init(): Promise<void> {
    const config = await this.configClient.getConfig(this.timeout);
    this.config = config;

    this.taskClient = new TaskClient(config);
    await this.taskClient.connect(this.timeout);

    console.log(`Client initialized successfully, node ID: ${config.nodeId}`);
  }

  async close(): Promise<void> {
    if (this.taskClient) {
      await this.taskClient.close();
      this.taskClient = null;
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
}