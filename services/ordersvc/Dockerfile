# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src

COPY go.work ./

COPY pkg/flags/go.mod ./pkg/flags/
COPY pkg/logging/go.mod ./pkg/logging/
COPY pkg/model/go.mod ./pkg/model/

COPY services/gatewaysvc/go.mod ./services/gatewaysvc/
COPY services/ordersvc/go.mod ./services/ordersvc/
COPY services/productsvc/go.mod ./services/productsvc/

COPY tests/e2e/go.mod ./tests/e2e/

WORKDIR /src/services/ordersvc

RUN go mod download

WORKDIR /src

COPY pkg/flags/ ./pkg/flags/
COPY pkg/logging/ ./pkg/logging/
COPY pkg/model/ ./pkg/model/

COPY services/gatewaysvc/ ./services/gatewaysvc/
COPY services/ordersvc/ ./services/ordersvc/
COPY services/productsvc/ ./services/productsvc/

WORKDIR /src/services/ordersvc

RUN CGO_ENABLED=0 go build -o /out/ordersvc .

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/ordersvc /ordersvc
EXPOSE 8082
ENTRYPOINT ["/ordersvc"]

