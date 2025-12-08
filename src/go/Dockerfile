# Stage 1: Build
FROM golang:1.25.5-alpine AS builder

# Basic dependencies
RUN apk add --no-cache git make

WORKDIR /service

COPY . .
RUN go mod download

RUN go build -o service ./cmd/consumer

# Stage 2: Final image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=builder /service/service .

CMD ["/service"]
