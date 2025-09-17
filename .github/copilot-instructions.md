# Copilot instructions for this repo

Purpose: Guide AI coding agents to be productive quickly in this monorepo. Keep changes small, runnable, and tested.

## Big picture
- Architecture: three services + infra + e2e tests
  - productsvc (Go, GraphQL API on :8080): PostgreSQL read, Redis cache, NATS for request/reply and events, feature flags via flagd. GraphQL server uses gqlgen with WebSocket subscriptions.
    - Key dirs: `services/productsvc/internal/{db,graphql,cache,flags,logging}`
  - ordersvc (Go, HTTP on :8081): Materializes order events into MongoDB and serves health. Subscribes to NATS subjects.
    - Key dirs: `services/ordersvc/internal/{mongo,nats,logging}`
  - frontend (Next.js/React on :3333 container, :3000 dev): Apollo client to productsvc GraphQL. Subscriptions via `graphql-ws`.
    - Key dirs: `services/frontend/src/app/**`, codegen config in `services/frontend/graphql/codegen.ts`.
  - infra: Local dev docker-compose and k3d/Helm charts for k8s. Bitnami charts for deps. Flagd chart and flags json.

## Local dev workflows
- All-in-one docker compose (recommended to demo end-to-end):
  - `make up` to build and start compose (`infra/compose.dev.yml`)
  - `make down` to stop and prune volumes
  - Services come up on:
    - productsvc: http://localhost:8080 (GraphQL + Playground at `/`)
    - ordersvc: http://localhost:8081/healthz
    - frontend: http://localhost:3000
- Kubernetes with k3d:
  - `make start` to create cluster from `infra/cluster.yaml`; `make install` to install infra and services Helm charts; `make upgrade` to roll out changes
- GraphQL code generation:
  - Backend: `make -C services/productsvc gqlgen` regenerates `internal/graphql/generated.go`
  - Frontend: `npm run codegen` or `npm run codegen-watch` in `services/frontend` (schema source reads `NEXT_PUBLIC_GRAPHQL_URL` or defaults to `http://localhost:8080/graphql`)
- Tests:
  - End-to-end: `make tests` runs Go e2e test in `tests/e2e`, hits productsvc GraphQL and asserts Mongo materialization in ordersvc

## Data flow
- Create order: frontend -> productsvc GraphQL mutation -> publish `order.created` event (NATS) -> ordersvc subscriber upserts to Mongo -> productsvc `orders` query does NATS request `orders.all` to ordersvc -> frontend displays.
- Subscriptions: productsvc subscribable fields (`lastOrderCreated`) stream NATS `order.created` events to connected WebSocket clients.

## Conventions and patterns
- Logging: project-local `internal/logging` packages for Go services provide `logging.With(ctx, ...attrs)` and `logging.From(ctx)` helpers; prefer context-scoped logging, no globals. See `services/*/internal/logging/logger.go`.
- Feature flags: `flags.Flags.RedisEnabled(ctx)` gates Redis usage; flagd deployed via Helm. Update `infra/flagd/flags.json` and run `make -C infra flags` to sync configmap value if needed.
- Caching: simple Redis wrapper in `productsvc/internal/cache`. Keys: `products:all`, `product:<id>`.
- GraphQL backend: schema in `services/productsvc/internal/graphql/schema.graphqls`; resolvers in `schema.resolvers.go`; DI in `resolver.go`.
- NATS subjects:
  - Publish: `order.created` (productsvc mutation)
  - Request/Reply: `orders.all` (productsvc query -> ordersvc)
  - Price example subject `products.price` is requested by productsvc resolver; provider not included in this repo.
- Frontend GraphQL client: `services/frontend/src/app/page.tsx` wires Apollo with split link; URL derived from `NEXT_PUBLIC_GRAPHQL_URL` env (fallback localhost). Use generated documents from `__generated__/graphql.ts` instead of inline strings.

## Env and ports
- Compose wires env:
  - productsvc: `DATABASE_URL`, `REDIS_ADDR`, `NATS_URL`, `AUTO_MIGRATE=true`, `FLAGD_HOST/PORT`
  - ordersvc: `MONGO_URI`, `NATS_URL`
  - frontend: exposes `PRODUCTSVC_URL`/`ORDERSVC_URL` at runtime (Next.js server), and for the browser GraphQL use `NEXT_PUBLIC_GRAPHQL_URL` if set.
  - productsvc WebSocket origin allowlist: set `WS_ALLOWED_ORIGINS` (comma-separated) to permit additional origins for subscriptions; defaults include `http://localhost:3000, http://localhost:3001`.

## Common tasks examples
- Add a GraphQL field:
  1) Edit `services/productsvc/internal/graphql/schema.graphqls`
  2) Run `make gql` (root) to regenerate server and frontend clients
  3) Implement resolver in `schema.resolvers.go`, using DI via `Resolver` fields
  4) Use the new generated query/mutation in frontend via `__generated__/graphql.ts`
- Add a NATS subscriber (ordersvc): implement in `services/ordersvc/internal/nats/*` and wire in `main.go`.

## Gotchas
- WebSocket CORS/origin: productsvc WS upgrader validates `Origin`. If your dev host differs, set `WS_ALLOWED_ORIGINS` accordingly or use same-host port.
- Frontend codegen schema URL now reads `NEXT_PUBLIC_GRAPHQL_URL`; set it for non-local runs to avoid mismatches.
- e2e test assumes p1 product exists and services are reachable on localhost ports from the host.

## Where to start
- For backend changes: `services/productsvc/internal/graphql/schema.graphqls` and `schema.resolvers.go`
- For frontend changes: `services/frontend/src/app/components/*`
- For infra: `infra/compose.dev.yml` and `infra/Makefile`

