# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./ 
# Ensure go.mod and go.sum are present if you have dependencies, otherwise remove this line or create empty ones.
# For this simple script, they might not be strictly necessary if there are no external packages beyond standard library.
# If you create go.mod: go mod init daypassed-bot; go mod tidy
RUN if [ -f go.mod ]; then go mod download; fi
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o daypassed-bot .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/daypassed-bot .
# The TZ environment variable should be set by the runtime environment (Docker Compose, Kubernetes)
# ENV TZ=Asia/Tokyo 
CMD ["./daypassed-bot"]
