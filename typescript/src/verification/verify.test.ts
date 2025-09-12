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

import * as crypto from 'crypto';
import { ec as EC } from 'elliptic';
import * as ed from '@noble/ed25519';
import { verifySignature } from './verify';
import { Protocol, Curve } from '../types';

// Test ED25519 verification
async function testED25519() {
  console.log('Testing ED25519 verification...');
  
  // Generate key pair using crypto.randomBytes
  const privateKey = crypto.randomBytes(32);
  const publicKey = ed.getPublicKey(privateKey);
  
  const message = Buffer.from('Hello, TEENet!');
  
  // Sign the message
  const signature = ed.sign(message, privateKey);
  
  // Test valid signature
  const valid = await verifySignature(
    message,
    Buffer.from(publicKey),
    Buffer.from(signature),
    Protocol.ECDSA, // ignored for ED25519
    Curve.ED25519
  );
  
  console.log(`✅ Valid ED25519 signature: ${valid}`);
  
  // Test invalid signature
  const invalidSig = Buffer.from(signature);
  invalidSig[0] ^= 0xFF;
  
  const invalid = await verifySignature(
    message,
    Buffer.from(publicKey),
    invalidSig,
    Protocol.ECDSA,
    Curve.ED25519
  );
  
  console.log(`✅ Invalid ED25519 signature rejected: ${!invalid}`);
}

// Test SECP256K1 ECDSA verification
function testSecp256k1ECDSA() {
  console.log('\nTesting SECP256K1 ECDSA verification...');
  
  const ec = new EC('secp256k1');
  
  // Generate key pair
  const keyPair = ec.genKeyPair();
  const message = Buffer.from('Bitcoin transaction');
  
  // Hash the message
  const msgHash = crypto.createHash('sha256').update(message).digest();
  
  // Sign with ECDSA
  const signature = keyPair.sign(msgHash);
  
  // Get signature in raw format (64 bytes)
  const r = signature.r.toBuffer('be', 32);
  const s = signature.s.toBuffer('be', 32);
  const rawSig = Buffer.concat([r, s]);
  
  // Test with different public key formats
  const formats = [
    { name: 'Uncompressed', key: Buffer.from(keyPair.getPublic().encode('array', false)) },
    { name: 'Compressed', key: Buffer.from(keyPair.getPublic().encode('array', true)) },
    { name: 'Raw (64 bytes)', key: Buffer.from(keyPair.getPublic().encode('array', false).slice(1)) }
  ];
  
  formats.forEach(async format => {
    const valid = await verifySignature(
      message,
      format.key,
      rawSig,
      Protocol.ECDSA,
      Curve.SECP256K1
    );
    console.log(`✅ ${format.name} format: ${valid}`);
  });
}

// Test SECP256R1 (P-256) ECDSA verification
function testSecp256r1ECDSA() {
  console.log('\nTesting SECP256R1 (P-256) ECDSA verification...');
  
  const ec = new EC('p256');
  
  // Generate key pair
  const keyPair = ec.genKeyPair();
  const message = Buffer.from('Web authentication');
  
  // Hash the message
  const msgHash = crypto.createHash('sha256').update(message).digest();
  
  // Sign with ECDSA
  const signature = keyPair.sign(msgHash);
  
  // Get signature in raw format (64 bytes)
  const r = signature.r.toBuffer('be', 32);
  const s = signature.s.toBuffer('be', 32);
  const rawSig = Buffer.concat([r, s]);
  
  // Test with uncompressed public key
  const publicKey = Buffer.from(keyPair.getPublic().encode('array', false));
  
  verifySignature(
    message,
    publicKey,
    rawSig,
    Protocol.ECDSA,
    Curve.SECP256R1
  ).then(valid => {
    console.log(`✅ P-256 ECDSA verification: ${valid}`);
  });
  
  // Test with compressed public key
  const compressedKey = Buffer.from(keyPair.getPublic().encode('array', true));
  
  verifySignature(
    message,
    compressedKey,
    rawSig,
    Protocol.ECDSA,
    Curve.SECP256R1
  ).then(valid => {
    console.log(`✅ P-256 compressed key verification: ${valid}`);
  });
}

// Run all tests
async function runTests() {
  console.log('=== TypeScript Verification Tests ===\n');
  
  await testED25519();
  testSecp256k1ECDSA();
  testSecp256r1ECDSA();
  
  console.log('\n✅ All tests completed!');
}

// Run if executed directly
if (require.main === module) {
  runTests().catch(console.error);
}

export { runTests };