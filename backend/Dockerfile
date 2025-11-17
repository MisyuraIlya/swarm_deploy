FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache curl && adduser -D -g '' appuser

USER appuser

COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1

CMD ["/app/server"]
