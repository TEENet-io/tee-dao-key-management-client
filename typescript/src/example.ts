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
  // Use correct gRPC address format: host:port
  const configServerAddr = 'localhost:50052';  // Config server address
  const client = new Client(configServerAddr);

  try {
    console.log('Initializing client...');
    await client.init();

    // 3. Execute signing (client now only supports signing)
    // Note: In a real scenario, you would get the public key from elsewhere (e.g., DKG service)
    const publicKey = new TextEncoder().encode('example-public-key-from-dkg-service'); // Placeholder
    const message = new TextEncoder().encode('Hello, TEE DAO!');
    
    console.log('Signing message...');
    const signature = await client.sign(message, publicKey, Protocol.ECDSA, Curve.ED25519);
    console.log(`Signing successful!`);
    console.log(`Message: ${new TextDecoder().decode(message)}`);
    console.log(`Signature: ${Buffer.from(signature).toString('hex')}`);

    console.log('Node ID:', client.getNodeId());

  } catch (error) {
    console.error('Error:', error);
  } finally {
    console.log('Closing client...');
    await client.close();
  }
}

async function customTimeoutExample() {
  const client = new Client('localhost:50052');

  client.setTimeout(60000);

  try {
    await client.init();
    
    const publicKey = new TextEncoder().encode('example-public-key');
    const message = new TextEncoder().encode('Custom timeout message');
    const signature = await client.sign(message, publicKey, Protocol.SCHNORR, Curve.SECP256K1);
    console.log('Custom timeout signing completed');
    
  } catch (error) {
    console.error('Custom timeout example error:', error);
  } finally {
    await client.close();
  }
}

if (require.main === module) {
  main().catch(console.error);
}