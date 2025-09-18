package graphql

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"rxw1/productsvc/internal/cache"
	"rxw1/productsvc/internal/flags"
	"rxw1/productsvc/internal/model"

	nats "github.com/nats-io/nats.go"
)

// fakeNATS implements natsx.Client
type fakeNATS struct {
	mu        sync.Mutex
	published []struct {
		subj string
		data []byte
	}
	reqReplies  map[string]func([]byte) ([]byte, error)
	subscribers map[string]nats.MsgHandler
}

func newFakeNATS() *fakeNATS {
	return &fakeNATS{
		reqReplies:  make(map[string]func([]byte) ([]byte, error)),
		subscribers: make(map[string]nats.MsgHandler),
	}
}

func (f *fakeNATS) Publish(subj string, data []byte) error {
	f.mu.Lock()
	f.published = append(f.published, struct {
		subj string
		data []byte
	}{subj, data})
	cb := f.subscribers[subj]
	f.mu.Unlock()
	if cb != nil {
		cb(&nats.Msg{Subject: subj, Data: data})
	}
	return nil
}

func (f *fakeNATS) Request(subj string, data []byte, _ time.Duration) (*nats.Msg, error) {
	f.mu.Lock()
	fn := f.reqReplies[subj]
	f.mu.Unlock()
	var out []byte
	var err error
	if fn != nil {
		out, err = fn(data)
	}
	return &nats.Msg{Subject: subj, Data: out}, err
}

func (f *fakeNATS) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	f.mu.Lock()
	f.subscribers[subj] = cb
	f.mu.Unlock()
	return nil, nil
}

func newResolverWithFakeNATS(fn *fakeNATS) *Resolver {
	// Only NC is needed for these tests; others can be nil or simple instances.
	return &Resolver{PG: nil, NC: fn, RC: &cache.Cache{}, FF: &flags.Flags{}}
}

func TestIntegration_CreateOrder_PublishesEvent(t *testing.T) {
	fn := newFakeNATS()
	res := newResolverWithFakeNATS(fn)
	svr := NewExecutableSchema(Config{Resolvers: res})
	c := testClientFromSchema(svr)

	var mResp struct {
		CreateOrder struct {
			ID, ProductId, EventId, CreatedAt string
			Qty                               int
		}
	}
	c.MustPost(`mutation($pid: ID!, $qty: Int!){ createOrder(productId: $pid, qty: $qty){ id productId qty createdAt eventId } }`, &mResp,
		Var("pid", "p1"), Var("qty", 2))

	if len(fn.published) == 0 || fn.published[0].subj != "order.created" {
		t.Fatalf("expected publish to order.created, got: %+v", fn.published)
	}
	// payload must contain productId and qty we sent
	var payload map[string]any
	if err := json.Unmarshal(fn.published[0].data, &payload); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if payload["productID"].(string) != "p1" || int(payload["qty"].(float64)) != 2 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestIntegration_Orders_Query_UsesNATSRequest(t *testing.T) {
	fn := newFakeNATS()
	// Reply to orders.all with two orders
	fn.reqReplies["orders.all"] = func(_ []byte) ([]byte, error) {
		rows := []*model.Order{{ID: "o1", ProductID: "p1", Qty: 1, CreatedAt: time.Now().Format(time.RFC3339)}, {ID: "o2", ProductID: "p2", Qty: 2, CreatedAt: time.Now().Format(time.RFC3339)}}
		return json.Marshal(rows)
	}
	res := newResolverWithFakeNATS(fn)
	svr := NewExecutableSchema(Config{Resolvers: res})
	c := testClientFromSchema(svr)

	var qResp struct {
		Orders []struct {
			ID, ProductId string
			Qty           int
		}
	}
	c.MustPost(`query { orders { id productId qty } }`, &qResp)
	if len(qResp.Orders) != 2 || qResp.Orders[0].ID != "o1" || qResp.Orders[1].ID != "o2" {
		t.Fatalf("unexpected orders: %+v", qResp.Orders)
	}
}

func TestIntegration_GetPrice_Query_UsesNATSRequest(t *testing.T) {
	fn := newFakeNATS()
	fn.reqReplies["products.price"] = func(in []byte) ([]byte, error) {
		// expect input is productID bytes
		return json.Marshal(777)
	}
	res := newResolverWithFakeNATS(fn)
	svr := NewExecutableSchema(Config{Resolvers: res})
	c := testClientFromSchema(svr)

	var resp struct{ GetPrice int }
	c.MustPost(`query($id: ID!){ getPrice(productId: $id) }`, &resp, Var("id", "pX"))
	if resp.GetPrice != 777 {
		t.Fatalf("price=%d want=777", resp.GetPrice)
	}
}

func TestIntegration_Subscription_LastOrderCreated_DeliversFromNATS(t *testing.T) {
	fn := newFakeNATS()
	res := newResolverWithFakeNATS(fn)
	// Use real subscription resolver method to get the channel wired to NATS subscribe
	ch, err := (&subscriptionResolver{res}).LastOrderCreated(context.Background())
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}

	// Publish an order through fake NATS; it should invoke the subscribed handler and deliver to channel
	o := &model.Order{ID: "o9", ProductID: "p9", Qty: 9, CreatedAt: time.Now().Format(time.RFC3339)}
	b, _ := json.Marshal(o)
	if err := fn.Publish("order.created", b); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case got := <-ch:
		if got.ID != "o9" || got.ProductID != "p9" || got.Qty != 9 {
			t.Fatalf("unexpected order: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscription delivery")
	}
}
