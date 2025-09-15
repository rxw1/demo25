<!-- Copilot instructions for the jfpoc workspace -->
# How to be productive in this repo (concise)

This file gives AI coding agents the essentials to be immediately useful in this mono-repo PoC.

- Big picture
  - Components: small Go microservices in `services/` (e.g. `productsvc`, `ordersvc`), a Next.js frontend in `frontend/`, and infra Helm charts in `infra/helm/`.
  - Runtime: services communicate via NATS (events), PostgreSQL is primary relational DB, MongoDB used by some services, Redis for cache. See top-level `README.md` for stack summary.
  - Deploy targets: Docker Compose for local dev (`docker/compose.dev.yml`), and k3d + Helm for Kubernetes-like deploys (`make k3d-up`, `infra/helm/*`).

- Key developer workflows (commands to run)
  - Start full dev environment: `make dev-up` (or `docker compose -f docker/compose.dev.yml up -d --build`).
  - Frontend dev: `cd frontend && npm install && npm run dev` (Next.js on http://localhost:3000).
  - GraphQL codegen (productsvc): `cd services/productsvc && go generate ./...` or `make graph` in `services/ordersvc`.
  - Build & deploy a service into k3d: from a service dir run `make build` then `make import` then `make upgrade` (see `services/*/Makefile`).
  - Kubernetes local cluster: `make k3d-up` (uses `k3d cluster create ... -c infra/k3d/cluster.yaml`).
  - Run migrations/seed: `make migration seed` (top-level Makefile) or check individual service `migrations/` and `main.go` for `AUTO_MIGRATE` usage.
  - E2E tests: `make test_e2e` (runs tests in `tests/e2e/`).

- Project-specific conventions and patterns
  - Single-binary services: migrations are embedded and controlled via `AUTO_MIGRATE=true` (see `services/*/main.go` and README notes).
  - GraphQL: uses gqlgen and codegen (`graph/generated.go`, `gqlgen.yml` in `productsvc`). Run `go generate` before compiling if codegen is stale.
  - Helm charts live under `infra/helm/*`. Service Makefiles call those charts for install/upgrade.
  - Use `k3d image import <name>:latest --cluster poc` when deploying local images to k3d (service Makefiles show this pattern).
  - Feature flags are driven by `flagd` Helm chart and `infra/helm/flagd` (ConfigMap toggles, see top-level README).

- Important files to inspect when changing behavior
  - Repo root: `README.md` (architecture & commands)
  - Frontend: `frontend/README.md`, `frontend/next.config.ts`, `frontend/src/app/*`
  - Services: `services/*/main.go`, `services/*/Makefile`, `services/*/migrations/`, GraphQL schema at `services/productsvc/graph/schema.graphqls`
  - Infra: `infra/k3d/cluster.yaml`, `infra/helm/*` (Helm charts for each service)
  - Docker compose for development: `docker/compose.dev.yml`

- Integration and cross-component notes
  - NATS is used for inter-service events; search `internal/nats` in services to find producers/consumers.
  - DB connections: PostgreSQL is the primary OLTP store; connection setup and migration hooks are in service `internal/db` or `migrations/` folders.
  - When editing schema (GraphQL or DB), update codegen/migrations and run `go generate` or the migration tool before running tests or building images.

- Quick examples
  - Add a GraphQL resolver: edit `services/productsvc/graph/schema.graphqls`, run `go generate ./...`, implement resolver in `graph/schema.resolvers.go`.
  - Add DB migration: put `xxxx_description.up.sql` and `.down.sql` in `services/*/migrations/`, then run the migration via the embedded startup or the migrate CLI.

- When in doubt
  - Read `README.md` and `infra/helm/*` for deployment patterns.
  - Look at `services/*/Makefile`—they contain the canonical dev build/deploy commands.

If anything here is unclear or you want more detail in a specific area (Helm values, GraphQL conventions, or NATS event shapes), tell me which area and I'll expand this file.

---

_Last updated: 2024-06-20_

#### Next Steps:

- Short "how to run tests locally" section per service
- Example NATS event shapes (where to find them)
- Example Helm value overrides or sample values.yaml snippets