FROM golang:1.23.7-alpine AS builder

WORKDIR /go/src/eventual

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o eventual .

FROM alpine:latest

RUN apk add --no-cache dumb-init

WORKDIR /app

COPY --from=builder /go/src/eventual/eventual .

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["./eventual"]

VOLUME [ "./data" ]
