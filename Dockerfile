FROM golang:1.25.1

# Set working directory
WORKDIR /app

# Disable CGO (simplifies builds)
ENV CGO_ENABLED=0
# Use Go proxy to reliably fetch modules
ENV GOPROXY=https://proxy.golang.org,direct
# Set port for Fly.io
ENV PORT=8080

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the Go app
RUN go build -o sales-log .

# Expose port (must match app)
EXPOSE 8080

# Run the app
CMD ["./sales-log"]
