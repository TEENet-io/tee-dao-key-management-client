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
// @ts-ignore
import * as wtfnode from 'wtfnode';

async function main() {
  // Configuration
  const configServerAddr = 'localhost:50052'; // TEE config server address
  
  console.log('=== TEE DAO Key Management Client with AppID Service Integration ===');

  // Create client
  const client = new Client(configServerAddr);

  try {
    // Initialize client with default voting handler (auto-approve)
    await client.init();
    console.log(`Client initialized successfully, Node ID: ${client.getNodeId()}`);

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

    // Example: Multi-party voting signature
    console.log('\n3. Multi-party voting signature');
    const targetAppIDs = ['secure-messaging-app', 'secure-messaging-app1', 'secure-messaging-app2'];
    const requiredVotes = 2;
    const votingMessage = new TextEncoder().encode('test message for multi-party voting'); // Contains "test" to trigger approval

    // Create request data similar to signature-tool
    const messageBase64 = Buffer.from(votingMessage).toString('base64');
    const requestData = {
      message: messageBase64,
      signer_app_id: appID,
      target_app_ids: targetAppIDs,
      required_votes: requiredVotes,
      is_forwarded: false
    };

    const requestBody = Buffer.from(JSON.stringify(requestData));

    // Create a mock HTTP request like signature-tool does
    const { IncomingMessage } = require('http');
    const mockReq = new IncomingMessage(null as any);
    mockReq.method = 'POST';
    mockReq.url = '/vote';
    mockReq.headers = {
      'content-type': 'application/json',
      'user-agent': 'TEE-DAO-Client/1.0'
    };
    // Add body to request (simulating parsed body)
    (mockReq as any).body = JSON.stringify(requestData);

    // Local approval decision (same logic as signature-tool)
    const localApproval = new TextDecoder().decode(votingMessage).toLowerCase().includes('test');

    console.log('Voting request:');
    console.log(`  - Message: ${new TextDecoder().decode(votingMessage)}`);
    console.log(`  - Signer App ID: ${appID}`);
    console.log(`  - Target App IDs: ${JSON.stringify(targetAppIDs)}`);
    console.log(`  - Required Votes: ${requiredVotes}/${targetAppIDs.length}`);
    console.log(`  - Local Approval: ${localApproval}`);

    try {
      // Use VotingSign with the constructed HTTP request (matching Go version exactly)
      const votingResult = await client.votingSign(mockReq, votingMessage, appID, targetAppIDs, requiredVotes, localApproval);
      console.log('\nVoting signature completed!');
      console.log(`Total Targets: ${votingResult.totalTargets}`);
      console.log(`Successful Votes: ${votingResult.successfulVotes}/${votingResult.requiredVotes}`);
      console.log(`Voting Complete: ${votingResult.votingComplete}`);
      console.log(`Final Result: ${votingResult.finalResult}`);
      
      if (votingResult.signature) {
        console.log(`Signature: ${Buffer.from(votingResult.signature).toString('hex')}`);
      } else {
        console.log('Signature: No signature (voting failed or incomplete)');
      }
      
      // Print detailed vote results
      console.log('\nVote Details:');
      votingResult.voteDetails.forEach((detail, index) => {
        const status = detail.success ? (detail.response ? 'APPROVED' : 'REJECTED') : 'FAILED';
        let output = `  ${index + 1}. ${detail.clientId}: ${status}`;
        if (detail.error) {
          output += ` (Error: ${detail.error})`;
        }
        console.log(output);
      });
    } catch (error) {
      console.error(`Voting signature failed: ${error}`);
    }


    console.log('\n=== Example completed ===');

  } catch (error) {
    console.error('Client initialization failed:', error);
  } finally {
    await client.close();
    console.log('🏁 Example finished');
  }
}

if (require.main === module) {
  main().catch(console.error);
}