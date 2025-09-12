// Use dynamically loaded APP_ID from window.FIXED_APP_ID set by index.html
function getAppId() {
    return window.FIXED_APP_ID || "default-app-id";
}

// Dynamic API base path detection - works for both direct access and proxy access
function getApiBasePath() {
    const currentPath = window.location.pathname;
    // If accessed through proxy, keep the current path as base
    // If accessed directly, use empty base
    return currentPath.endsWith('/') ? currentPath : currentPath + '/';
}

async function makeApiCall(endpoint, options = {}) {
    const basePath = getApiBasePath();
    const url = basePath + 'api/' + endpoint;
    return fetch(url, options);
}

async function getPublicKey() {
    const resultDiv = document.getElementById('publicKeyResult');

    showResult(resultDiv, 'Getting public key...', 'loading');

    try {
        const response = await makeApiCall('get-public-key', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ app_id: getAppId() })
        });

        const data = await response.json();
        
        if (data.success) {
            const result = JSON.stringify({
                app_id: data.app_id,
                protocol: data.protocol,
                curve: data.curve,
                public_key: data.public_key
            }, null, 2);
            showResult(resultDiv, result, 'success');
            
            // Note: Advanced form elements were removed, no auto-fill needed
        } else {
            showResult(resultDiv, 'Error: ' + data.error, 'error');
        }
    } catch (error) {
        showResult(resultDiv, 'Network error: ' + error.message, 'error');
    }
}

async function signWithAppID() {
    const message = document.getElementById('message1').value.trim();
    const resultDiv = document.getElementById('signAppIDResult');
    
    if (!message) {
        showResult(resultDiv, 'Please enter a message', 'error');
        return;
    }

    showResult(resultDiv, 'Signing message...', 'loading');

    try {
        const response = await makeApiCall('sign-with-appid', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                app_id: getAppId(),
                message: message 
            })
        });

        const data = await response.json();
        
        if (data.success) {
            const result = JSON.stringify({
                message: data.message,
                app_id: data.app_id,
                signature: data.signature
            }, null, 2);
            showResult(resultDiv, result, 'success');
            
            // Auto-fill verification form with the latest signature
            document.getElementById('verifyMessage1').value = message;
            document.getElementById('verifySignature1').value = data.signature;
        } else {
            showResult(resultDiv, 'Error: ' + data.error, 'error');
        }
    } catch (error) {
        showResult(resultDiv, 'Network error: ' + error.message, 'error');
    }
}


async function verifyWithAppID() {
    const message = document.getElementById('verifyMessage1').value.trim();
    const signature = document.getElementById('verifySignature1').value.trim();
    const resultDiv = document.getElementById('verifyAppIDResult');
    
    if (!message || !signature) {
        showResult(resultDiv, 'Please enter message and signature', 'error');
        return;
    }

    showResult(resultDiv, 'Verifying signature...', 'loading');

    try {
        const response = await makeApiCall('verify-with-appid', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                app_id: getAppId(),
                message: message,
                signature: signature
            })
        });

        const data = await response.json();
        
        if (data.success) {
            const result = JSON.stringify({
                valid: data.valid,
                message: data.message,
                app_id: data.app_id,
                public_key: data.public_key,
                protocol: data.protocol,
                curve: data.curve,
                verification_result: data.valid ? '✅ Valid signature' : '❌ Invalid signature'
            }, null, 2);
            showResult(resultDiv, result, data.valid ? 'success' : 'error');
        } else {
            showResult(resultDiv, 'Error: ' + data.error, 'error');
        }
    } catch (error) {
        showResult(resultDiv, 'Network error: ' + error.message, 'error');
    }
}

// Check if message contains approval keywords
function checkMessageApproval() {
    const message = document.getElementById('votingMessage').value.toLowerCase();
    const tipElement = document.getElementById('approvalTip');
    
    if (!tipElement) {
        // Create tip element if it doesn't exist
        const tip = document.createElement('div');
        tip.id = 'approvalTip';
        tip.style.cssText = 'margin-top: 8px; font-size: 13px; padding: 8px 12px; border-radius: 4px; transition: all 0.3s ease;';
        document.getElementById('votingMessage').parentNode.appendChild(tip);
    }
    
    const tip = document.getElementById('approvalTip');
    
    if (message.includes('test')) {
        tip.innerHTML = '✅ <strong>Good!</strong> Your message contains "test" - demo nodes will approve this message';
        tip.style.backgroundColor = '#f6ffed';
        tip.style.color = '#52c41a';
        tip.style.border = '1px solid #b7eb8f';
    } else {
        tip.innerHTML = '⚠️ <strong>Note:</strong> Your message doesn\'t contain "test" - demo nodes may reject this vote';
        tip.style.backgroundColor = '#fff7e6';
        tip.style.color = '#fa8c16';
        tip.style.border = '1px solid #ffd591';
    }
}

// Voting functionality
async function initiateVoting() {
    const message = document.getElementById('votingMessage').value.trim();
    const resultDiv = document.getElementById('votingResult');
    
    if (!message) {
        showResult(resultDiv, 'Please enter a message to sign', 'error');
        return;
    }

    showResult(resultDiv, 'Initiating voting round...', 'loading');

    try {
        const response = await makeApiCall('vote', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: btoa(message), // base64 encode
                signer_app_id: getAppId(),
                is_forwarded: false
                // Target app IDs and required votes are fetched from server
            })
        });

        const data = await response.json();
        
        if (data.success) {
            // Check if there's an error in voting_results
            const hasError = data.voting_results?.error;
            
            const result = JSON.stringify({
                voting_complete: data.voting_results?.voting_complete,
                successful_votes: data.voting_results?.successful_votes,
                required_votes: data.voting_results?.required_votes,
                total_responses: data.voting_results?.total_targets, // 使用 total_targets 字段
                final_result: data.voting_results?.final_result,
                vote_details: data.voting_results?.vote_details || [],
                signature: data.signature || 'No signature',
                timestamp: data.timestamp,
                error: hasError ? data.voting_results.error : undefined
            }, null, 2);
            
            // Show as error if there's an error field, otherwise as success
            showResult(resultDiv, result, hasError ? 'error' : 'success');
            
            // Auto-fill voting verification form if voting was successful
            if (data.signature) {
                document.getElementById('verifyVotingMessage').value = message;
                document.getElementById('verifyVotingSignature').value = data.signature;
                document.getElementById('verifyVotingAppId').value = getAppId();
            }
        } else {
            showResult(resultDiv, 'Error: ' + data.message, 'error');
        }
    } catch (error) {
        showResult(resultDiv, 'Network error: ' + error.message, 'error');
    }
}


// Verify voting signature
async function verifyVotingSignature() {
    const message = document.getElementById('verifyVotingMessage').value.trim();
    const signature = document.getElementById('verifyVotingSignature').value.trim();
    const appId = document.getElementById('verifyVotingAppId').value.trim();
    const resultDiv = document.getElementById('verifyVotingResult');
    
    if (!message || !signature || !appId) {
        showResult(resultDiv, 'Please fill in all fields', 'error');
        return;
    }

    showResult(resultDiv, 'Verifying voting signature...', 'loading');

    try {
        const response = await makeApiCall('verify-with-appid', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                app_id: appId,
                message: message,
                signature: signature
            })
        });

        const data = await response.json();
        
        if (data.success) {
            const result = JSON.stringify({
                valid: data.valid,
                message: data.message,
                app_id: data.app_id,
                public_key: data.public_key,
                protocol: data.protocol,
                curve: data.curve,
                verification_result: data.valid ? '✅ Valid voting signature' : '❌ Invalid voting signature',
                signature_type: 'Multi-party voting signature'
            }, null, 2);
            showResult(resultDiv, result, data.valid ? 'success' : 'error');
        } else {
            showResult(resultDiv, 'Error: ' + data.error, 'error');
        }
    } catch (error) {
        showResult(resultDiv, 'Network error: ' + error.message, 'error');
    }
}


function showResult(element, content, type) {
    element.textContent = content;
    element.className = 'result ' + type;
    element.style.display = 'block';
}

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    // Check message approval on page load
    setTimeout(function() {
        checkMessageApproval();
    }, 100);
});