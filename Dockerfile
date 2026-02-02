# ---------- Build stage ----------
# docker build -t graph-editor .
# To keep graph edits after container restarts:
# docker run \
#  -p 8080:8080 \
#  -v $(pwd)/cmd/graph-data.json:/app/graph-data.json \
#  graph-editor
    
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build the binary from cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o server ./cmd

# ---------- Runtime stage ----------
# Since .json file is /app/*.json, the golan code needs to read just the 
# *.json file without putting cmd/ directory in the path.
FROM alpine:3.19

# Create non-root user
RUN adduser -D appuser

WORKDIR /app

# Copy compiled binary
COPY --from=builder /app/server /app/server

# Copy static frontend
COPY --from=builder /app/static /app/static

# Copy graph data file
COPY --from=builder /app/cmd/graph-data.json /app/graph-data.json

# Set ownership
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./server"]
