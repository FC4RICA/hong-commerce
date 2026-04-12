FROM golang:1.26-alpine3.23 AS base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


FROM base AS dev

RUN go install github.com/air-verse/air@v1.61.5
COPY . .
CMD ["air", "-c", ".air.toml"]


FROM base AS builder

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /gateway \
    ./gateway


FROM scratch AS prod

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]