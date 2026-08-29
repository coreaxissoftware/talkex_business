# Multi-stage build for the Go API. Producing a small final image via
# a distroless base keeps the container startup fast and the attack
# surface minimal.
FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/talkex-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/talkex-api /app/talkex-api
# Uploads directory (media library). Docker Compose mounts a volume here.
VOLUME ["/app/uploads"]
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/talkex-api"]
