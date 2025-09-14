# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY . .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ordersvc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/ordersvc /ordersvc
EXPOSE 8081
ENTRYPOINT ["/ordersvc"]

