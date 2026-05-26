FROM golang:1.26-alpine3.23 AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.5

# We'll skip go mod download for now since go.sum might be missing
# and let air/go build handle it if needed, or we copy what we have.
COPY go.mod ./
# Note: go.sum might not exist yet if I couldn't run go mod tidy
# COPY go.sum ./

CMD ["air", "-c", ".air.toml"]


FROM golang:1.26-alpine3.23 AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/order-service ./cmd/main.go


FROM alpine:3.20 AS prod

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/order-service /bin/order-service

EXPOSE 8083
ENTRYPOINT ["/bin/order-service"]
