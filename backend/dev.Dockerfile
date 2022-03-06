FROM golang:1.17

VOLUME /app

WORKDIR /app

RUN go get github.com/githubnemo/CompileDaemon
