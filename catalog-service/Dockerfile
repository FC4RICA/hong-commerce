FROM golang:1.26-alpine3.23 AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.5

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]


FROM golang:1.26-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/catalog-service ./cmd/main.go


FROM alpine:3.20 AS prod

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/catalog-service /bin/catalog-service

EXPOSE 8082
ENTRYPOINT ["/bin/catalog-service"]