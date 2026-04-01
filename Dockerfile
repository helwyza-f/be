# STAGE 1: Build binary menggunakan Debian Bookworm
FROM golang:1.25-bookworm AS builder

# Install build-essential jika ada dependency CGO (opsional tapi aman)
RUN apt-get update && apt-get install -y git ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod & sum dulu agar Docker layer caching jalan
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary Go secara statis
# CGO_ENABLED=0 memastikan binary tidak tergantung library OS host
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# STAGE 2: Final image menggunakan Debian Bookworm Slim (Lebih ringan buat Running)
FROM debian:bookworm-slim

# Install ca-certificates (WAJIB buat koneksi HTTPS ke AWS S3/RDS) & timezone data
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/main .

# Copy folder migrations untuk auto-migrate di server
COPY --from=builder /app/migrations ./migrations

# Gunakan port 8080 sesuai config Batam Engine
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]