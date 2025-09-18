package graphql

import (
	"context"
	"testing"
	"time"

	"rxw1/productsvc/internal/model"

	"github.com/99designs/gqlgen/client"
	gqlg "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
)

// fakeRoot implements ResolverRoot for testing using in-memory fixtures.
type fakeRoot struct {
	prods   []*model.Product
	orders  []*model.Order
	prices  map[string]int32
	lastSub chan *model.Order
}

// The generated code expects a type that can return the resolver interfaces.
func (f *fakeRoot) Mutation() MutationResolver         { return f }
func (f *fakeRoot) Query() QueryResolver               { return f }
func (f *fakeRoot) Subscription() SubscriptionResolver { return f }

// Query resolvers
func (f *fakeRoot) ProductByID(_ context.Context, productID string) (*model.Product, error) {
	for _, p := range f.prods {
		if p.ID == productID {
			return p, nil
		}
	}
	return nil, nil
}
func (f *fakeRoot) Products(_ context.Context) ([]*model.Product, error) { return f.prods, nil }
func (f *fakeRoot) Orders(_ context.Context) ([]*model.Order, error)     { return f.orders, nil }
func (f *fakeRoot) GetPrice(_ context.Context, productID string) (int32, error) {
	if v, ok := f.prices[productID]; ok {
		return v, nil
	}
	return 0, nil
}

// Mutation resolvers
func (f *fakeRoot) CreateOrder(_ context.Context, productID string, qty int32) (*model.Order, error) {
	o := &model.Order{
		ID:        "o-123",
		EventID:   "e-123",
		ProductID: productID,
		Qty:       qty,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	f.orders = append(f.orders, o)
	return o, nil
}

// Subscription resolvers (not exercised here)
func (f *fakeRoot) CurrentTime(context.Context) (<-chan *model.Time, error) { return nil, nil }

func (f *fakeRoot) LastOrderCreated(context.Context) (<-chan *model.Order, error) {
	// Provide a channel-backed subscription similar to gqlgen examples.
	if f.lastSub == nil {
		f.lastSub = make(chan *model.Order, 8)
	}
	return f.lastSub, nil
}

func (f *fakeRoot) OrdersByEvent(context.Context, string) (<-chan *model.Order, error) {
	return nil, nil
}

func (f *fakeRoot) OrdersByOrderID(context.Context, string) (<-chan *model.Order, error) {
	return nil, nil
}

func newTestClient(f *fakeRoot) *client.Client {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: f}))
	return client.New(srv)
}

// testClientFromSchema returns a gqlgen client for any ExecutableSchema.
func testClientFromSchema(es gqlg.ExecutableSchema) *client.Client {
	srv := handler.New(es)
	return client.New(srv)
}

// Var is a shorthand to pass variables to gqlgen/client.MustPost.
func Var(name string, v any) client.Option { return client.Var(name, v) }

func TestProductsAndProductByID(t *testing.T) {
	f := &fakeRoot{
		prods: []*model.Product{
			{ID: "p1", Name: "Widget", Price: 123},
			{ID: "p2", Name: "Gadget", Price: 456},
		},
	}
	c := newTestClient(f)

	// products query
	var resp struct {
		Products []struct {
			ID    string
			Name  string
			Price int
		}
	}
	c.MustPost(`query { products { id name price } }`, &resp)
	if got, want := len(resp.Products), 2; got != want {
		t.Fatalf("products len=%d want=%d", got, want)
	}
	if resp.Products[0].ID != "p1" || resp.Products[1].ID != "p2" {
		t.Fatalf("unexpected product ids: %+v", resp.Products)
	}

	// productById query
	var resp2 struct {
		ProductById *struct {
			ID   string
			Name string
		}
	}
	c.MustPost(`query($id: ID!) { productById(productId: $id) { id name } }`, &resp2, client.Var("id", "p2"))
	if resp2.ProductById == nil || resp2.ProductById.ID != "p2" {
		t.Fatalf("productById mismatch: %+v", resp2.ProductById)
	}
}

func TestCreateOrderMutationAndOrdersQuery(t *testing.T) {
	f := &fakeRoot{prods: []*model.Product{{ID: "p1", Name: "Widget", Price: 123}}}
	c := newTestClient(f)

	// createOrder mutation
	var mResp struct {
		CreateOrder struct {
			ID, ProductId, EventId, CreatedAt string
			Qty                               int
		}
	}
	c.MustPost(`mutation($pid: ID!, $qty: Int!){ createOrder(productId: $pid, qty: $qty){ id productId qty createdAt eventId } }`, &mResp,
		client.Var("pid", "p1"), client.Var("qty", 3))

	if mResp.CreateOrder.ProductId != "p1" || mResp.CreateOrder.Qty != 3 {
		t.Fatalf("createOrder mismatch: %+v", mResp.CreateOrder)
	}
	if mResp.CreateOrder.ID == "" || mResp.CreateOrder.EventId == "" || mResp.CreateOrder.CreatedAt == "" {
		t.Fatalf("fields should be populated: %+v", mResp.CreateOrder)
	}

	// orders query should include the created order
	var qResp struct {
		Orders []struct {
			ID, ProductId string
			Qty           int
		}
	}
	c.MustPost(`query { orders { id productId qty } }`, &qResp)
	if len(qResp.Orders) != 1 || qResp.Orders[0].ProductId != "p1" || qResp.Orders[0].Qty != 3 {
		t.Fatalf("orders mismatch: %+v", qResp.Orders)
	}
}

func TestGetPrice(t *testing.T) {
	f := &fakeRoot{prices: map[string]int32{"p1": 999}}
	c := newTestClient(f)

	var resp struct{ GetPrice int }
	c.MustPost(`query($id: ID!){ getPrice(productId: $id) }`, &resp, client.Var("id", "p1"))
	if resp.GetPrice != 999 {
		t.Fatalf("getPrice=%d want=999", resp.GetPrice)
	}
}

func TestSubscription_LastOrderCreated(t *testing.T) {
	// We test the subscription resolver behavior using a channel, like the Star Wars example.
	f := &fakeRoot{lastSub: make(chan *model.Order, 2)}

	// Subscribe (directly via resolver; client-level subscription is typically exercised in integration tests)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := f.LastOrderCreated(ctx)
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}

	// Publish two orders
	o1 := &model.Order{ID: "o1", ProductID: "p1", Qty: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	o2 := &model.Order{ID: "o2", ProductID: "p2", Qty: 2, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	f.lastSub <- o1
	f.lastSub <- o2

	// Receive them from subscription
	got1 := <-ch
	got2 := <-ch

	if got1.ID != "o1" || got1.ProductID != "p1" || got1.Qty != 1 {
		t.Fatalf("unexpected first order: %+v", got1)
	}
	if got2.ID != "o2" || got2.ProductID != "p2" || got2.Qty != 2 {
		t.Fatalf("unexpected second order: %+v", got2)
	}
}
