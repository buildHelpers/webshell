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
            <input type="password" id="tokenInput" autocomplete="off" placeholder="Authentication token" style="width: 400px; padding: 10px; margin: 5px; border: 1px solid #ddd; border-radius: 3px;">
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
curl -X POST https://localhost:8080/execute -d "ls -la"

# With authentication token (if enabled)
curl -X POST https://localhost:8080/execute \
  -H "X-Auth-Token: your-token-here" \
  -d "ls -la"

# Command with arguments (returns raw output)
curl -X POST https://localhost:8080/execute -d "find . -name '*.go'"

# Get JSON response with metadata
curl -X POST https://localhost:8080/execute \
  -H "Accept: application/json" \
  -H "X-Auth-Token: your-token-here" \
  -d "uname -a"

        </div>
    </div>
    
    <script>
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
                const basePath = window.location.pathname;
                const headers = {};
                const token = document.getElementById('tokenInput').value;
                if (token) {
                    headers['X-Auth-Token'] = token;
                }
                
                const response = await fetch(basePath + 'execute', {
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
        
        document.getElementById('terminalLink').href = './terminal';
    </script>
</body>
</html>`

var homeTemplate, homeTemplateError = template.New("home").Parse(HomeTemplate)

func GetHomeTemplate() (*template.Template, error) {
	return homeTemplate, homeTemplateError
}
