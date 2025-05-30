# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --no-cache gcc musl-dev curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o app

# Run stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/app .

EXPOSE 8080

CMD ["./app"] 