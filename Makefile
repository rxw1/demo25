SHELL=/bin/zsh

GIT_TAG=$(shell git describe --tags --abbrev=0)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
FRONTEND_VERSION=$(shell jq .version < services/frontend/package.json)
VERSION="$(GIT_TAG)-$(GIT_COMMIT)"

#############################################################################

COMPOSE_FILE=infra/compose.dev.yml

default: up

#COMPOSE_PLAIN=--progress plain

.PHONY: dev-up
up: dev-up
dev-up:
	BUILD_VERSION=$(VERSION) COMPOSE_BAKE=true docker compose $(COMPOSE_PLAIN) -f $(COMPOSE_FILE) up --detach --build

.PHONY: dev-down
down: dev-down
dev-down:
	docker compose -f $(COMPOSE_FILE) down -v

.PHONY: dev-prune
prune: dev-prune
dev-prune:
	docker compose -f $(COMPOSE_FILE) rm -fsv
	docker rmi -f $(shell docker images -qa)

#############################################################################

CLUSTER_NAME=poc

.PHONY: k3d-up
k3d-up:
start: k3d-up
	k3d cluster create $(CLUSTER_NAME) -c infra/cluster.yaml

.PHONY: k3d-down
k3d-down:
stop: k3d-down
	k3d cluster delete $(CLUSTER_NAME)

.PHONY: k3d-prune
k3d-prune:
	# TODO

#############################################################################

.PHONY: install
install:
	$(MAKE) -C infra install
	$(MAKE) -C services/productsvc build import install
	$(MAKE) -C services/ordersvc build import install
	$(MAKE) -C services/frontend build import install

.PHONY: upgrade
upgrade:
	$(MAKE) -C infra upgrade
	$(MAKE) -C services/productsvc build import upgrade
	$(MAKE) -C services/ordersvc build import upgrade
	$(MAKE) -C services/frontend build import upgrade

.PHONY: uninstall
uninstall:
	$(MAKE) -C services/frontend uninstall
	$(MAKE) -C services/ordersvc uninstall
	$(MAKE) -C services/productsvc uninstall
	$(MAKE) -C infra uninstall

#############################################################################

.PHONY: graphql
gql: graphql
graphql:
	$(MAKE) -C services/productsvc gqlgen
	$(MAKE) -C services/ordersvc gqlgen
	$(MAKE) -C services/frontend codegen

.PHONY: lint
lint:
	$(MAKE) -C infra lint
	$(MAKE) -C services/productsvc lint
	$(MAKE) -C services/ordersvc lint
	$(MAKE) -C services/frontend lint

#############################################################################

.PHONY: tests
tests:
	cd tests/e2e && go test -v

#############################################################################

.PHONY: frontend
frontend:
	cd services/frontend && npm run dev

.PHONY: logs
logs:
	docker compose -f infra/compose.dev.yml logs -fn10 {product,order}svc
