package templates

import (
	"html/template"
)

// TerminalTemplate is the HTML template for the terminal page
const TerminalTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>SSH Fun - Web Terminal</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2328a745'%3E%3Crect x='2' y='4' width='20' height='16' rx='2' fill='%2328a745'/%3E%3Cpath d='M6 8h12M6 12h8M6 16h10' stroke='white' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@400;600;700&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js" integrity="sha384-xjfWUeCWdMtvpAb/SmM6lMzS6pQGcQa0loOl1d97j6Odw0vjK9nW3+dTb/bn/mwH" crossorigin="anonymous"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css" integrity="sha384-9ftsg11+LSxVUaknegCfeKvlkO9EdIPI2op725RqY87IvhyGjElmpjZlP3LhTQjn" crossorigin="anonymous" />
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.min.js" integrity="sha384-dpjGwSSISUTz2taP54Bor7qkyMR20sSO9oe11UVYnGs2/YdUBf7HW30XKQx9PCzn" crossorigin="anonymous"></script>
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
        <span class="status-badge disconnected" id="statusBadge">Disconnected</span>
        <p>WebShell</p>
        <div style="margin-top: 10px; display: flex; align-items: center; gap: 5px;">
            <input type="password" id="tokenInput" autocomplete="off" placeholder="Authentication token" style="padding: 5px 10px; border: 1px solid #ccc; border-radius: 3px; font-size: 12px; width: 200px;">
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
        <a id="homeLink" href="/">← Back to Home</a>
    </div>

    <script>
        // Get base path from current URL
        function getBasePath() {
            const path = window.location.pathname;
            // Extract base path (everything before the last segment)
            // e.g., /xx/terminal -> /xx/
            // e.g., /xx/ -> /xx/
            // e.g., / -> /
            const segments = path.split('/').filter(s => s);
            if (segments.length > 0) {
                // Remove the last segment (like 'terminal' or empty)
                segments.pop();
            }
            const basePath = '/' + segments.join('/');
            return basePath.endsWith('/') ? basePath : basePath + '/';
        }
        
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

        }

        // Connect to WebSocket
        async function connect() {
            if (isConnected) return;

            const token = document.getElementById('tokenInput').value;
            if (!token) {
                updateStatus('Enter the authentication token', 'disconnected');
                return;
            }

            updateStatus('Connecting...', 'connecting');
            updateButtons(true, false);
            try {
                const basePath = getBasePath();
                const response = await fetch(basePath + 'terminal-ticket', {
                    method: 'POST',
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                if (!response.ok) throw new Error('Terminal authorization failed');
                const { ticket } = await response.json();
                const wsUrl = 'wss://' + window.location.host + basePath + 'ws';
                socket = new WebSocket(wsUrl, ['webshell', ticket]);
            } catch (error) {
                updateStatus(error.message, 'disconnected');
                updateButtons(false, false);
                return;
            }

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

        // Update home link with base path
        function updateHomeLink() {
            const basePath = getBasePath();
            const link = document.getElementById('homeLink');
            if (link) {
                link.href = basePath;
            }
        }
        
        // Initialize when page loads
        window.addEventListener('load', function() {
            updateHomeLink();
            initTerminal();
        });
    </script>
</body>
</html>`

var terminalTemplate, terminalTemplateError = template.New("terminal").Parse(TerminalTemplate)

func GetTerminalTemplate() (*template.Template, error) {
	return terminalTemplate, terminalTemplateError
}
