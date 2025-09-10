POSTGRES_PASSWORD=postgres

# NOTE about calling gqlgen
#
# go run github.com/99designs/gqlgen generate
#
# - directly runs the gqlgen generator via the module path; always executes
#   gqlgen and writes the generated files.
# - no reliance on any source comment directives; works anywhere (assuming
#   module available).
#
# go generate ./...
#
# - scans source files for //go:generate comments and runs the commands found
#   there, package by package.
# - it does NOT implicitly run gqlgen unless you have a go:generate directive
#   that calls it (or an installed binary named in the directive).
# - it will be a no-op if there are no go:generate directives pointing to
#   gqlgen (or if the directives point to a binary that isn’t installed).

.PHONY: gen
gen: gql fe

.PHONY: gql
gql:
	cd services/product-svc && go run github.com/99designs/gqlgen generate

.PHONY: fe
fe:
	cd frontend && npm run codegen || true

###

COMPOSE_FILE=docker/compose.dev.yml

.PHONY: dev-up
dev-up:
	COMPOSE_BAKE=true docker compose -f $(COMPOSE_FILE) up -d --build

.PHONY: dev-down
dev-down:
	docker compose -f $(COMPOSE_FILE) down -v

###

.PHONY: k3d-up
k3d-up:
	k3d cluster create poc -c infra/k3d/cluster.yaml

.PHONY: k3d-down
k3d-down:
	k3d cluster delete poc

###

.PHONY: helm-setup
helm-setup:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add nats https://nats-io.github.io/k8s/helm/charts/
	helm repo update

.PHONY: helm
helm: helm-setup nats pg mongo redis flagd infra-jobs products-svc order-svc frontend

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall frontend || true
	helm uninstall product-svc order-svc -n app || true
	helm uninstall flagd -n app || true
	helm uninstall nats pg mongo redis -n infra || true

.PHONY: helm-lint
helm-lint:
	for i in infra/helm/*; do helm lint $$i; done

###

.PHONY: nats
nats:
	@helm upgrade --install nats nats/nats -n infra --create-namespace

.PHONY: pg
pg:
	@helm upgrade --install pg bitnami/postgresql -n infra --set auth.postgresPassword=$(POSTGRES_PASSWORD)

.PHONY: mongo
mongo:
	@helm upgrade --install mongo bitnami/mongodb -n infra

.PHONY: redis
redis:
	@helm upgrade --install redis bitnami/redis -n infra --set architecture=standalone

.PHONY: infra-jobs
infra-jobs:
	# FIXME
	@helm --debug upgrade --install infra-jobs infra/helm/infra-jobs -n infra

###

.PHONY: flagd
flagd:
	# FIXME
	@helm upgrade --install flagd infra/helm/flagd -n app --create-namespace

.PHONY: product-svc
product-svc:
	@helm upgrade --install product-svc infra/helm/product-svc -n app

.PHONY: order-svc
order-svc:
	@helm upgrade --install order-svc infra/helm/order-svc -n app

.PHONY: frontend
frontend:
	@helm upgrade --install frontend infra/helm/frontend -n app

###

.PHONY: seed
seed:
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "create table if not exists products(id text primary key, name text, price int);"
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing;"

.PHONY: migration-seed
migration-seed:
	kubectl apply -f infra/helm/product-svc/templates/configmap-migrations.yaml
	kubectl apply -f infra/helm/product-svc/templates/configmap-seed.yaml
	kubectl apply -f infra/helm/product-svc/templates/job-migrate.yaml

.PHONY: test-e2e
test-e2e:
	cd tests/e2e && go test -v

.PHONY: clean
clean: helm-uninstall k3d-down dev-down
	
.PHONY: clean-gql
clean-gql:
	rm services/product-svc/graph/model/models_gen.go
	rm services/product-svc/graph/generated.go

