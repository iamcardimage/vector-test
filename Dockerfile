FROM golang:1.24.5-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
    
COPY go.mod go.sum ./
RUN go mod download
    
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/sync ./cmd/sync
    
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app
COPY --from=builder /out/sync /app/sync
    
ENV PORT=8080
EXPOSE 8080
    
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
CMD curl -fsS http://localhost:8080/healthz || exit 1
    
ENTRYPOINT ["/app/sync"]