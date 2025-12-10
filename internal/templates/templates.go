package templates

import (
	"html/template"
)

// HomeTemplate is the HTML template for the home page
const HomeTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>SSH Fun - Command Executor</title>
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
            <a id="terminalLink" href="/terminal" style="display: inline-block; background: #28a745; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; font-size: 18px;">
                🖥️ Open Web Terminal
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

// TerminalTemplate is the HTML template for the terminal page
const TerminalTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>SSH Fun - Web Terminal</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@400;600;700&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.min.js"></script>
    <style>
        body {
            margin: 0;
            padding: 0;
            background-color: #000000;
            font-family: 'Courier New', monospace;
            color: #ffffff;
            overflow: hidden;
            height: 100vh;
        }
        .header {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            background-color: #ffffff;
            padding: 10px 20px;
            z-index: 1000;
            border-bottom: 2px solid #333333;
            font-family: 'Open Sans', sans-serif;
        }
        .header h1 {
            color: #333333;
            margin: 0;
            font-size: 18px;
            display: inline-block;
            font-family: 'Open Sans', sans-serif;
            font-weight: 600;
        }
        .header p {
            color: #666666;
            margin: 5px 0 0 0;
            font-size: 12px;
            font-family: 'Open Sans', sans-serif;
        }
        .hostname {
            color: #007bff;
            font-weight: bold;
            font-size: 14px;
            margin-left: 20px;
            font-family: 'Open Sans', sans-serif;
        }
        .status {
            position: fixed;
            top: 10px;
            right: 20px;
            color: #00ff00;
            font-size: 12px;
            z-index: 1001;
        }
        .terminal-container {
            position: fixed;
            top: 60px;
            left: 0;
            right: 0;
            bottom: 60px;
            background-color: #000000;
            padding: 10px;
            margin: 0;
        }
        .controls {
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            background-color: #ffffff;
            padding: 10px 20px;
            border-top: 2px solid #333333;
            z-index: 1000;
        }
        .btn {
            background-color: #333333;
            color: #ffffff;
            border: none;
            padding: 8px 16px;
            margin: 0 5px;
            border-radius: 3px;
            cursor: pointer;
            font-weight: bold;
            font-size: 12px;
        }
        .btn:hover {
            background-color: #555555;
        }
        .btn:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        .back-link {
            position: fixed;
            bottom: 10px;
            right: 20px;
            z-index: 1001;
        }
        .back-link a {
            color: #00ff00;
            text-decoration: none;
            font-size: 12px;
        }
        .back-link a:hover {
            text-decoration: underline;
        }
        .xterm {
            height: 100% !important;
        }
        .xterm-viewport {
            background-color: #000000 !important;
        }
        .status-badge {
            padding: 4px 8px;
            border-radius: 3px;
            font-size: 10px;
            font-weight: bold;
            margin-left: 10px;
        }
        .status-badge.connected {
            background-color: #28a745;
            color: white;
        }
        .status-badge.disconnected {
            background-color: #dc3545;
            color: white;
        }
        .status-badge.connecting {
            background-color: #ffc107;
            color: black;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>SSH - Web Terminal</h1>
        <span class="hostname">@{{.Hostname}}</span>
        <span class="status-badge disconnected" id="statusBadge">Disconnected</span>
        <p>WebShell</p>
        <div style="margin-top: 10px; display: flex; align-items: center; gap: 5px;">
            <input type="password" id="tokenInput" placeholder="Auth Token (optional)" style="padding: 5px 10px; border: 1px solid #ccc; border-radius: 3px; font-size: 12px; width: 200px;">
            <button onclick="saveToken()" style="padding: 5px 10px; border: none; border-radius: 3px; background: #28a745; color: white; cursor: pointer; font-size: 12px;">Save</button>
            <button onclick="clearToken()" style="padding: 5px 10px; border: none; border-radius: 3px; background: #dc3545; color: white; cursor: pointer; font-size: 12px;">Clear</button>
        </div>
    </div>
    
    <div class="status" id="status">Ready to connect</div>
    
    <div class="terminal-container" id="terminal"></div>
    
    <div class="controls">
        <button class="btn" id="connectBtn" onclick="connect()">Connect</button>
        <button class="btn" id="disconnectBtn" onclick="disconnect()" disabled>Disconnect</button>
        <button class="btn" id="clearBtn" onclick="clearTerminal()">Clear</button>
        <button class="btn" id="fullscreenBtn" onclick="toggleFullscreen()">Fullscreen</button>
    </div>
    
    <div class="back-link">
        <a href="/">← Back to Home</a>
    </div>

    <script>
        let term;
        let socket;
        let fitAddon;
        let isConnected = false;

        // Initialize terminal
        function initTerminal() {
            term = new Terminal({
                cursorBlink: true,
                fontSize: 14,
                fontFamily: 'Courier New, monospace',
                theme: {
                    background: '#000000',
                    foreground: '#ffffff',
                    cursor: '#ffffff',
                    selection: '#ffffff',
                    black: '#000000',
                    red: '#ff0000',
                    green: '#00ff00',
                    yellow: '#ffff00',
                    blue: '#0000ff',
                    magenta: '#ff00ff',
                    cyan: '#00ffff',
                    white: '#ffffff',
                    brightBlack: '#666666',
                    brightRed: '#ff6666',
                    brightGreen: '#66ff66',
                    brightYellow: '#ffff66',
                    brightBlue: '#6666ff',
                    brightMagenta: '#ff66ff',
                    brightCyan: '#66ffff',
                    brightWhite: '#ffffff'
                }
            });

            fitAddon = new FitAddon.FitAddon();
            term.loadAddon(fitAddon);
            term.open(document.getElementById('terminal'));
            fitAddon.fit();

            // Handle terminal input
            term.onData(data => {
                if (socket && socket.readyState === WebSocket.OPEN) {
                    socket.send(data);
                }
            });

            // Handle window resize
            window.addEventListener('resize', () => {
                if (fitAddon) {
                    fitAddon.fit();
                    if (socket && socket.readyState === WebSocket.OPEN) {
                        const dims = fitAddon.proposeDimensions();
                        if (dims) {
                            socket.send(JSON.stringify({
                                type: 'resize',
                                cols: dims.cols,
                                rows: dims.rows
                            }));
                        }
                    }
                }
            });

            // Auto-connect after terminal initialization
            setTimeout(() => {
                connect();
            }, 100);
        }

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
        
        // Save token to localStorage
        function saveToken() {
            const tokenInput = document.getElementById('tokenInput');
            const token = tokenInput.value.trim();
            if (token) {
                localStorage.setItem(TOKEN_STORAGE_KEY, token);
                alert('Token saved');
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
        }
        
        // Load token from storage on page load
        function loadToken() {
            const storedToken = localStorage.getItem(TOKEN_STORAGE_KEY);
            const tokenInput = document.getElementById('tokenInput');
            if (storedToken && tokenInput) {
                tokenInput.value = storedToken;
            }
            
            // Also check URL parameter
            const urlParams = new URLSearchParams(window.location.search);
            const urlToken = urlParams.get('token');
            if (urlToken && tokenInput && !tokenInput.value) {
                tokenInput.value = urlToken;
                localStorage.setItem(TOKEN_STORAGE_KEY, urlToken);
            }
        }
        
        // Connect to WebSocket
        function connect() {
            if (isConnected) return;

            updateStatus('Connecting...', 'connecting');
            updateButtons(true, false);

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            // Get token from multiple sources
            const token = getToken();
            let wsUrl = protocol + '//' + window.location.host + '/ws';
            if (token) {
                wsUrl += '?token=' + encodeURIComponent(token);
            }
            
            socket = new WebSocket(wsUrl);

            socket.onopen = function(event) {
                isConnected = true;
                updateStatus('Connected', 'connected');
                updateButtons(false, true);
                term.write('\r\nConnected to WebShell\r\n');
                
                // Send initial resize
                const dims = fitAddon.proposeDimensions();
                if (dims) {
                    socket.send(JSON.stringify({
                        type: 'resize',
                        cols: dims.cols,
                        rows: dims.rows
                    }));
                }
            };

            socket.onmessage = function(event) {
                term.write(event.data);
            };

            socket.onclose = function(event) {
                isConnected = false;
                updateStatus('Disconnected', 'disconnected');
                updateButtons(false, false);
                term.write('\r\nDisconnected from WebShell\r\n');
                
                // Auto-reconnect after 2 seconds
                setTimeout(() => {
                    if (!isConnected) {
                        term.write('\r\nReconnecting...\r\n');
                        connect();
                    }
                }, 2000);
            };

            socket.onerror = function(error) {
                isConnected = false;
                updateStatus('Connection Error', 'disconnected');
                updateButtons(false, false);
                term.write('\r\nConnection error\r\n');
                console.error('WebSocket error:', error);
            };
        }

        // Disconnect from WebSocket
        function disconnect() {
            if (socket) {
                socket.close();
            }
        }

        // Clear terminal
        function clearTerminal() {
            term.clear();
        }

        // Toggle fullscreen
        function toggleFullscreen() {
            if (!document.fullscreenElement) {
                document.documentElement.requestFullscreen();
            } else {
                document.exitFullscreen();
            }
        }

        // Update status display
        function updateStatus(message, badgeClass) {
            document.getElementById('status').textContent = message;
            const badge = document.getElementById('statusBadge');
            badge.textContent = message;
            badge.className = 'status-badge ' + badgeClass;
        }

        // Update button states
        function updateButtons(connecting, connected) {
            document.getElementById('connectBtn').disabled = connecting || connected;
            document.getElementById('disconnectBtn').disabled = !connected;
        }

        // Initialize when page loads
        window.addEventListener('load', function() {
            loadToken();
            initTerminal();
        });
    </script>
</body>
</html>`

var (
	homeTemplate     *template.Template
	terminalTemplate *template.Template
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

// GetTerminalTemplate returns the parsed terminal page template
func GetTerminalTemplate() (*template.Template, error) {
	if terminalTemplate == nil {
		var err error
		terminalTemplate, err = template.New("terminal").Parse(TerminalTemplate)
		if err != nil {
			return nil, err
		}
	}
	return terminalTemplate, nil
}
