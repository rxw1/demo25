# Quickstart

## Requirements

- Go 1.25, Node 18+, Docker, Docker Compose, k3d, Helm

## 1) Dev-Loop (using Docker Compose)

- `make dev-up`

```sh
docker compose -f docker/compose.dev.yml up -d --build
```

- Services lokal starten (optional) oder via Compose-Images.

- Frontend: `cd frontend && npm install && npm run dev` (opens `http://localhost:3000`)
- GraphQL Playground: `http://localhost:8080` (product-svc)

## 2) Migration + Seed

- `make seed` (once)
- `make migration_seed`

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

- `make test_e2e`

## 7) Aufräumen

- `make clean`

Damit ist der PoC skeletonfähig mit Migration/Seed, Frontend Deployment und automatisiertem E2E-Test.

## Hinweise
- **gqlgen-Code:** einmal `go generate ./...` im `product-svc` ausführen (oder im Docker-Build). Eine `tools.go` mit `//go:build tools` kann `gqlgen` pinnen.
- **Sicherheit:** für den PoC sind Passwörter/URIs simpel gehalten. Bitte *nicht* in Produktion.
- **Observability:** absichtlich weggelassen; kann mit OTEL/Prom + Grafana später ergänzt werden.
```


