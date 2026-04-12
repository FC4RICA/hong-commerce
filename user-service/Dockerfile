FROM golang:1.26-alpine3.23 AS dev

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.5

COPY go.mod go.sum ./
RUN go mod download

# source is mounted at runtime via volume
CMD ["air", "-c", ".air.toml"]


FROM golang:1.26-alpine3.23 AS builder

WORKDIR /app

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/user-service ./cmd/main.go


FROM alpine:3.20 AS prod

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/user-service /bin/user-service

EXPOSE 8081
ENTRYPOINT ["/bin/user-service"]