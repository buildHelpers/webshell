package templates

import (
	"html/template"
)

// HomeTemplate is the HTML template for the home page
const HomeTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>SSH Fun - Command Executor</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2328a745'%3E%3Crect x='2' y='4' width='20' height='16' rx='2' fill='%2328a745'/%3E%3Cpath d='M6 8h12M6 12h8M6 16h10' stroke='white' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .method { font-weight: bold; color: #007bff; }
        .url { font-family: monospace; background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        .example { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; font-family: monospace; }
        .allowed-commands { display: grid; grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); gap: 10px; margin: 20px 0; }
        .command { background: #e9ecef; padding: 8px; text-align: center; border-radius: 3px; font-family: monospace; }
        .test-form { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .test-form input, .test-form button { padding: 10px; margin: 5px; border: 1px solid #ddd; border-radius: 3px; }
        .test-form button { background: #007bff; color: white; cursor: pointer; }
        .test-form button:hover { background: #0056b3; }
        .result { background: #e9ecef; padding: 15px; border-radius: 5px; margin: 10px 0; white-space: pre-wrap; font-family: monospace; }
        .hostname {
            color: #007bff;
            font-weight: bold;
            font-size: 14px;
            margin-left: 20px;
            font-family: 'Open Sans', sans-serif;
        }
        .status-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 600;
            font-family: 'Open Sans', sans-serif;
            margin-left: 15px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .status-badge.connected {
            background-color: #28a745;
            color: #ffffff;
        }
        .status-badge.disconnected {
            background-color: #dc3545;
            color: #ffffff;
        }
        .status-badge.connecting {
            background-color: #ffc107;
            color: #000000;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.min.js"></script>
</head>
<body>
    <div class="container">
        <h1>SSH Fun - Command Executor</h1>
        
        <h2>Available Endpoints</h2>
        
        <div class="endpoint">
            <div><span class="method">GET</span> <span class="url">/</span></div>
            <p>This page - shows usage information and available commands.</p>
        </div>
        
        <div class="endpoint">
            <div><span class="method">POST</span> <span class="url">/execute</span></div>
            <p>Execute a command using raw body text. Returns raw output by default, or JSON with Accept header.</p>
            <div class="example">
Request Body (raw text):
ls -la

Response (raw text):
total 8
drwxr-xr-x  2 user  staff  68 Dec 20 10:30 .
drwxr-xr-x  3 user  staff  102 Dec 20 10:30 ..
-rw-r--r--  1 user  staff  1234 Dec 20 10:30 main.go

Response (JSON with Accept: application/json):
{
    "success": true,
    "output": "total 8\ndrwxr-xr-x  2 user  staff  68 Dec 20 10:30 .\n...",
    "exit_code": 0,
    "duration": "15.2ms",
    "timestamp": "2023-12-20T10:30:00Z",
    "command": "ls -la"
}
            </div>
        </div>
        
        <div class="endpoint">
            <div><span class="method">GET</span> <span class="url">/health</span></div>
            <p>Health check endpoint.</p>
        </div>
        
        <div class="endpoint">
            <div><span class="method">GET</span> <span class="url">/terminal</span></div>
            <p>Interactive web SSH terminal with full shell access.</p>
        </div>
        
        <h2>Authentication</h2>
        <div class="test-form" style="margin-bottom: 20px;">
            <input type="password" id="tokenInput" placeholder="Enter authentication token (optional)" style="width: 400px; padding: 10px; margin: 5px; border: 1px solid #ddd; border-radius: 3px;">
            <button onclick="saveToken()" style="padding: 10px; margin: 5px; border: 1px solid #ddd; border-radius: 3px; background: #007bff; color: white; cursor: pointer;">Save Token</button>
            <button onclick="clearToken()" style="padding: 10px; margin: 5px; border: 1px solid #ddd; border-radius: 3px; background: #dc3545; color: white; cursor: pointer;">Clear</button>
            <div id="tokenStatus" style="margin-top: 5px; font-size: 12px; color: #666;"></div>
        </div>
        
        <h2>Test Command Execution</h2>
        <div class="test-form">
            <input type="text" id="commandInput" placeholder="Enter command (e.g., ls -la)" style="width: 300px;">
            <button onclick="executeCommand()">Execute</button>
            <div id="result" class="result" style="display: none;"></div>
        </div>
        
        <h2>Quick Access</h2>
        <div style="text-align: center; margin: 20px 0;">
            <a id="terminalLink" href="/terminal" style="display: inline-flex; align-items: center; gap: 8px; background: #28a745; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; font-size: 18px;">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" style="vertical-align: middle;">
                    <rect x="3" y="4" width="18" height="16" rx="2" fill="currentColor" opacity="0.2"/>
                    <rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="1.5" fill="none"/>
                    <line x1="6" y1="8" x2="18" y2="8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    <line x1="6" y1="12" x2="14" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    <line x1="6" y1="16" x2="16" y2="16" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    <circle cx="19" cy="6" r="1.5" fill="currentColor"/>
                </svg>
                Open Web Terminal
            </a>
        </div>
        
        <h2>Allowed Commands</h2>
        <div class="allowed-commands">
            {{range .AllowedCommands}}
            <div class="command">{{.}}</div>
            {{end}}
        </div>
        
        <h2>Security Notes</h2>
        <ul>
            <li>Only predefined commands are allowed for security</li>
            <li>Commands are executed with a 30-second timeout</li>
            <li>All command output is sanitized</li>
            <li>Use with caution in production environments</li>
        </ul>
        
        <h2>Example Usage</h2>
        <div class="example">
# Simple command (returns raw output)
curl -X POST http://localhost:8080/execute -d "ls -la"

# With authentication token (if enabled)
curl -X POST http://localhost:8080/execute \
  -H "X-Auth-Token: your-token-here" \
  -d "ls -la"

# Command with arguments (returns raw output)
curl -X POST http://localhost:8080/execute -d "find . -name '*.go'"

# Get JSON response with metadata
curl -X POST http://localhost:8080/execute \
  -H "Accept: application/json" \
  -H "X-Auth-Token: your-token-here" \
  -d "uname -a"

# Using query parameter for token
curl -X POST "http://localhost:8080/execute?token=your-token-here" \
  -d "uptime"
        </div>
    </div>
    
    <script>
        // Token storage key
        const TOKEN_STORAGE_KEY = 'webshell_auth_token';
        
        // Get token from multiple sources (priority: input > localStorage > URL)
        function getToken() {
            // First check input field
            const tokenInput = document.getElementById('tokenInput');
            if (tokenInput && tokenInput.value) {
                return tokenInput.value;
            }
            
            // Then check localStorage
            const storedToken = localStorage.getItem(TOKEN_STORAGE_KEY);
            if (storedToken) {
                return storedToken;
            }
            
            // Finally check URL parameter
            const urlParams = new URLSearchParams(window.location.search);
            return urlParams.get('token');
        }
        
        // Save token to localStorage and update UI
        function saveToken() {
            const tokenInput = document.getElementById('tokenInput');
            const token = tokenInput.value.trim();
            const statusDiv = document.getElementById('tokenStatus');
            
            if (token) {
                localStorage.setItem(TOKEN_STORAGE_KEY, token);
                statusDiv.textContent = '✓ Token saved';
                statusDiv.style.color = '#28a745';
                updateTerminalLink();
            } else {
                clearToken();
            }
        }
        
        // Clear token from storage and input
        function clearToken() {
            localStorage.removeItem(TOKEN_STORAGE_KEY);
            const tokenInput = document.getElementById('tokenInput');
            if (tokenInput) {
                tokenInput.value = '';
            }
            const statusDiv = document.getElementById('tokenStatus');
            statusDiv.textContent = 'Token cleared';
            statusDiv.style.color = '#dc3545';
            updateTerminalLink();
        }
        
        // Load token from storage on page load
        function loadToken() {
            const storedToken = localStorage.getItem(TOKEN_STORAGE_KEY);
            const tokenInput = document.getElementById('tokenInput');
            if (storedToken && tokenInput) {
                tokenInput.value = storedToken;
                const statusDiv = document.getElementById('tokenStatus');
                statusDiv.textContent = '✓ Token loaded from storage';
                statusDiv.style.color = '#28a745';
            }
            
            // Also check URL parameter
            const urlParams = new URLSearchParams(window.location.search);
            const urlToken = urlParams.get('token');
            if (urlToken && tokenInput && !tokenInput.value) {
                tokenInput.value = urlToken;
                localStorage.setItem(TOKEN_STORAGE_KEY, urlToken);
                const statusDiv = document.getElementById('tokenStatus');
                statusDiv.textContent = '✓ Token loaded from URL';
                statusDiv.style.color = '#28a745';
            }
        }
        
        // Update terminal link with token if present
        function updateTerminalLink() {
            const token = getToken();
            const link = document.getElementById('terminalLink');
            if (token && link) {
                link.href = '/terminal?token=' + encodeURIComponent(token);
            } else if (link) {
                link.href = '/terminal';
            }
        }
        
        async function executeCommand() {
            const command = document.getElementById('commandInput').value.trim();
            if (!command) {
                alert('Please enter a command');
                return;
            }
            
            const resultDiv = document.getElementById('result');
            resultDiv.style.display = 'block';
            resultDiv.textContent = 'Executing...';
            
            try {
                const headers = {};
                const token = getToken();
                if (token) {
                    headers['X-Auth-Token'] = token;
                }
                
                const response = await fetch('/execute', {
                    method: 'POST',
                    headers: headers,
                    body: command
                });
                
                if (!response.ok) {
                    const errorText = await response.text();
                    resultDiv.textContent = 'Error: ' + errorText;
                    return;
                }
                
                const data = await response.json();
                resultDiv.textContent = JSON.stringify(data, null, 2);
            } catch (error) {
                resultDiv.textContent = 'Error: ' + error.message;
            }
        }
        
        // Allow Enter key to execute command
        document.getElementById('commandInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                executeCommand();
            }
        });
        
        // Allow Enter key to save token
        document.getElementById('tokenInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                saveToken();
            }
        });
        
        // Initialize on page load
        loadToken();
        updateTerminalLink();
    </script>
</body>
</html>`

var (
	homeTemplate *template.Template
)

// GetHomeTemplate returns the parsed home page template
func GetHomeTemplate() (*template.Template, error) {
	if homeTemplate == nil {
		var err error
		homeTemplate, err = template.New("home").Parse(HomeTemplate)
		if err != nil {
			return nil, err
		}
	}
	return homeTemplate, nil
}
