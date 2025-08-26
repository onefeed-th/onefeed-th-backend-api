# ===========================
# Stage 1: Build
# ===========================
FROM golang:1.24.5 AS builder

WORKDIR /app

# ติดตั้ง dependencies ก่อนเพื่อ cache layer
COPY go.mod go.sum ./
RUN go mod download

# copy source code ทั้งหมด
COPY . .

# Build แบบ static binary เพื่อลด dependency ใน runtime
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# ===========================
# Stage 2: Runtime
# ===========================
FROM gcr.io/distroless/base-debian12 AS final

WORKDIR /app

# คัดลอก binary ที่ build แล้วมา
COPY --from=builder /app/main .

# กำหนด user ที่ไม่ใช่ root เพื่อความปลอดภัย
USER nonroot:nonroot

# คำสั่งเริ่มต้น
ENTRYPOINT ["./main"]
