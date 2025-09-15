NAMESPACE=infra
POSTGRES_PASSWORD=app

.PHONY: helm
helm:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add nats https://nats-io.github.io/k8s/helm/charts/
	helm repo update

.PHONY: install
install:
	helm upgrade --install flagd ./flagd/chart -n $(NAMESPACE) --create-namespace
	helm upgrade --install nats nats/nats -n $(NAMESPACE) 
	helm upgrade --install mongo bitnami/mongodb -n $(NAMESPACE) 
	helm upgrade --install redis bitnami/redis -n $(NAMESPACE) --set architecture=standalone 
	helm upgrade --install pg bitnami/postgresql -n $(NAMESPACE) --set auth.postgresPassword=$(POSTGRES_PASSWORD) 

.PHONY: upgrade
upgrade:
	helm upgrade flagd ./flagd/chart -n $(NAMESPACE) --reuse-values
	helm upgrade nats nats/nats -n $(NAMESPACE) --reuse-values
	helm upgrade mongo bitnami/mongodb -n $(NAMESPACE) --reuse-values
	helm upgrade redis bitnami/redis -n $(NAMESPACE) --reuse-values
	helm upgrade pg bitnami/postgresql -n $(NAMESPACE) --reuse-values

.PHONY: uninstall
uninstall:
	helm uninstall nats pg mongo redis flagd -n $(NAMESPACE) || true

.PHONY: lint
lint:
	helm lint flagd/chart

.PHONY: flags
flags:
	yq flagd/chart/templates/configmap-flags.yaml | yq '.data["flags.json"]' | jq . > flagd/flags.json
