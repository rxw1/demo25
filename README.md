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
- GraphQL Playground: `http://localhost:8080` (product-svc)

## 2) Migration + Seed

- `kubectl apply -f infra/helm/product-svc/templates/configmap-migrations.yaml`
- `kubectl apply -f infra/helm/product-svc/templates/configmap-seed.yaml`
- `kubectl apply -f infra/helm/product-svc/templates/job-migrate.yaml`

## 3) Kubernetes (k3d + Helm)

- `make k3d-up`
- `make helm-install`
- Status prüfen: `kubectl get pods -n app` und `kubectl get pods -n infra`
- Port-Forward Frontend (falls containerisiert): `kubectl port-forward svc/product-svc 8080:80 -n app`

## 4) Frontend über Helm deployen

- `helm upgrade --install frontend infra/helm/frontend -n app`

## 5) Feature Flag

- flagd ConfigMap editieren, um Redis zu toggeln: `kubectl edit configmap flagd-config -n app`

## 6) E2E Test

- `go test ./tests/e2e -v`

## 7) Aufräumen

- `make helm-uninstall && make k3d-down`
- `make dev-down`

Damit ist der PoC skeletonfähig mit Migration/Seed, Frontend Deployment und automatisiertem E2E-Test.

## Hinweise
- **gqlgen-Code:** einmal `go generate ./...` im `product-svc` ausführen (oder im Docker-Build). Eine `tools.go` mit `//go:build tools` kann `gqlgen` pinnen.
- **Sicherheit:** für den PoC sind Passwörter/URIs simpel gehalten. Bitte *nicht* in Produktion.
- **Observability:** absichtlich weggelassen; kann mit OTEL/Prom + Grafana später ergänzt werden.
```


