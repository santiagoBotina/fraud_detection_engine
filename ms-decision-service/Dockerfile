FROM golang:1.26-bookworm

WORKDIR /app

# CA bundle for Netskope / corporate proxy
COPY combined-ca-bundle.pem /usr/local/share/ca-certificates/combined-ca-bundle.crt
RUN update-ca-certificates

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

RUN apt-get update -y && \
    apt-get install -y git build-essential

RUN go install github.com/air-verse/air@latest
ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

COPY . .

EXPOSE ${DECISION_APP_PORT}

ENTRYPOINT ["air", "-c", ".air.toml"]
