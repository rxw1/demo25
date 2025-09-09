.PHONY: gen gql fe dev-up dev-down k3d-up k3d-down helm-deps helm-install helm-uninstall

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

gen: gql fe

gql:
	cd services/product-svc && go run github.com/99designs/gqlgen generate
	#cd services/product-svc && go generate -v -x ./...

fe:
	cd frontend && npm run codegen || true

###

dev-up:
	COMPOSE_BAKE=true docker compose -f docker/compose.dev.yml up -d --build

dev-down:
	docker compose -f docker/compose.dev.yml down -v

###

k3d-up:
	k3d cluster create poc -c infra/k3d/cluster.yaml

k3d-down:
	k3d cluster delete poc

###

helm-deps:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add nats https://nats-io.github.io/k8s/helm/charts/
	helm repo update

helm-install: helm-deps
	helm upgrade --install nats nats/nats -n infra --create-namespace
	helm upgrade --install pg bitnami/postgresql -n infra --set auth.postgresPassword=postgres
	helm upgrade --install mongo bitnami/mongodb -n infra
	helm upgrade --install redis bitnami/redis -n infra --set architecture=standalone
	helm upgrade --install infra-jobs infra/helm/infra-jobs -n infra
	helm upgrade --install flagd infra/helm/flagd -n app --create-namespace
	helm upgrade --install product-svc infra/helm/product-svc -n app
	helm upgrade --install order-svc infra/helm/order-svc -n app
	helm upgrade --install frontend infra/helm/frontend -n app

helm-uninstall:
	helm uninstall product-svc order-svc -n app || true
	helm uninstall flagd -n app || true
	helm uninstall nats pg mongo redis -n infra || true

###

.PHONY: seed
seed:
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "create table if not exists products(id text primary key, name text, price int);"
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing;"

.PHONY: migration_seed
migration_seed:
	kubectl apply -f infra/helm/product-svc/templates/configmap-migrations.yaml
	kubectl apply -f infra/helm/product-svc/templates/configmap-seed.yaml
	kubectl apply -f infra/helm/product-svc/templates/job-migrate.yaml

.PHONY: test_e2e
test_e2e:
	cd tests/e2e && go test -v

.PHONY: clean
clean: helm-uninstall k3d-down dev-down
	
.PHONY: clean_generated_gql
clean_generated_gql:
	rm services/product-svc/graph/model/models_gen.go
	rm services/product-svc/graph/generated.go

