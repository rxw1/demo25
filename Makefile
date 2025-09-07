.PHONY: gen gql fe dev-up dev-down k3d-up k3d-down helm-deps helm-install helm-uninstall

gen: gql fe

gql:
	cd services/product-svc && go run github.com/99designs/gqlgen generate
	cd services/product-svc && go generate -v -x ./... # FIXME

fe:
	cd frontend && npm run codegen || true

# --- Docker Compose Dev (schneller Dev-Loop) ---

dev-up:
	docker compose -f docker/compose.dev.yml up -d --build

dev-down:
	docker compose -f docker/compose.dev.yml down -v

# --- k3d + Helm ---

k3d-up:
	k3d cluster create poc -c infra/k3d/cluster.yaml

k3d-down:
	k3d cluster delete poc

helm-deps:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add nats https://nats-io.github.io/k8s/helm/charts/
	helm repo update

helm-install: helm-deps
	helm upgrade --install nats nats/nats -n infra --create-namespace
	helm upgrade --install pg bitnami/postgresql -n infra --set auth.postgresPassword=postgres
	helm upgrade --install mongo bitnami/mongodb -n infra
	helm upgrade --install redis bitnami/redis -n infra --set architecture=standalone
	helm upgrade --install flagd infra/helm/flagd -n app --create-namespace
	helm upgrade --install product-svc infra/helm/product-svc -n app
	helm upgrade --install order-svc infra/helm/order-svc -n app

helm-uninstall:
	helm uninstall product-svc order-svc -n app || true
	helm uninstall flagd -n app || true
	helm uninstall nats pg mongo redis -n infra || true

.PHONY: clean-gql
clean-gql:
	rm services/product-svc/graph/model/models_gen.go
	rm services/product-svc/graph/generated.go

seed:
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "create table if not exists products(id text primary key, name text, price int);"
	docker exec -it $(docker ps -qf name=postgres) psql -U app -d app -c "insert into products(id,name,price) values('p1','Widget',199) on conflict do nothing;"
