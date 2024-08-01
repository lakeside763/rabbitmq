# Build stage
FROM golang:1.22-alpine AS build

# Install git and other dependencies
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o consumer consumer.go

# Final stage
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=build /src/consumer .

# Command to run the executable
CMD ["./consumer"]

