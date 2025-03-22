FROM golang:1.20-alpine

WORKDIR /app

# Copy go.mod
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY main.go ./

# Build the application
RUN go build -o proxy

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./proxy"]