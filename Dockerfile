# syntax=docker/dockerfile:1.7
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
# Cache de dependencias: solo se re-descarga si go.mod/go.sum cambian
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Cache de compilacion: reutiliza artefactos entre builds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -trimpath \
      -o server .

# Runtime stage — imagen minima sin shell ni herramientas innecesarias
FROM alpine:3.21

RUN apk --no-cache add ca-certificates wget \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/server .

# Ejecutar como usuario no-root
USER app

EXPOSE 8080

ENTRYPOINT ["./server"]
