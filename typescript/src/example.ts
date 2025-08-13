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

import { Client } from './client';
import { Protocol, Curve } from './types';

async function main() {
  // Configuration
  const configServerAddr = 'localhost:50052'; // TEE config server address
  
  console.log('=== TEE DAO Key Management Client with AppID Service Integration ===');

  // Create client
  const client = new Client(configServerAddr);

  try {
    await client.init();
    console.log(`Client connected, Node ID: ${client.getNodeId()}`);

    // Example: Get public key by app ID
    console.log('\n1. Get public key by app ID');
    const appID = 'secure-messaging-app'; // Replace with actual app ID
    
    try {
      const { publickey, protocol, curve } = await client.getPublicKeyByAppID(appID);
      console.log(`Public key for app ID ${appID}:`);
      console.log(`  - Protocol: ${protocol}`);
      console.log(`  - Curve: ${curve}`);
      console.log(`  - Public Key: ${publickey}`);
    } catch (error) {
      console.error(`Failed to get public key by app ID: ${error}`);
    }

    // Example: Sign with app ID
    console.log('\n2. Sign message with app ID');
    const message = new TextEncoder().encode('Hello from AppID Service!');

    try {
      const signature = await client.signWithAppID(message, appID);
      console.log('Signing with app ID successful!');
      console.log(`Message: ${new TextDecoder().decode(message)}`);
      console.log(`Signature: ${Buffer.from(signature).toString('hex')}`);
    } catch (error) {
      console.error(`Signing with app ID failed: ${error}`);
    }

    console.log('\n=== Example completed ===');

  } catch (error) {
    console.error('Client initialization failed:', error);
  } finally {
    await client.close();
  }
}

if (require.main === module) {
  main().catch(console.error);
}