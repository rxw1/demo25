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

## 2) Kubernetes (k3d + Helm)
- `make k3d-up`
- `make helm-install`
- Status prüfen: `kubectl get pods -n app` und `kubectl get pods -n infra`
- Port-Forward Frontend (falls containerisiert): `kubectl port-forward svc/product-svc 8080:80 -n app`

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
    - cd services/product-svc && go vet ./...
    - cd ../order-svc && go vet ./...

go-test:
  stage: test
  image: golang:1.22
  script:
    - cd services/product-svc && go test ./... || true
    - cd ../order-svc && go test ./... || true

docker-build-product:
  stage: build
  image: gcr.io/kaniko-project/executor:latest
  script:
    - /kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile services/product-svc/Dockerfile --destination $CI_REGISTRY_IMAGE/product-svc:${CI_COMMIT_SHORT_SHA}

docker-build-order:
  stage: build
  image: gcr.io/kaniko-project/executor:latest
  script:
    - /kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile services/order-svc/Dockerfile --destination $CI_REGISTRY_IMAGE/order-svc:${CI_COMMIT_SHORT_SHA}
```

---

## Dockerfiles (Beispiel: product-svc & order-svc)

**services/product-svc/Dockerfile**
```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY services/product-svc/go.mod services/product-svc/go.sum ./
RUN go mod download
COPY services/product-svc/. .
RUN go generate ./... && CGO_ENABLED=0 go build -o /out/product-svc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/product-svc /product-svc
EXPOSE 8080
ENTRYPOINT ["/product-svc"]
```

**services/order-svc/Dockerfile**
```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY services/order-svc/go.mod services/order-svc/go.sum ./
RUN go mod download
COPY services/order-svc/. .
RUN CGO_ENABLED=0 go build -o /out/order-svc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/order-svc /order-svc
EXPOSE 8081
ENTRYPOINT ["/order-svc"]
```

---

## Hinweise
- **gqlgen-Code:** einmal `go generate ./...` im `product-svc` ausführen (oder im Docker-Build). Eine `tools.go` mit `//go:build tools` kann `gqlgen` pinnen.
- **Sicherheit:** für den PoC sind Passwörter/URIs simpel gehalten. Bitte *nicht* in Produktion.
- **Observability:** absichtlich weggelassen; kann mit OTEL/Prom + Grafana später ergänzt werden.
```


