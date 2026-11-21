package terminal

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/adaptive-scale/webshell/internal/auth"
	"github.com/gorilla/websocket"
)

const (
	ticketTTL         = 30 * time.Second
	maxPendingTickets = 128
)

// Tickets authorizes a single browser connection without sending the API token
// in the WebSocket handshake. The API token remains usable in request headers.
type Tickets struct {
	mu      sync.Mutex
	pending map[string]time.Time
}

func NewTickets() *Tickets {
	return &Tickets{pending: make(map[string]time.Time)}
}

func sameOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) != 1 {
		return false
	}
	origin, err := url.Parse(origins[0])
	return err == nil && origin.Scheme == "https" && strings.EqualFold(origin.Host, r.Host) &&
		origin.User == nil && origin.Path == "" && origin.RawQuery == "" && origin.Fragment == ""
}

func (tickets *Tickets) Issue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !sameOrigin(r) {
		http.Error(w, "Forbidden origin", http.StatusForbidden)
		return
	}

	ticket := rand.Text()
	if !tickets.add(ticket, time.Now()) {
		http.Error(w, "Too many pending terminal tickets", http.StatusTooManyRequests)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]string{"ticket": ticket})
}

func (tickets *Tickets) add(ticket string, now time.Time) bool {
	tickets.mu.Lock()
	defer tickets.mu.Unlock()
	for key, expiry := range tickets.pending {
		if !now.Before(expiry) {
			delete(tickets.pending, key)
		}
	}
	if len(tickets.pending) >= maxPendingTickets {
		return false
	}
	tickets.pending[ticket] = now.Add(ticketTTL)
	return true
}

func (tickets *Tickets) authorize(r *http.Request) bool {
	if len(r.Header.Values("Origin")) == 0 {
		return auth.Authorized(r)
	}
	if !sameOrigin(r) || !websocket.IsWebSocketUpgrade(r) {
		return false
	}
	protocols := websocket.Subprotocols(r)
	if len(protocols) != 2 || protocols[0] != "webshell" {
		return false
	}

	tickets.mu.Lock()
	defer tickets.mu.Unlock()
	expiry, ok := tickets.pending[protocols[1]]
	delete(tickets.pending, protocols[1])
	return ok && time.Now().Before(expiry)
}
