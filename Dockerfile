# first stage: build the app
# golang images
FROM golang:1.24-alpine as builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /supreme-socket-server/app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o build/server/server ./cmd/server

# second stage: build the final image
FROM alpine:latest

# Set the timezone
ENV TZ=Asia/Taipei

# Install tzdata
RUN apk add --no-cache tzdata

# Install ca-certificates
RUN apk add --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /supreme-socket-server/app/build/server/server .

# Expose port 8080 to the outside world
EXPOSE 443

# Command to run the executable
CMD ["./server"]