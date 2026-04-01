# STAGE 1: Build binary (Di GitHub Actions)
FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# STAGE 2: Runner (Image Final)
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
# Ambil binary dari builder
COPY --from=builder /app/main .
# Ambil folder migrations untuk database
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./main"]