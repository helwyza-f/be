# STAGE 1: Build binary
FROM golang:1.22-alpine AS builder

# Install git buat download dependency kalau perlu
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod dulu biar caching layer Docker efisien
COPY go.mod go.sum ./
RUN go mod download

# Copy semua file project
COPY . .

# Build binary Go ke file bernama 'main'
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# STAGE 2: Final image yang ringan
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/main .
# Copy folder migrations karena app butuh buat auto-migrate saat start
COPY --from=builder /app/migrations ./migrations

# Expose port sesuai settingan .env kamu
EXPOSE 8080

# Jalankan app
CMD ["./main"]