# syntax=docker/dockerfile:1
FROM golang:1.19-alpine as builder
WORKDIR /build
ADD go.mod go.sum /build/
RUN --mount=type=cache,target=~/.cache/go-build go mod download
ADD . /build/
RUN --mount=type=cache,target=~/.cache/go-build CGO_ENABLED=0 go test ./...
RUN --mount=type=cache,target=~/.cache/go-build CGO_ENABLED=0 go build -o event-tap .

FROM alpine

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/event-tap /app/event-tap
ENTRYPOINT ["/app/event-tap"]
