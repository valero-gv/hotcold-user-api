# syntax=docker/dockerfile:1

FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags "-s -w" -o server ./cmd/api

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /app/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]