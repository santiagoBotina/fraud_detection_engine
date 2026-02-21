FROM golang:1.26-bookworm

WORKDIR /app

RUN apt-get -o Acquire::https::Verify-Peer=false update -y && \
    apt-get -o Acquire::https::Verify-Peer=false install -y git build-essential

ENV GONOSUMDB=*
ENV GOINSECURE=*
ENV GONOPROXY=*
ENV GONOSUMCHECK=*
ENV GOPROXY=off

RUN git config --global http.sslVerify false

RUN go install github.com/air-verse/air@latest
ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

COPY . .

EXPOSE 3000

ENTRYPOINT ["air", "-c", ".air.toml"]
