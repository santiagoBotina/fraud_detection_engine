FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY combined-ca-bundle.pem /usr/local/share/ca-certificates/combined-ca-bundle.crt
RUN update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/service

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/server /app/server

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

EXPOSE ${DECISION_APP_PORT}

ENTRYPOINT ["/app/server"]
