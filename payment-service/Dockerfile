# --- 1. Base Stage (เตรียมของ) ---
FROM golang:1.26-alpine3.23 AS base
WORKDIR /app

# ลง Git และ Air (สำหรับ Hot Reload)
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest

# Copy dependency มาโหลดเก็บไว้ (Cache Layer)
COPY go.mod go.sum ./
RUN go mod download

# --- 2. Dev Stage (สำหรับ Docker Compose ในเครื่อง) ---
FROM base AS dev
# Copy โค้ดทั้งหมดเข้าไป
COPY . .
# สั่งรัน Air เป็นหลัก
CMD ["go" , "run" , "cmd/main.go"]

# --- 3. Builder Stage (เตรียม Build สำหรับ Prod) ---
FROM base AS builder
COPY . .
# Build เป็น Binary (ระบุ path cmd/main.go ให้ถูก)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server cmd/main.go

# --- 4. Production Stage (ตัวจริง ไฟล์เล็ก) ---
FROM alpine:3.23.2 AS prod
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Bangkok
WORKDIR /app

# Copy เฉพาะไฟล์ Binary จาก Builder มา
COPY --from=builder /app/server .

CMD ["./server"]

