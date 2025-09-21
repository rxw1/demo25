# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src

COPY go.work ./

# Workspace-aware dependency caching
COPY pkg/logging/go.mod ./pkg/logging/
COPY services/gatewaysvc/go.mod ./services/gatewaysvc/
COPY services/productsvc/go.mod ./services/productsvc/
COPY services/ordersvc/go.mod ./services/ordersvc/
COPY tests/e2e/go.mod ./tests/e2e/

# Download dependencies
WORKDIR /src/services/productsvc
RUN go mod download

# Copy sources
WORKDIR /src
COPY pkg/logging/ ./pkg/logging/
COPY services/gatewaysvc/ ./services/gatewaysvc/
COPY services/productsvc/ ./services/productsvc/
COPY services/ordersvc/ ./services/ordersvc/

# Build
WORKDIR /src/services/productsvc
RUN CGO_ENABLED=0 go build -o /out/productsvc .

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/productsvc /productsvc
EXPOSE 8081
ENTRYPOINT ["/productsvc"]

