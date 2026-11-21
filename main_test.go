package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adaptive-scale/webshell/internal/config"
	"github.com/gorilla/websocket"
)

func TestFixedTerminalEntry(t *testing.T) {
	const token = "existing-long-term-test-token"
	config.SetAuthToken(token)
	t.Cleanup(func() { config.SetAuthToken("") })

	mux := http.NewServeMux()
	setupRoutes(mux, "/managed/")
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	client := server.Client()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/managed/api/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	status, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	status.Body.Close()
	if status.StatusCode != http.StatusOK {
		t.Fatalf("Chumen status API declined the long-term token: %s", status.Status)
	}

	const terminalPath = "/managed/terminal"
	response, err := client.Get(server.URL + terminalPath)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("fixed terminal page returned %s", response.Status)
	}

	for _, origin := range []string{"https://invalid.example", server.URL} {
		request, err := http.NewRequest(http.MethodPost, server.URL+"/managed/terminal-ticket", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Origin", origin)
		request.Header.Set("X-Auth-Token", token)
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		if origin != server.URL {
			response.Body.Close()
			if response.StatusCode != http.StatusForbidden {
				t.Fatalf("unexpected status for another origin: %s", response.Status)
			}
			continue
		}

		var issued struct{ Ticket string }
		decodeErr := json.NewDecoder(response.Body).Decode(&issued)
		response.Body.Close()
		if response.StatusCode != http.StatusOK || decodeErr != nil || issued.Ticket == "" {
			t.Fatalf("long-term token did not obtain a ticket: %s, %v", response.Status, decodeErr)
		}

		dialer := websocket.Dialer{
			TLSClientConfig: client.Transport.(*http.Transport).TLSClientConfig,
			Subprotocols:    []string{"webshell", issued.Ticket},
		}
		wsURL := "wss://" + strings.TrimPrefix(server.URL, "https://") + "/managed/ws"
		headers := http.Header{"Origin": []string{server.URL}}
		conn, _, err := dialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
		_, failed, err := dialer.Dial(wsURL, headers)
		if err == nil || failed == nil || failed.StatusCode != http.StatusUnauthorized {
			t.Fatal("the used ticket must be rejected")
		}
		failed.Body.Close()
	}
}
