# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY . .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/order-svc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/order-svc /order-svc
EXPOSE 8081
ENTRYPOINT ["/order-svc"]

