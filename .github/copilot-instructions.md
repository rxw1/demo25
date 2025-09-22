# Copilot instructions for this repo

Purpose: Guide AI coding agents to be productive quickly in this monorepo. Keep changes small, runnable, and tested.

## Big picture
- Architecture: gateway + three backend services + frontend + infra + e2e tests
  - gatewaysvc (Go, GraphQL API on :8080): Single GraphQL endpoint using gqlgen with WebSocket subscriptions; orchestrates data via NATS request/reply to backend services; feature flags via flagd; optional Redis cache for product reads.
    - Key dirs: `services/gatewaysvc/internal/{graphql,cache}`
  - productsvc (Go, HTTP on :8081): Provides product data over NATS request/reply (subjects `products.*`) backed by PostgreSQL. Runs DB migrations when `AUTO_MIGRATE=true`.
    - Key dirs: `services/productsvc/internal/{db,handle}`
  - ordersvc (Go, HTTP on :8082): Materializes order events into MongoDB and serves health. Subscribes to NATS (`order.created`) and responds to queries (`orders.*`).
    - Key dirs: `services/ordersvc/internal/{db,handle}`
  - usersvc (Go, HTTP on :8083): Example service for user data via NATS (`users.get`).
    - Key dirs: `services/usersvc/internal/{db,handle}`
  - frontend (Next.js/React): Apollo client to gatewaysvc GraphQL on :8080. Subscriptions via `graphql-ws`.
    - Key dirs: `services/frontend/src/app/**`, codegen config in `services/frontend/graphql/codegen.ts`.
  - infra: Local dev docker-compose and k3d/Helm charts for k8s. Bitnami charts for deps. Flagd chart and flags json.

## Local dev workflows
- All-in-one docker compose (recommended to demo end-to-end):
  - `make up` to build and start compose (`infra/compose.dev.yml`)
  - `make down` to stop and prune volumes
  - Services come up on:
    - gatewaysvc: http://localhost:8080 (GraphQL + Playground at `/`)
    - productsvc: http://localhost:8081/healthz
    - ordersvc: http://localhost:8082/healthz
    - usersvc: http://localhost:8083/healthz
    - frontend: http://localhost:8088
- Kubernetes with k3d:
  - `make start` to create cluster from `infra/cluster.yaml`; `make install` to install infra and services Helm charts; `make upgrade` to roll out changes
- GraphQL code generation:
  - Backend: `make -C services/gatewaysvc gqlgen` regenerates `services/gatewaysvc/internal/graphql/generated.go` and models
  - Frontend: `npm run codegen` or `npm run codegen-watch` in `services/frontend` (schema source reads `NEXT_PUBLIC_GRAPHQL_URL` or defaults to `http://localhost:8080/graphql`)
- Tests:
  - End-to-end: `make tests` runs Go e2e test in `tests/e2e`, hits gatewaysvc GraphQL (`http://localhost:8080/graphql`) and asserts Mongo materialization in ordersvc

## Data flow
- Create order: frontend -> gatewaysvc GraphQL mutation -> publish `order.created` (NATS) -> ordersvc subscribes and upserts to Mongo -> gatewaysvc `orders` query does NATS request `orders.all` to ordersvc -> frontend displays.
- Subscriptions: gatewaysvc subscribable fields (`lastOrderCreated`, `flagState`) stream NATS events (`order.created`, `flags.state`) to connected WebSocket clients.

## Conventions and patterns
- Logging: shared `pkg/logging` exposes `logging.With(ctx, ...)` and `logging.From(ctx)`; prefer context-scoped logging, no globals.
- Feature flags: `pkg/flags` with flagd; `RedisEnabled(ctx)` and `ThrottleEnabled(ctx)` gate cache/throttle in resolvers/subscribers. Update `infra/flagd/flags.json` and run `make -C infra flags` to sync the configmap template.
- Caching: simple Redis wrapper in `services/gatewaysvc/internal/cache`. Keys often `product:<id>`; cache use is guarded by feature flag.
- GraphQL backend: schema in `services/gatewaysvc/internal/graphql/schema.graphqls`; resolvers in `schema.resolvers.go`; DI in `resolver.go`.
- NATS subjects (current):
  - Events (publish): `order.created`, `order.canceled`, `flags.state`
  - Request/Reply (gateway -> services): `orders.all`, `orders.get`, `orders.by_user`, `products.all`, `products.get`, `users.get`
- Frontend GraphQL client: `services/frontend/src/app/page.tsx` wires Apollo with split link; URL derived from `NEXT_PUBLIC_GRAPHQL_URL` (fallback `http://localhost:8080/graphql`). Use generated documents in `src/app/__generated__/` rather than inline strings.

## Env and ports
- Compose wires env:
  - gatewaysvc: `NATS_URL`, `REDIS_ADDR`, `FLAGD_HOST/PORT`, `WS_ALLOWED_ORIGINS`
  - productsvc: `DATABASE_URL`, `NATS_URL`, `AUTO_MIGRATE=true`, `FLAGD_HOST/PORT`
  - ordersvc: `MONGO_URI`, `NATS_URL`, `FLAGD_HOST/PORT`
  - frontend: uses `NEXT_PUBLIC_GRAPHQL_URL` for browser GraphQL codegen/runtime
  - WebSocket origin allowlist: set `WS_ALLOWED_ORIGINS` (comma-separated) to permit additional origins for subscriptions; gateway defaults include `http://localhost:8088` and always allow same-host.

## Common tasks examples
- Add a GraphQL field:
  1) Edit `services/gatewaysvc/internal/graphql/schema.graphqls`
  2) Run `make gql` (root) to regenerate gateway server and frontend clients
  3) Implement resolver in `services/gatewaysvc/internal/graphql/schema.resolvers.go` (DI via `Resolver` fields)
  4) Use the new generated query/mutation in frontend via generated docs in `services/frontend/src/app/__generated__/`
- Add/adjust a NATS handler:
  - productsvc: implement in `services/productsvc/internal/handle/*` and subscribe in `main.go`
  - ordersvc: implement in `services/ordersvc/internal/handle/*` and subscribe in `main.go`
  - usersvc: implement in `services/usersvc/internal/handle/*` and subscribe in `main.go`

## Gotchas
- WebSocket CORS/origin: gatewaysvc WS upgrader validates `Origin`. If your dev host differs, set `WS_ALLOWED_ORIGINS` accordingly or use same-host port.
- Frontend codegen schema URL now reads `NEXT_PUBLIC_GRAPHQL_URL`; set it for non-local runs to avoid mismatches.
- e2e test assumes p1 product exists and services are reachable on localhost ports from the host.

## Where to start
- For gateway GraphQL changes: `services/gatewaysvc/internal/graphql/*`
- For frontend changes: `services/frontend/src/app/components/*`
- For infra: `infra/compose.dev.yml` and `infra/Makefile`

