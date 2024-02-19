# Start from the official Golang image to build your application
FROM golang:1.21-bullseye as builder

# Set the environment to enable CGO
ENV CGO_ENABLED=1

# Install git and gcc.
# GCC is required for cgo, which is needed by go-sqlite3.
RUN apt-get update && apt-get install -y git gcc

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app with CGO enabled
RUN go build -o anniversaryAPI ./cmd

# Start a new stage from debian:buster
FROM debian:buster

# Install ca-certificates, and the libraries needed by SQLite
RUN apt-get update && apt-get install -y ca-certificates libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/anniversaryAPI .

# Expose port 8080, 8081 and 9090 to the outside world
EXPOSE 8080
EXPOSE 8081
EXPOSE 9090

# Command to run the executable
CMD ["./anniversaryAPI"]
