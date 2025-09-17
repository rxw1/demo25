# Stack 9/11/2025

- Golang (Backend)
- React/Next.js (Frontend)
- Events/Messaging (~~RabbitMQ~~ NATS)
- Databases (PostgreSQL, MongoDB)
- Caching (Redis)
- GraphQL (gqlgen, GraphQL Codegen, Apollo Server)
- Docker/Compose and Kubernetes (k3s)
- CI/CD (~~GitLab CI~~ GitHub Actions)
- IaC (~~Terraform~~ Helm)
- Feature Flags (flagd)

## Requirements

- Go 1.25, Node 18+, Docker, Docker Compose, k3d, Helm

## Development

### Docker Compose

- `make up`
- `make down`

### Kubernetes (k3d)

- `make start`
- `make stop`
- `make install`: Install Helm charts.

### Tests

- `make tests-e2e`
- `make tests`: Run all tests

### Makefiles

- See the various Makefiles for more information about how things are set up:

        Makefile
        infra/Makefile
        services/frontend/Makefile
        services/ordersvc/Makefile
        services/productsvc/Makefile

### Frontend Dev

In `services/frontend` run
- `npm run dev` and
- `npm run codegen-watch`

## Infrastructure

- General configuration lives in `infra` (see `infra/Makefile`), specific package configuration lives in the package directory, e.g. `services/productsvc/Dockerfile` or `services/productsvc/chart` for Helm charts.

- Flagd configuration lives in `infra/flagd`. Run `make -C infra flagd` to synchronize `infra/flagd/flags.json` and `infra/flagd/chart/templates/configmap-flags.yaml`. **Only edit `infra/flagd/flags.json`, then run `make flagd` in `infra` to update the data value in `infra/flagd/chart/templates/configmap-flags.yaml`.**
