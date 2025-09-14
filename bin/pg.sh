# from:
# helm upgrade --install pg bitnami/postgresql -n infra --set
# auth.postgresPassword=postgres --create-namespace

# PostgreSQL can be accessed via port 5432 on the following DNS names from
# within your cluster: pg-postgresql.infra.svc.cluster.local - Read/Write
# connection

# To get the password for "postgres" run:
export POSTGRES_PASSWORD=$(kubectl get secret --namespace infra \
    pg-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)

# To connect to your database run the following command:
kubectl run pg-postgresql-client --rm --tty -i --restart='Never' \
    --namespace infra --image docker.io/bitnami/postgresql:17.6.0-debian-12-r4 \
    --env="PGPASSWORD=$POSTGRES_PASSWORD" --command -- \
    psql --host pg-postgresql -U postgres -d postgres -p 5432

# NOTE: If you access the container using bash, make sure that you execute
# "/opt/bitnami/scripts/postgresql/entrypoint.sh /bin/bash" in order to avoid
# the error "psql: local user with ID 1001} does not exist"

# To connect to your database from outside the cluster execute the following commands:
kubectl port-forward --namespace infra svc/pg-postgresql 5432:5432 &
PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U postgres -d postgres -p 5432
