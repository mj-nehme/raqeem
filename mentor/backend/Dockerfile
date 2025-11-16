# Dockerfile
# Using explicit registry reference for official Go image
FROM docker.io/library/golang:1.25-alpine

# Version and metadata
LABEL version="0.1.0"
LABEL description="Raqeem Mentor Backend - Go device management and analytics service"

WORKDIR /app

COPY src/go.mod ./
COPY src/go.sum ./
RUN go mod download

COPY ./src .

RUN go build -o main .

CMD ["./main"]
