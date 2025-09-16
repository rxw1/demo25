package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/machinebox/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Test_CreateOrder_MaterializesInMongo(t *testing.T) {
	graphqlURL := getenv("GRAPHQL_URL", "http://localhost:8080/graphql")
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")

	ctx := context.Background()
	client := graphql.NewClient(graphqlURL)

	// 1) Mutation ausführen
	req := graphql.NewRequest(`mutation($pid:ID!,$qty:Int!){ createOrder(productId:$pid, qty:$qty){ id productId qty createdAt } }`)
	req.Var("pid", "p1")
	req.Var("qty", 1)
	var resp struct {
		CreateOrder struct {
			ID, ProductID, CreatedAt string
			Qty                      int
		}
	}
	if err := client.Run(ctx, req, &resp); err != nil {
		t.Fatalf("graphql mutation failed: %v", err)
	}
	if resp.CreateOrder.ID == "" {
		t.Fatalf("expected order id")
	}

	// 2) Auf Mongo warten (einfaches Polling)
	mcli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("mongo connect: %v", err)
	}
	col := mcli.Database("app").Collection("orders")

	deadline := time.Now().Add(10 * time.Second)
	for {
		var doc bson.M
		err := col.FindOne(ctx, bson.M{"eventId": resp.CreateOrder.ID}).Decode(&doc)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("order not materialized in mongo in time: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// package e2e

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"testing"
// )

// func TestCreateOrderFlow(t *testing.T) {
// 	query := `mutation($productId:ID!,$qty:Int!){ createOrder(productId:$productId, qty:$qty){ id productId qty } }`
// 	payload := map[string]any{"query": query, "variables": map[string]any{"productId": "p1", "qty": 1}}
// 	b, _ := json.Marshal(payload)
// 	resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(b))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if resp.StatusCode != 200 {
// 		t.Fatalf("unexpected status: %v", resp.Status)
// 	}
// }
