# Build stage
FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kommit .

# Final stage
FROM alpine:3.19

# Install git (required for git operations)
RUN apk --no-cache add git

# Set the working directory
WORKDIR /workdir

# Copy the binary from builder
COPY --from=builder /app/kommit /usr/local/bin/kommit

# Set the entrypoint to the kommit binary
ENTRYPOINT ["kommit"]
