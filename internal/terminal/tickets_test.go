package terminal

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestExpiredTicket(t *testing.T) {
	tickets := NewTickets()
	if !tickets.add("expired-ticket-fixture", time.Now().Add(-ticketTTL)) {
		t.Fatal("could not add ticket fixture")
	}
	request := httptest.NewRequest("GET", "https://example.test/ws", nil)
	request.Header.Set("Origin", "https://example.test")
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Protocol", "webshell, expired-ticket-fixture")
	if tickets.authorize(request) {
		t.Fatal("expired ticket was accepted")
	}
}
