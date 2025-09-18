package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type wsMsg struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func Test_Subscription_LastOrderCreated_Works(t *testing.T) {
	wsURL := getenv("GRAPHQL_WS_URL", "ws://localhost:8080/graphql")
	httpURL := getenv("GRAPHQL_URL", "http://localhost:8080/graphql")

	u, _ := url.Parse(wsURL)
	hdr := http.Header{}
	// Match same-host rule in server CheckOrigin
	hdr.Set("Origin", "http://"+u.Host)

	d := websocket.Dialer{Subprotocols: []string{"graphql-transport-ws"}}
	conn, resp, err := d.Dial(wsURL, hdr)
	if err != nil {
		if resp != nil {
			t.Fatalf("ws dial failed: %v (status=%s)", err, resp.Status)
		}
		t.Fatalf("ws dial failed: %v", err)
	}
	defer conn.Close()

	// Deadlines to avoid hanging tests
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// connection_init
	if err := conn.WriteJSON(wsMsg{Type: "connection_init"}); err != nil {
		t.Fatalf("write init: %v", err)
	}

	// expect connection_ack
	var ack wsMsg
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("read ack: %v", err)
	}
	if ack.Type != "connection_ack" {
		t.Fatalf("expected connection_ack, got %q", ack.Type)
	}

	// subscribe
	subQuery := `subscription { lastOrderCreated { id productId qty createdAt } }`
	subPayload := map[string]any{"query": subQuery}
	b, _ := json.Marshal(subPayload)
	if err := conn.WriteJSON(wsMsg{Type: "subscribe", ID: "1", Payload: b}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}

	// trigger mutation over HTTP
	mut := `mutation($pid:ID!,$qty:Int!){ createOrder(productId:$pid, qty:$qty){ id } }`
	body := map[string]any{"query": mut, "variables": map[string]any{"pid": "p1", "qty": 1}}
	bb, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, httpURL, bytes.NewReader(bb))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("mutation http: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("mutation status: %v", res.Status)
	}

	// expect next with our data
	deadline := time.Now().Add(10 * time.Second)
	var msg wsMsg
	for {
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for subscription event")
		}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read next: %v", err)
		}
		if msg.Type == "next" && msg.ID == "1" {
			break
		}
	}

	// parse payload
	var payload struct {
		Data struct {
			LastOrderCreated struct {
				ID        string
				ProductID string `json:"productId"`
				Qty       int
				CreatedAt string
			} `json:"lastOrderCreated"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Data.LastOrderCreated.ID == "" || payload.Data.LastOrderCreated.CreatedAt == "" {
		t.Fatalf("missing fields: %+v", payload.Data.LastOrderCreated)
	}
}
