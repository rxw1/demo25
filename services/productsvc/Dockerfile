# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src

# Use Go workspace and seed module cache efficiently
# 1) Copy workspace file and module manifests only (better caching)
COPY go.work ./
# Copy go.mod for ALL modules listed in go.work so workspace validation passes
COPY pkg/logging/go.mod ./pkg/logging/
COPY services/productsvc/go.mod ./services/productsvc/
COPY services/ordersvc/go.mod ./services/ordersvc/
COPY tests/e2e/go.mod ./tests/e2e/

# 2) Download deps for productsvc in workspace context
WORKDIR /src/services/productsvc
RUN go mod download

# 3) Copy only the needed source trees
WORKDIR /src
COPY pkg/logging/ ./pkg/logging/
COPY services/productsvc/ ./services/productsvc/

# 4) Build productsvc
WORKDIR /src/services/productsvc
RUN go generate ./... && CGO_ENABLED=0 go build -o /out/productsvc .

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/productsvc /productsvc
EXPOSE 8080
ENTRYPOINT ["/productsvc"]
