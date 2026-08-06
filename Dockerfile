### Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /leasewebassignment ./cmd

### Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /leasewebassignment /usr/local/bin/leasewebassignment
COPY config.yaml ./config.yaml
COPY data ./data
COPY web ./web
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/leasewebassignment", "-config=config.yaml"]
