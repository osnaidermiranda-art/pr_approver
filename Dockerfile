# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-s -w" \
      -trimpath \
      -o server .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/server .
RUN chmod +x /app/server

USER app

EXPOSE 8080

ENTRYPOINT ["/app/server"]
