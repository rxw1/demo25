# Ultra-light PoC Skeleton (NATS + k3d, Redis-Flag, Compose-Dev)

> Monorepo-Skelett für Go (product/order), NATS, Postgres, Mongo, Redis (via Feature Flag), GraphQL (gqlgen), Next.js, Helm/k3d. **Fokus:** möglichst kurzer E2E-Durchstich + schneller Dev-Loop.

---

## Repo-Struktur
```
.
├─ Makefile
├─ README.md
├─ docker/
│  └─ compose.dev.yml
├─ graphql/
│  ├─ schema.graphql
│  └─ codegen.yml
├─ services/
│  ├─ productsvc/
│  │  ├─ go.mod
│  │  ├─ go.sum (generiert)
│  │  ├─ gqlgen.yml
│  │  ├─ graph/
│  │  │  ├─ model/models.go        # von gqlgen generiert (vereinfachte Vorlage unten)
│  │  │  ├─ schema.resolvers.go    # Handler, inkl. NATS Publish & Redis Flag
│  │  │  └─ generated.go           # generiert
│  │  ├─ internal/
│  │  │  ├─ db/pg.go
│  │  │  ├─ cache/redis.go
│  │  │  └─ flags/flags.go
│  │  └─ main.go
│  └─ ordersvc/
│     ├─ go.mod
│     ├─ go.sum (generiert)
│     ├─ internal/
│     │  ├─ mongo/store.go
│     │  └─ nats/subscriber.go
│     └─ main.go
├─ frontend/
│  ├─ package.json
│  ├─ next.config.js
│  ├─ app/
│  │  └─ page.tsx
│  └─ graphql/
│     ├─ schema.graphql  # optional, meist via remote/monorepo import
│     └─ codegen.ts
└─ infra/
   ├─ k3d/cluster.yaml
   └─ helm/
      ├─ productsvc/
      │  ├─ Chart.yaml
      │  ├─ values.yaml
      │  └─ templates/{deployment.yaml,service.yaml,configmap.env.yaml}
      ├─ ordersvc/
      │  ├─ Chart.yaml
      │  ├─ values.yaml
      │  └─ templates/{deployment.yaml,service.yaml}
      └─ flagd/
         ├─ Chart.yaml
         ├─ values.yaml
         └─ templates/{deployment.yaml,service.yaml,configmap-flags.yaml}
```

---

## Makefile (Kurz, dev & k8s)
```makefile
.PHONY: gen gql fe dev-up dev-down k3d-up k3d-down helm-deps helm-install helm-uninstall

gen: gql fe

gql:
	cd services/productsvc && go generate ./...

fe:
	cd frontend && npm run codegen || true

# --- Docker Compose Dev (schneller Dev-Loop) ---

dev-up:
	docker compose -f docker/compose.dev.yml up -d --build

dev-down:
	docker compose -f docker/compose.dev.yml down -v

# --- k3d + Helm ---

k3d-up:
	k3d cluster create poc -c infra/k3d/cluster.yaml

k3d-down:
	k3d cluster delete poc

helm-deps:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add nats https://nats-io.github.io/k8s/helm/charts/
	helm repo update

helm-install: helm-deps
	helm upgrade --install nats nats/nats -n infra --create-namespace
	helm upgrade --install pg bitnami/postgresql -n infra --set auth.postgresPassword=postgres
	helm upgrade --install mongo bitnami/mongodb -n infra
	helm upgrade --install redis bitnami/redis -n infra --set architecture=standalone
	helm upgrade --install flagd infra/helm/flagd -n app --create-namespace
	helm upgrade --install productsvc infra/helm/productsvc -n app
	helm upgrade --install ordersvc infra/helm/ordersvc -n app

helm-uninstall:
	helm uninstall productsvc ordersvc -n app || true
	helm uninstall flagd -n app || true
	helm uninstall nats pg mongo redis -n infra || true
```

---

## docker/compose.dev.yml (Dev-Loop)
```yaml
version: "3.9"
services:
  nats:
    image: nats:2.10-alpine
    ports: ["4222:4222", "8222:8222"]
  postgres:
    image: bitnami/postgresql:17
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRESQL_USERNAME=app
      - POSTGRESQL_PASSWORD=app
      - POSTGRESQL_DATABASE=app
    ports: ["5432:5432"]
  mongo:
    image: bitnami/mongodb:7.0
    ports: ["27017:27017"]
  redis:
    image: bitnami/redis:7
    environment:
      - ALLOW_EMPTY_PASSWORD=yes
    ports: ["6379:6379"]
  flagd:
    image: ghcr.io/open-feature/flagd:latest
    command: ["start", "--uri", "file:/etc/flagd/flags.json"]
    volumes:
      - ./infra/helm/flagd/templates/flags.json:/etc/flagd/flags.json:ro
    ports: ["8013:8013"]
  productsvc:
    build: ./services/productsvc
    environment:
      - DATABASE_URL=postgres://app:app@postgres:5432/app?sslmode=disable
      - REDIS_ADDR=redis:6379
      - NATS_URL=nats://nats:4222
      - FLAGD_HOST=flagd
      - FLAGD_PORT=8013
    depends_on: [postgres, redis, nats, flagd]
    ports: ["8080:8080"]
  ordersvc:
    build: ./services/ordersvc
    environment:
      - MONGO_URI=mongodb://mongo:27017
      - NATS_URL=nats://nats:4222
    depends_on: [mongo, nats]
    ports: ["8081:8081"]
```

---

## graphql/schema.graphql (klein & klar)
```graphql
type Product { id: ID!, name: String!, price: Int! }

type Order { id: ID!, productId: ID!, qty: Int!, createdAt: String! }

type Query {
  productById(id: ID!): Product
  orderById(id: ID!): Order
}

type Mutation {
  createOrder(productId: ID!, qty: Int!): Order
}
```

---

## services/productsvc/go.mod
```go
module productsvc

go 1.22

require (
	github.com/99designs/gqlgen v0.17.44
	github.com/go-chi/chi/v5 v5.0.12
	github.com/jackc/pgx/v5 v5.5.4
	github.com/nats-io/nats.go v1.41.0
	github.com/redis/go-redis/v9 v9.5.1
	github.com/open-feature/go-sdk v1.14.0
)
```

### services/productsvc/gqlgen.yml (minimal)
```yaml
schema:
  - ../../graphql/schema.graphql
generated: graph/generated.go
models:
  Product:
    model: productsvc/graph/model.Product
  Order:
    model: productsvc/graph/model.Order
resolver:
  layout: follow-schema
  dir: graph
  package: graph
```

### services/productsvc/graph/model/models.go (vereinfachte Vorlage)
```go
package model

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Order struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	Qty       int    `json:"qty"`
	CreatedAt string `json:"createdAt"`
}
```

### services/productsvc/internal/db/pg.go
```go
package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct{ Pool *pgxpool.Pool }

func Connect(ctx context.Context, url string) (*PG, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil { return nil, err }
	return &PG{Pool: pool}, nil
}

func (p *PG) GetProduct(ctx context.Context, id string) (string, string, int, error) {
	row := p.Pool.QueryRow(ctx, `select id, name, price from products where id=$1`, id)
	var pid, name string; var price int
	return pid, name, price, row.Scan(&pid, &name, &price)
}
```

### services/productsvc/internal/cache/redis.go
```go
package cache

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

type Cache struct{ R *redis.Client }

func New(addr string) *Cache { return &Cache{R: redis.NewClient(&redis.Options{Addr: addr})} }

func (c *Cache) Get(ctx context.Context, k string) (string, error) { return c.R.Get(ctx, k).Result() }
func (c *Cache) Set(ctx context.Context, k, v string, ttl time.Duration) error { return c.R.Set(ctx, k, v, ttl).Err() }
```

### services/productsvc/internal/flags/flags.go
```go
package flags

import (
	"context"
	of "github.com/open-feature/go-sdk/openfeature"
)

type Flags struct{ client of.Client }

func New() *Flags { return &Flags{client: of.NewClient("productsvc")} }

func (f *Flags) RedisEnabled(ctx context.Context) bool {
	val, err := f.client.BooleanValue(ctx, "redisCacheEnabled", false, of.EvaluationOptions{})
	return err == nil && val
}
```

### services/productsvc/graph/schema.resolvers.go (Kernlogik)
```go
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"productsvc/graph/model"
	"productsvc/internal/cache"
	"productsvc/internal/db"
	"productsvc/internal/flags"

	"github.com/nats-io/nats.go"
)

type Resolver struct {
	PG *db.PG
	NC *nats.Conn
	RC *cache.Cache
	FF *flags.Flags
}

func (r *Resolver) ProductByID(ctx context.Context, id string) (*model.Product, error) {
	useCache := r.FF.RedisEnabled(ctx)
	if useCache {
		if s, err := r.RC.Get(ctx, "product:"+id); err == nil {
			var p model.Product; if json.Unmarshal([]byte(s), &p) == nil { return &p, nil }
		}
	}
	pid, name, price, err := r.PG.GetProduct(ctx, id)
	if err != nil { return nil, err }
	p := &model.Product{ID: pid, Name: name, Price: price}
	if useCache { b,_ := json.Marshal(p); _ = r.RC.Set(ctx, "product:"+id, string(b), 5*time.Minute) }
	return p, nil
}

func (r *Resolver) CreateOrder(ctx context.Context, productId string, qty int) (*model.Order, error) {
	// Publish event to NATS; ordersvc materialisiert.
	event := map[string]any{"id": fmt.Sprintf("evt-%d", time.Now().UnixNano()), "productId": productId, "qty": qty, "createdAt": time.Now().UTC().Format(time.RFC3339)}
	b, _ := json.Marshal(event)
	if err := r.NC.Publish("orders.created", b); err != nil { return nil, err }
	// Für Demo: return sync Confirmation
	return &model.Order{ID: event["id"].(string), ProductID: productId, Qty: qty, CreatedAt: event["createdAt"].(string)}, nil
}
```

### services/productsvc/main.go (Server + Wiring)
```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"productsvc/graph"
	"productsvc/internal/cache"
	"productsvc/internal/db"
	"productsvc/internal/flags"
)

func main() {
	ctx := context.Background()
	pg, err := db.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil { log.Fatal(err) }
	
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil { log.Fatal(err) }
	
	rc := cache.New(os.Getenv("REDIS_ADDR"))
	_ = redis.NewClient // keep import
	ff := flags.New()

	res := &graph.Resolver{PG: pg, NC: nc, RC: rc, FF: ff}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: res}))

	r := chi.NewRouter()
	r.Handle("/", playground.Handler("GraphQL", "/graphql"))
	r.Handle("/graphql", srv)

	log.Println("productsvc up on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
```

> **Hinweis:** `generated.go` & `NewExecutableSchema` werden via `go generate` von gqlgen erzeugt.

---

## services/ordersvc/go.mod
```go
module ordersvc

go 1.22

require (
	github.com/nats-io/nats.go v1.41.0
	go.mongodb.org/mongo-driver v1.15.0
	github.com/go-chi/chi/v5 v5.0.12
)
```

### services/ordersvc/internal/mongo/store.go
```go
package mongo

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct{ C *mongo.Collection }

func Connect(ctx context.Context, uri string) (*Store, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil { return nil, err }
	return &Store{C: cli.Database("app").Collection("orders")}, nil
}

func (s *Store) UpsertOrder(ctx context.Context, evtID, productID string, qty int, createdAt time.Time) error {
	_, err := s.C.UpdateOne(ctx, bson.M{"eventId": evtID}, bson.M{"$setOnInsert": bson.M{
		"eventId": evtID, "productId": productID, "qty": qty, "createdAt": createdAt,
	}}, options.Update().SetUpsert(true))
	return err
}
```

### services/ordersvc/internal/nats/subscriber.go
```go
package nats

import (
	"context"
	"encoding/json"
	"time"

	"ordersvc/internal/mongo"
	"github.com/nats-io/nats.go"
)

type Event struct { ID, ProductID string; Qty int; CreatedAt string }

func Start(ctx context.Context, nc *nats.Conn, store *mongo.Store) error {
	_, err := nc.Subscribe("orders.created", func(m *nats.Msg) {
		var e Event; if json.Unmarshal(m.Data, &e) != nil { return }
		ts, _ := time.Parse(time.RFC3339, e.CreatedAt)
		_ = store.UpsertOrder(ctx, e.ID, e.ProductID, e.Qty, ts)
	})
	return err
}
```

### services/ordersvc/main.go
```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"ordersvc/internal/mongo"
	nsub "ordersvc/internal/nats"
)

func main() {
	ctx := context.Background()
	store, err := mongo.Connect(ctx, os.Getenv("MONGO_URI"))
	if err != nil { log.Fatal(err) }
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil { log.Fatal(err) }
	if err := nsub.Start(ctx, nc, store); err != nil { log.Fatal(err) }

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	log.Println("ordersvc up on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
```

---

## frontend/package.json (Next.js + Apollo)
```json
{
  "name": "poc-frontend",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "codegen": "graphql-codegen --config ./graphql/codegen.ts"
  },
  "dependencies": {
    "next": "14.2.5",
    "react": "18.2.0",
    "react-dom": "18.2.0",
    "@apollo/client": "3.10.0",
    "graphql": "16.9.0"
  },
  "devDependencies": {
    "@graphql-codegen/cli": "5.0.2",
    "@graphql-codegen/client-preset": "4.4.0",
    "typescript": "5.4.5"
  }
}
```

### frontend/graphql/codegen.ts
```ts
import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: 'http://localhost:8080/graphql',
  documents: ['./app/**/*.{ts,tsx}'],
  generates: {
    './app/__generated__/': {
      preset: 'client',
    },
  },
}
export default config
```

### frontend/app/page.tsx (sehr knapp)
```tsx
'use client'
import { ApolloClient, InMemoryCache, ApolloProvider, gql, useQuery, useMutation } from '@apollo/client'

const client = new ApolloClient({ uri: 'http://localhost:8080/graphql', cache: new InMemoryCache() })

const Q = gql`query($id: ID!){ productById(id:$id){ id name price } }`
const M = gql`mutation($productId:ID!,$qty:Int!){ createOrder(productId:$productId, qty:$qty){ id productId qty createdAt } }`

function PageInner(){
  const { data } = useQuery(Q, { variables: { id: 'p1' } })
  const [createOrder] = useMutation(M)
  return (
    <div style={{ padding: 24 }}>
      <h1>PoC</h1>
      <pre>{JSON.stringify(data?.productById, null, 2)}</pre>
      <button onClick={()=>createOrder({ variables: { productId:'p1', qty:1 } })}>Create Order</button>
    </div>
  )
}
export default function Page(){ return <ApolloProvider client={client}><PageInner/></ApolloProvider> }
```

---

## infra/k3d/cluster.yaml (einfach)
```yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: poc
servers: 1
agents: 1
ports:
  - port: 3000:3000
    nodeFilters: [loadbalancer]
```

---

## infra/helm/productsvc/Chart.yaml
```yaml
apiVersion: v2
name: productsvc
version: 0.1.0
appVersion: "0.1.0"
```

### infra/helm/productsvc/values.yaml
```yaml
image:
  repository: productsvc
  tag: latest
  pullPolicy: IfNotPresent

env:
  DATABASE_URL: postgres://app:app@pg-postgresql.infra.svc.cluster.local:5432/app?sslmode=disable
  REDIS_ADDR: redis-master.infra.svc.cluster.local:6379
  NATS_URL: nats://nats.infra.svc.cluster.local:4222
  FLAGD_HOST: flagd.app.svc.cluster.local
  FLAGD_PORT: "8013"
service:
  port: 80
  targetPort: 8080
```

### infra/helm/productsvc/templates/configmap.env.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: productsvc-env
  labels: { app: productsvc }
data:
{{- range $k, $v := .Values.env }}
  {{ $k }}: "{{ $v }}"
{{- end }}
```

### infra/helm/productsvc/templates/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: productsvc
spec:
  replicas: 1
  selector: { matchLabels: { app: productsvc } }
  template:
    metadata:
      labels: { app: productsvc }
    spec:
      containers:
        - name: productsvc
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          envFrom:
            - configMapRef: { name: productsvc-env }
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet: { path: /, port: 8080 }
          livenessProbe:
            httpGet: { path: /, port: 8080 }
```

### infra/helm/productsvc/templates/service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: productsvc
spec:
  selector: { app: productsvc }
  ports:
    - port: {{ .Values.service.port }}
      targetPort: {{ .Values.service.targetPort }}
```

---

## infra/helm/ordersvc/Chart.yaml
```yaml
apiVersion: v2
name: ordersvc
version: 0.1.0
appVersion: "0.1.0"
```

### infra/helm/ordersvc/values.yaml
```yaml
image:
  repository: ordersvc
  tag: latest
  pullPolicy: IfNotPresent

env:
  MONGO_URI: mongodb://mongo-mongodb.infra.svc.cluster.local:27017
  NATS_URL: nats://nats.infra.svc.cluster.local:4222
service:
  port: 80
  targetPort: 8081
```

### infra/helm/ordersvc/templates/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ordersvc
spec:
  replicas: 1
  selector: { matchLabels: { app: ordersvc } }
  template:
    metadata:
      labels: { app: ordersvc }
    spec:
      containers:
        - name: ordersvc
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          envFrom:
            - configMapRef: { name: ordersvc-env }
          ports:
            - containerPort: 8081
          readinessProbe:
            httpGet: { path: /healthz, port: 8081 }
          livenessProbe:
            httpGet: { path: /healthz, port: 8081 }
---
apiVersion: v1
kind: ConfigMap
metadata: { name: ordersvc-env }
data:
{{- range $k, $v := .Values.env }}
  {{ $k }}: "{{ $v }}"
{{- end }}
```

### infra/helm/ordersvc/templates/service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: ordersvc
spec:
  selector: { app: ordersvc }
  ports:
    - port: {{ .Values.service.port }}
      targetPort: {{ .Values.service.targetPort }}
```

---

## infra/helm/flagd/Chart.yaml
```yaml
apiVersion: v2
name: flagd
version: 0.1.0
```

### infra/helm/flagd/values.yaml
```yaml
image: ghcr.io/open-feature/flagd:latest
service:
  port: 8013
```

### infra/helm/flagd/templates/configmap-flags.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: flagd-config
  labels: { app: flagd }
data:
  flags.json: |
    {
      "$schema": "https://flagd.dev/schema/v0/flags.json",
      "flags": {
        "redisCacheEnabled": { "state": "ENABLED", "defaultVariant": "on", "variants": { "on": true, "off": false } }
      }
    }
```

### infra/helm/flagd/templates/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flagd
spec:
  replicas: 1
  selector: { matchLabels: { app: flagd } }
  template:
    metadata:
      labels: { app: flagd }
    spec:
      containers:
        - name: flagd
          image: {{ .Values.image }}
          args: ["start", "--uri", "kubernetes://flagd-config/flags.json"]
          ports: [{ containerPort: 8013 }]
```

### infra/helm/flagd/templates/service.yaml
```yaml
apiVersion: v1
kind: Service
metadata: { name: flagd }
spec:
  selector: { app: flagd }
  ports:
    - port: {{ .Values.service.port }}
      targetPort: 8013
```

---

## README.md (Kurzablauf)
```md
# PoC Quickstart

## 0) Vorbereitungen
- Go 1.22, Node 18+, Docker, Docker Compose, k3d, Helm

## 1) Dev-Loop (Compose)
- `make dev-up`
- Seed (einmalig):
  - `docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "create table if not exists products(id text primary key, name text, price int);"`
  - `docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing;"`
- Services lokal starten (optional) oder via Compose-Images.
- Frontend: `cd frontend && npm install && npm run dev` (öffnet `http://localhost:3000`)
- GraphQL Playground: `http://localhost:8080` (productsvc)

## 2) Kubernetes (k3d + Helm)
- `make k3d-up`
- `make helm-install`
- Status prüfen: `kubectl get pods -n app` und `kubectl get pods -n infra`
- Port-Forward Frontend (falls containerisiert): `kubectl port-forward svc/productsvc 8080:80 -n app`

## 3) Feature Flag
- flagd ConfigMap editieren, um Redis zu toggeln: `kubectl edit configmap flagd-config -n app`

## 4) Räumen
- `make helm-uninstall && make k3d-down`
- `make dev-down`
```

---

## .gitlab-ci.yml (Minimal, ohne Deploy)
```yaml
stages: [lint, test, build]

golang-lint:
  stage: lint
  image: golang:1.22
  script:
    - cd services/productsvc && go vet ./...
    - cd ../ordersvc && go vet ./...

go-test:
  stage: test
  image: golang:1.22
  script:
    - cd services/productsvc && go test ./... || true
    - cd ../ordersvc && go test ./... || true

docker-build-product:
  stage: build
  image: gcr.io/kaniko-project/executor:latest
  script:
    - /kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile services/productsvc/Dockerfile --destination $CI_REGISTRY_IMAGE/productsvc:${CI_COMMIT_SHORT_SHA}

docker-build-order:
  stage: build
  image: gcr.io/kaniko-project/executor:latest
  script:
    - /kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile services/ordersvc/Dockerfile --destination $CI_REGISTRY_IMAGE/ordersvc:${CI_COMMIT_SHORT_SHA}
```

---

## Dockerfiles (Beispiel: productsvc & ordersvc)

**services/productsvc/Dockerfile**
```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY services/productsvc/go.mod services/productsvc/go.sum ./
RUN go mod download
COPY services/productsvc/. .
RUN go generate ./... && CGO_ENABLED=0 go build -o /out/productsvc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/productsvc /productsvc
EXPOSE 8080
ENTRYPOINT ["/productsvc"]
```

**services/ordersvc/Dockerfile**
```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY services/ordersvc/go.mod services/ordersvc/go.sum ./
RUN go mod download
COPY services/ordersvc/. .
RUN CGO_ENABLED=0 go build -o /out/ordersvc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/ordersvc /ordersvc
EXPOSE 8081
ENTRYPOINT ["/ordersvc"]
```

---

## Hinweise
- **gqlgen-Code:** einmal `go generate ./...` im `productsvc` ausführen (oder im Docker-Build). Eine `tools.go` mit `//go:build tools` kann `gqlgen` pinnen.
- **Sicherheit:** für den PoC sind Passwörter/URIs simpel gehalten. Bitte *nicht* in Produktion.
- **Observability:** absichtlich weggelassen; kann mit OTEL/Prom + Grafana später ergänzt werden.
```

