FROM golang:1.24-alpine AS base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


FROM base AS dev

RUN go install github.com/air-verse/air@latest
COPY . .
CMD ["air", "-c", "services/gateway/.air.toml"]


FROM base AS builder

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /gateway \
    ./services/gateway


FROM scratch AS prod

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]