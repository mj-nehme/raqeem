# Dockerfile
FROM golang:1.23-alpine

# Version and metadata
LABEL version="0.0.4"
LABEL description="Raqeem Mentor Backend - Go device management and analytics service"

WORKDIR /app

COPY src/go.mod ./
COPY src/go.sum ./
RUN go mod download

COPY ./src .

RUN go build -o main .

CMD ["./main"]
