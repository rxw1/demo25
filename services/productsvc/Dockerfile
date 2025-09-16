# syntax=docker/dockerfile:1
FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY . .
RUN go mod download
COPY . .
RUN go generate ./... && CGO_ENABLED=0 go build -o /out/productsvc ./

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/productsvc /productsvc
EXPOSE 8080
ENTRYPOINT ["/productsvc"]

