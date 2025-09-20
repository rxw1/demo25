NAME=ordersvc
CLUSTER_NAME=poc
NAMESPACE=app
CHART=./chart

ok:

.PHONY: build
build:
	docker build -t $(NAME):latest .

.PHONY: import
import:
	k3d image import $(NAME):latest --cluster $(CLUSTER_NAME)

.PHONY: install
install:
	helm upgrade --install $(NAME) $(CHART) -n $(NAMESPACE) --create-namespace

.PHONY: upgrade
upgrade:
	helm upgrade $(NAME) $(CHART) -n $(NAMESPACE) --reuse-values

.PHONY: restart
restart:
	kubectl rollout restart deployment $(NAME) -n $(NAMESPACE)
	kubectl rollout status deployment $(NAME) -n $(NAMESPACE)

.PHONY: deploy
deploy: build import upgrade restart

.PHONY: run
run:
	docker run -d --name $(NAME) $(NAME):latest

.PHONY: lint
lint:
	helm lint chart
	staticcheck ./...

.PHONY: test-dns
test-dns:
	kubectl run -it --rm busybox --image=busybox --restart=Never -- nslookup $(NAME).$(NAMESPACE).svc.cluster.local

.PHONY: gqlgen
graphql: gqlgen
gqlgen:
	go run github.com/99designs/gqlgen generate

.PHONY: name
name:
	@awk -F/ 'NR==1{print $$NF}' < go.mod

###

