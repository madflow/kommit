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

# Install bash and git (required for git operations)
RUN apk --no-cache add bash git

# Set the working directory
WORKDIR /workdir

# Copy the binary from builder
COPY --from=builder /app/kommit /usr/local/bin/kommit

# Set Git configuration through environment variables with defaults
ENV GIT_USER_NAME="Kommit User"
ENV GIT_USER_EMAIL="kommit@example.com"

# Create a script to configure Git and run kommit
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'git config --global user.name "$GIT_USER_NAME"' >> /entrypoint.sh && \
    echo 'git config --global user.email "$GIT_USER_EMAIL"' >> /entrypoint.sh && \
    echo 'exec "$@"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Set the entrypoint to use the script
ENTRYPOINT ["/entrypoint.sh", "kommit"]
