# Ultra-light PoC Skeleton (NATS + k3d, Redis-Flag, Compose-Dev)

> Monorepo-Skelett für Go (product/order), NATS, Postgres, Mongo, Redis (via Feature Flag), GraphQL (gqlgen), Next.js, Helm/k3d. **Fokus:** kurzer E2E-Durchstich + schneller Dev-Loop. **Jetzt erweitert um:**
> - Datenbank-Migration & Seed
> - Frontend Dockerfile + Helm-Chart
> - Mini E2E-Test (Go)

---

## Neue Teile

### services/product-svc/migrations/001_init.sql
```sql
create table if not exists products (
  id text primary key,
  name text not null,
  price int not null
);
```

### services/product-svc/seed/seed.go
```go
package main

import (
  "context"
  "log"
  "os"

  "github.com/jackc/pgx/v5/pgxpool"
)

func main(){
  url := os.Getenv("DATABASE_URL")
  ctx := context.Background()
  pool, err := pgxpool.New(ctx, url)
  if err != nil { log.Fatal(err) }
  _, err = pool.Exec(ctx, `insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing`)
  if err != nil { log.Fatal(err) }
  log.Println("Seed done")
}
```

### infra/helm/product-svc/templates/job-migrate.yaml
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: product-svc-migrate
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: migrate
          image: product-svc:latest
          command: ["sh","-c"]
          args:
            - >
              psql $DATABASE_URL -f /migrations/001_init.sql &&
              go run /seed/seed.go
          envFrom:
            - configMapRef: { name: product-svc-env }
          volumeMounts:
            - name: migrations
              mountPath: /migrations
            - name: seed
              mountPath: /seed
      volumes:
        - name: migrations
          configMap:
            name: product-svc-migrations
        - name: seed
          configMap:
            name: product-svc-seed
```

### infra/helm/product-svc/templates/configmap-migrations.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: product-svc-migrations
data:
  001_init.sql: |
    create table if not exists products (
      id text primary key,
      name text not null,
      price int not null
    );
```

### infra/helm/product-svc/templates/configmap-seed.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: product-svc-seed
data:
  seed.go: |
    package main
    import ("context";"log";"os";"github.com/jackc/pgx/v5/pgxpool")
    func main(){
      url := os.Getenv("DATABASE_URL")
      ctx := context.Background()
      pool, err := pgxpool.New(ctx, url)
      if err != nil { log.Fatal(err) }
      _, err = pool.Exec(ctx, `insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing`)
      if err != nil { log.Fatal(err) }
      log.Println("Seed done")
    }
```

---

### frontend/Dockerfile
```dockerfile
# syntax=docker/dockerfile:1
FROM node:18-alpine AS build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/. .
RUN npm run build

FROM node:18-alpine
WORKDIR /app
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./
EXPOSE 3000
CMD ["npm","start"]
```

### infra/helm/frontend/Chart.yaml
```yaml
apiVersion: v2
name: frontend
version: 0.1.0
```

### infra/helm/frontend/values.yaml
```yaml
image:
  repository: frontend
  tag: latest
  pullPolicy: IfNotPresent
service:
  port: 80
  targetPort: 3000
```

### infra/helm/frontend/templates/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 1
  selector: { matchLabels: { app: frontend } }
  template:
    metadata:
      labels: { app: frontend }
    spec:
      containers:
        - name: frontend
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          ports:
            - containerPort: 3000
```

### infra/helm/frontend/templates/service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
spec:
  selector: { app: frontend }
  ports:
    - port: {{ .Values.service.port }}
      targetPort: {{ .Values.service.targetPort }}
```

---

### tests/e2e/e2e_test.go
```go
package e2e

import (
  "bytes"
  "encoding/json"
  "net/http"
  "testing"
)

func TestCreateOrderFlow(t *testing.T){
  query := `mutation($productId:ID!,$qty:Int!){ createOrder(productId:$productId, qty:$qty){ id productId qty } }`
  payload := map[string]any{"query": query, "variables": map[string]any{"productId": "p1", "qty": 1}}
  b,_ := json.Marshal(payload)
  resp, err := http.Post("http://localhost:8080/graphql", "application/json", bytes.NewReader(b))
  if err != nil { t.Fatal(err) }
  if resp.StatusCode != 200 { t.Fatalf("unexpected status: %v", resp.Status) }
}
```

---

## README.md (ergänzt)
```md
### 1a) Migration + Seed
- `kubectl apply -f infra/helm/product-svc/templates/configmap-migrations.yaml`
- `kubectl apply -f infra/helm/product-svc/templates/configmap-seed.yaml`
- `kubectl apply -f infra/helm/product-svc/templates/job-migrate.yaml`

### 2a) Frontend über Helm deployen
- `helm upgrade --install frontend infra/helm/frontend -n app`

### 3a) E2E Test
- `go test ./tests/e2e -v`
```

---

Damit ist der PoC skeletonfähig mit Migration/Seed, Frontend Deployment und automatisiertem E2E-Test.



---

# Ergänzungen: Migrations & Seed-Jobs (k8s), Frontend-Container/Helm, E2E-Test

## 1) Mini-Migrations & Seeds als Kubernetes Jobs (Helm Hooks)

**Ziel:** Kein externes Tooling nötig. Wir lassen Helm zwei Jobs ausführen:
- **Migration** (idempotent): erzeugt Tabelle `products` (falls fehlt)
- **Seed** (idempotent): fügt `p1` ein (falls fehlt)
- **Mongo Index** (idempotent): Unique-Index auf `eventId`

### Neue Dateien
```
infra/helm/infra-jobs/
  Chart.yaml
  values.yaml
  templates/
    configmap-pg.sql.yaml     # SQL für Migration & Seeds
    job-pg-migrate.yaml       # führt Migration aus (psql)
    job-pg-seed.yaml          # führt Seed aus (psql)
    job-mongo-index.yaml      # legt Unique-Index an (mongosh)
```

### `infra/helm/infra-jobs/Chart.yaml`
```yaml
apiVersion: v2
name: infra-jobs
version: 0.1.0
```

### `infra/helm/infra-jobs/values.yaml`
```yaml
postgres:
  host: pg-postgresql.infra.svc.cluster.local
  user: app
  password: app
  db: app
mongo:
  uri: mongodb://mongo-mongodb.infra.svc.cluster.local:27017/app
```

### `infra/helm/infra-jobs/templates/configmap-pg.sql.yaml`
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pg-init-sql
  labels: { app: infra-jobs }
data:
  migrate.sql: |
    create table if not exists products (
      id text primary key,
      name text not null,
      price int not null
    );
  seed.sql: |
    insert into products(id,name,price)
    values('p1','Widget',199)
    on conflict (id) do nothing;
```

### `infra/helm/infra-jobs/templates/job-pg-migrate.yaml`
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pg-migrate
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: psql
          image: bitnami/postgresql:17
          command: ["/bin/sh","-c"]
          args:
            - >-
              PGPASSWORD=${PGPASSWORD} psql -h ${PGHOST} -U ${PGUSER} -d ${PGDATABASE} -f /sql/migrate.sql;
          env:
            - name: PGHOST
              value: {{ .Values.postgres.host | quote }}
            - name: PGUSER
              value: {{ .Values.postgres.user | quote }}
            - name: PGPASSWORD
              value: {{ .Values.postgres.password | quote }}
            - name: PGDATABASE
              value: {{ .Values.postgres.db | quote }}
          volumeMounts:
            - name: sql
              mountPath: /sql
      volumes:
        - name: sql
          configMap:
            name: pg-init-sql
            items: [{ key: migrate.sql, path: migrate.sql }]
```

### `infra/helm/infra-jobs/templates/job-pg-seed.yaml`
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pg-seed
  annotations:
    helm.sh/hook: post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: psql
          image: bitnami/postgresql:17
          command: ["/bin/sh","-c"]
          args:
            - >-
              PGPASSWORD=${PGPASSWORD} psql -h ${PGHOST} -U ${PGUSER} -d ${PGDATABASE} -f /sql/seed.sql;
          env:
            - name: PGHOST
              value: {{ .Values.postgres.host | quote }}
            - name: PGUSER
              value: {{ .Values.postgres.user | quote }}
            - name: PGPASSWORD
              value: {{ .Values.postgres.password | quote }}
            - name: PGDATABASE
              value: {{ .Values.postgres.db | quote }}
          volumeMounts:
            - name: sql
              mountPath: /sql
      volumes:
        - name: sql
          configMap:
            name: pg-init-sql
            items: [{ key: seed.sql, path: seed.sql }]
```

### `infra/helm/infra-jobs/templates/job-mongo-index.yaml`
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: mongo-index
  annotations:
    helm.sh/hook: post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: mongosh
          image: bitnami/mongodb:7.0
          command: ["/bin/bash","-c"]
          args:
            - >-
              mongosh "${MONGO_URI}" --eval 'db.orders.createIndex({eventId:1},{unique:true})';
          env:
            - name: MONGO_URI
              value: {{ .Values.mongo.uri | quote }}
```

**Install-Reihenfolge** (angepasst):
- `helm upgrade --install infra-jobs infra/helm/infra-jobs -n infra` **zwischen** DB-Install und App-Deploy.

> Ergänze im `Makefile` unter `helm-install` direkt nach den DB-Charts: `helm upgrade --install infra-jobs infra/helm/infra-jobs -n infra`.

---

## 2) Frontend: Dockerfile & Helm-Chart

### `frontend/Dockerfile`
```dockerfile
# syntax=docker/dockerfile:1
FROM node:18-alpine AS deps
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

FROM node:18-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/. .
RUN npm run build

FROM node:18-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.next ./.next
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./package.json
EXPOSE 3000
CMD ["npm","start"]
```

### Helm-Chart für das Frontend
```
infra/helm/frontend/
  Chart.yaml
  values.yaml
  templates/{deployment.yaml,service.yaml}
```

**`infra/helm/frontend/Chart.yaml`**
```yaml
apiVersion: v2
name: frontend
version: 0.1.0
appVersion: "0.1.0"
```

**`infra/helm/frontend/values.yaml`**
```yaml
image:
  repository: frontend
  tag: latest
  pullPolicy: IfNotPresent
service:
  port: 80
  targetPort: 3000
  type: ClusterIP
env:
  GRAPHQL_URL: http://product-svc.app.svc.cluster.local/graphql
```

**`infra/helm/frontend/templates/deployment.yaml`**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 1
  selector: { matchLabels: { app: frontend } }
  template:
    metadata:
      labels: { app: frontend }
    spec:
      containers:
        - name: frontend
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          env:
            - name: NEXT_PUBLIC_GRAPHQL_URL
              value: {{ .Values.env.GRAPHQL_URL | quote }}
          ports:
            - containerPort: 3000
          readinessProbe:
            httpGet: { path: /, port: 3000 }
          livenessProbe:
            httpGet: { path: /, port: 3000 }
```

**`infra/helm/frontend/templates/service.yaml`**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
spec:
  selector: { app: frontend }
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: {{ .Values.service.targetPort }}
```

**Frontend kleine Änderung:** `frontend/app/page.tsx` → GraphQL-URL über `process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:8080/graphql'` beziehen.

**Makefile-Ergänzung (`helm-install`):**
```
helm upgrade --install frontend infra/helm/frontend -n app
```

---

## 3) Kleiner E2E-Test (Go)

**Ziel:** Gegen laufendes System testen (Compose **oder** k3d). Test ruft GraphQL-Mutation auf und verifiziert, dass ein Order-Dokument in Mongo landet.

### Struktur
```
tests/e2e/
  go.mod
  e2e_test.go
```

### `tests/e2e/go.mod`
```go
module e2e

go 1.22

require (
	github.com/machinebox/graphql v0.2.2
	go.mongodb.org/mongo-driver v1.15.0
)
```

### `tests/e2e/e2e_test.go`
```go
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
	var resp struct{ CreateOrder struct{ ID, ProductID, CreatedAt string; Qty int } }
	if err := client.Run(ctx, req, &resp); err != nil { t.Fatalf("graphql mutation failed: %v", err) }
	if resp.CreateOrder.ID == "" { t.Fatalf("expected order id") }

	// 2) Auf Mongo warten (einfaches Polling)
	mcli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil { t.Fatalf("mongo connect: %v", err) }
	col := mcli.Database("app").Collection("orders")

	deadline := time.Now().Add(10 * time.Second)
	for {
		var doc bson.M
		err := col.FindOne(ctx, bson.M{"eventId": resp.CreateOrder.ID}).Decode(&doc)
		if err == nil { break }
		if time.Now().After(deadline) { t.Fatalf("order not materialized in mongo in time: %v", err) }
		time.Sleep(500 * time.Millisecond)
	}
}

func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
```

**Ausführung:**
- **Compose:** `GRAPHQL_URL=http://localhost:8080/graphql MONGO_URI=mongodb://localhost:27017 go test ./tests/e2e -v`
- **k3d:** Port-forward GraphQL & Mongo oder benutze die Cluster-Services und führe den Test in einem Pod/Runner aus.

---

## Recap der Makefile-Erweiterungen
- In `helm-install` **nach** den DB-Charts:
```
helm upgrade --install infra-jobs infra/helm/infra-jobs -n infra
helm upgrade --install frontend infra/helm/frontend -n app
```

Damit sind Migration/Seed automatisiert, das Frontend containerisiert & deploybar, und du hast einen kleinen E2E-Gurt. ✅

