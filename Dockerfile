FROM golang:1.23.9-alpine AS builder

# Install required build dependencies for CGO
RUN apk add --no-cache dumb-init gcc musl-dev

WORKDIR /go/src/eventual

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o eventual .

FROM alpine:latest

WORKDIR /app

# Copy dumb-init and the compiled binary
COPY --from=builder /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=builder /go/src/eventual/eventual /usr/bin/eventual
COPY --from=builder /go/src/eventual/start.sh /app/start.sh

# Create data directory if it doesn't exist
RUN mkdir -p /app/data

RUN chmod +x /app/start.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/app/start.sh"]

VOLUME ["/app/data"]
