# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable modules and CGO off for static binary
ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o server .

# Runtime stage
FROM alpine:3.20

WORKDIR /app

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/server /app/server

EXPOSE 8080

ENV PORT=8080

CMD ["/app/server"]
