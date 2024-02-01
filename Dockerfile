FROM node:16-alpine as clientbuilder

WORKDIR /build

ADD frontend /build/
RUN npm install
RUN npm run build

FROM golang:1.17-alpine as builder

WORKDIR /build

RUN apk update \
  && apk add --no-cache git \
  && apk add --no-cache ca-certificates \
  && apk add --update gcc musl-dev \
  && update-ca-certificates

COPY backend/go.mod .
COPY backend/go.sum .
RUN go mod download

ADD backend /build/
# RUN go test -v ./...
COPY --from=clientbuilder /build/dist /build/dist/

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o moneropot

FROM alpine:3.14 as app

WORKDIR /app
VOLUME [ "/data" ]
COPY --from=builder /build/moneropot .

CMD ["/app/moneropot", "--bind", "0.0.0.0:5000"]