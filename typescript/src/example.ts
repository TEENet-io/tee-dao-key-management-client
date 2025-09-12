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

import { Client, SignRequest, SignResult } from './index';
// @ts-ignore
import * as wtfnode from 'wtfnode';

async function main() {
  // Configuration
  const configServerAddr = 'localhost:50052'; // TEE config server address
  
  console.log('=== TEE DAO Key Management Client with AppID Service Integration ===');

  // Create client
  const teeClient = new Client(configServerAddr);

  try {
    // Initialize client with default voting handler (auto-approve)
    await teeClient.init();
    console.log('Client initialized successfully');

    // Example: Get public key by app ID
    console.log('\n1. Get public key by app ID');
    const appID = 'secure-messaging-app'; // Replace with actual app ID
    
    try {
      const { publickey, protocol, curve } = await teeClient.getPublicKeyByAppID(appID);
      console.log(`Public key for app ID ${appID}:`);
      console.log(`  - Protocol: ${protocol}`);
      console.log(`  - Curve: ${curve}`);
      console.log(`  - Public Key: ${publickey}`);
    } catch (error) {
      console.error(`Failed to get public key by app ID: ${error}`);
    }

    // Example: Sign message using Sign method
    console.log('\n2. Sign message');
    const message = new TextEncoder().encode('Hello from AppID Service!');

    const signReq: SignRequest = {
      message: message,
      appID: appID,
      enableVoting: false
    };

    try {
      const signResult = await teeClient.sign(signReq);
      if (signResult.success && signResult.signature) {
        console.log('Signing successful!');
        console.log(`Message: ${new TextDecoder().decode(message)}`);
        console.log(`Signature: ${Buffer.from(signResult.signature).toString('hex')}`);
        console.log(`Success: ${signResult.success}`);
        if (signResult.error) {
          console.log(`Error: ${signResult.error}`);
        }
      } else {
        console.error(`Signing failed: ${signResult.error}`);
      }
    } catch (error) {
      console.error(`Signing failed: ${error}`);
    }

    // Example: Multi-party voting signature
    console.log('\n3. Multi-party voting signature example');
    const votingMessage = new TextEncoder().encode('test message for multi-party voting'); // Contains "test" to trigger approval

    console.log('Voting request:');
    console.log(`  - Message: ${new TextDecoder().decode(votingMessage)}`);
    console.log(`  - Signer App ID: ${appID}`);
    console.log(`  - Voting Enabled: true`);

    // Create HTTP request body similar to signature-tool
    const requestData = {
      message: Buffer.from(votingMessage).toString('base64'),
      signer_app_id: appID,
      is_forwarded: false
    };

    const requestBody = Buffer.from(JSON.stringify(requestData));

    // Create a mock HTTP request like signature-tool does
    const { IncomingMessage } = require('http');
    const httpReq = new IncomingMessage(null as any);
    httpReq.method = 'POST';
    httpReq.url = '/vote';
    httpReq.headers = {
      'content-type': 'application/json'
    };
    // Add body to request (simulating parsed body)
    (httpReq as any).body = JSON.stringify(requestData);

    // Make vote decision: approve if message contains "test"
    const localApproval = new TextDecoder().decode(votingMessage).toLowerCase().includes('test');
    console.log(`  - Local Approval: ${localApproval}`);

    // Sign with voting enabled
    const votingSignReq: SignRequest = {
      message: votingMessage,
      appID: appID,
      enableVoting: true,
      localApproval: localApproval,
      httpRequest: httpReq
    };

    let votingSignResult: SignResult | undefined;
    try {
      votingSignResult = await teeClient.sign(votingSignReq);
      if (votingSignResult.success) {
        console.log('\nVoting signature completed!');
        console.log(`Success: ${votingSignResult.success}`);
        if (votingSignResult.signature) {
          console.log(`Signature: ${Buffer.from(votingSignResult.signature).toString('hex')}`);
        }
        
        // Display voting information if available
        if (votingSignResult.votingInfo) {
          console.log('\nVoting Details:');
          console.log(`  - Total Targets: ${votingSignResult.votingInfo.totalTargets}`);
          console.log(`  - Successful Votes: ${votingSignResult.votingInfo.successfulVotes}`);
          console.log(`  - Required Votes: ${votingSignResult.votingInfo.requiredVotes}`);
          
          console.log('\nIndividual Votes:');
          votingSignResult.votingInfo.voteDetails.forEach((vote: any, i: number) => {
            console.log(`  ${i + 1}. Client ${vote.clientId}: Success=${vote.success}`);
          });
        }
        
        if (votingSignResult.error) {
          console.log(`Error: ${votingSignResult.error}`);
        }
      } else {
        console.error(`Voting signature failed: ${votingSignResult.error}`);
      }
    } catch (error) {
      console.error(`Voting signature failed: ${error}`);
    }

    // Example: Verify signature
    console.log('\n4. Verify signature');
    try {
      // First, let's sign a message to get a signature to verify
      const verifyTestMessage = Buffer.from('Test message for verification');
      const verifySignReq: SignRequest = {
        message: verifyTestMessage,
        appID: appID,
        enableVoting: false
      };
      
      const verifySignResult = await teeClient.sign(verifySignReq);
      if (verifySignResult.success && verifySignResult.signature) {
        // Now verify the signature we just created
        const isValid = await teeClient.verify(
          verifyTestMessage, 
          Buffer.from(verifySignResult.signature), 
          appID
        );
        console.log(`Signature verification result: ${isValid}`);
        console.log(`  - Message: ${verifyTestMessage.toString()}`);
        console.log(`  - Signature: ${Buffer.from(verifySignResult.signature).toString('hex')}`);
        console.log(`  - App ID: ${appID}`);
        console.log(`  - Valid: ${isValid}`);

        // Test with wrong message
        const wrongMessage = Buffer.from('Wrong message');
        const isValidWrong = await teeClient.verify(
          wrongMessage, 
          Buffer.from(verifySignResult.signature), 
          appID
        );
        console.log(`\nVerification with wrong message: ${isValidWrong} (expected false)`);
      }
    } catch (error) {
      console.error(`Verification failed: ${error}`);
    }

    // Example: Verify voting signature
    console.log('\n5. Verify voting signature');
    if (votingSignResult && votingSignResult.signature) {
      try {
        // Verify the voting signature from section 3
        const isValid = await teeClient.verify(
          Buffer.from(votingMessage), 
          Buffer.from(votingSignResult.signature), 
          appID
        );
        console.log(`Voting signature verification result: ${isValid}`);
        console.log(`  - Message: ${new TextDecoder().decode(votingMessage)}`);
        console.log(`  - Signature: ${Buffer.from(votingSignResult.signature).toString('hex')}`);
        console.log(`  - App ID: ${appID}`);
        console.log(`  - Valid: ${isValid}`);
      } catch (error) {
        console.error(`Voting signature verification failed: ${error}`);
      }
    }

    console.log('\n=== Example completed ===');

  } catch (error) {
    console.error('Client initialization failed:', error);
  } finally {
    await teeClient.close();
    console.log('🏁 Example finished');
  }
}

if (require.main === module) {
  main().catch(console.error);
}