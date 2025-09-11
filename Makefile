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

.PHONY: helm-install-infra
helm-install-infra: nats pg mongo redis flagd infra-jobs

.PHONY: helm-install-app
helm-install-app: product-svc-chart order-svc-chart frontend-chart


.PHONY: helm
install: helm-setup helm-install-infra helm-install-app

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall product-svc order-svc frontend -n app || true
	helm uninstall nats pg mongo redis flagd -n infra || true

.PHONY: helm-lint
helm-lint:
	for i in infra/helm/*; do helm lint $$i; done

###

.PHONY: nats
nats:
	-helm upgrade --install nats nats/nats -n infra --create-namespace

.PHONY: pg
pg:
	-helm upgrade --install pg bitnami/postgresql -n infra --set auth.postgresPassword=$(POSTGRES_PASSWORD) --create-namespace

.PHONY: mongo
mongo:
	-helm upgrade --install mongo bitnami/mongodb -n infra --create-namespace

.PHONY: redis
redis:
	-helm upgrade --install redis bitnami/redis -n infra --set architecture=standalone --create-namespace

.PHONY: flagd
flagd:
	-helm upgrade --install flagd infra/helm/flagd -n infra --create-namespace

.PHONY: infra-jobs
infra-jobs:
	-helm --debug upgrade --install infra-jobs infra/helm/infra-jobs -n infra --create-namespace

###

.PHONY: product-svc-chart
product-svc-chart:
	-helm upgrade --install product-svc infra/helm/product-svc -n app --create-namespace

.PHONY: order-svc-chart
order-svc-chart:
	-helm upgrade --install order-svc infra/helm/order-svc -n app --create-namespace

.PHONY: frontend-chart
frontend-chart:
	-helm upgrade --install frontend infra/helm/frontend -n app --create-namespace

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

.PHONY: status
status: helm_status pod_status docker_status

.PHONY: helm_status
helm_status:
	@echo; helm list -A

.PHONY: pod_status
pod_status:
	@echo; kubectl get pods -A

.PHONY: docker_status
docker_status:
	@echo; docker ps --format '{{.Names}}\n\tContainer ID: {{.ID}}\n\tCommand: {{.Command}}\n\tImage: {{.Image}}\n\tCreatedAt: {{.CreatedAt}}\n\tStatus: {{.Status}}\n'

# image
.PHONY: docker-product-svc
docker-product-svc:
	docker build -t docker-product-svc:latest services/product-svc

# image
.PHONY: docker-order-svc
docker-order-svc:
	docker build -t docker-order-svc:latest services/order-svc

# container
.PHONY: product-svc-container
product-svc-container: docker-product-svc
	docker run -d --name product-svc docker-product-svc:latest

# container
.PHONY: order-svc-container
order-svc-container: docker-order-svc
	docker run -d --name order-svc docker-order-svc:latest

