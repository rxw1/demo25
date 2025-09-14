#!/usr/bin/env zsh

# from:
# `helm upgrade --install mongo bitnami/mongodb -n infra --create-namespace`

# MongoDB&reg; can be accessed on the following DNS name(s) and ports from
# within your cluster: mongo-mongodb.infra.svc.cluster.local

# To get the root password run:
export MONGODB_ROOT_PASSWORD=$(kubectl get secret --namespace infra \
    mongo-mongodb -o jsonpath="{.data.mongodb-root-password}" | base64 -d)

# To connect to your database, create a MongoDB&reg; client container:
kubectl run --namespace infra mongo-mongodb-client --rm --tty -i \
    --restart='Never' --env="MONGODB_ROOT_PASSWORD=$MONGODB_ROOT_PASSWORD" \
    --image docker.io/bitnami/mongodb:8.0.13-debian-12-r0 --command -- bash

# Then, run the following command:
mongosh admin --host "mongo-mongodb" --authenticationDatabase admin -u \
    $MONGODB_ROOT_USER -p $MONGODB_ROOT_PASSWORD

# To connect to your database from outside the cluster execute the following commands:
kubectl port-forward --namespace infra svc/mongo-mongodb 27017:27017 &
mongosh --host 127.0.0.1 --authenticationDatabase admin -p $MONGODB_ROOT_PASSWORD
