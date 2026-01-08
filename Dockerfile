# Этап 1: Сборка
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты: компилятор, musl-dev и sqlite-dev (для заголовков sqlite3.h)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем код и собираем с флагом для совместимости
COPY *.go ./
RUN CGO_ENABLED=1 GOOS=linux \
    CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
    go build -o /mindforge .

# Этап 2: Финальный минимальный образ
FROM alpine:latest

# Runtime зависимости
RUN apk add --no-cache sqlite-libs ca-certificates

WORKDIR /root/
COPY --from=builder /mindforge .
VOLUME ["/root/data"]
EXPOSE 8080
ENTRYPOINT ["./mindforge"]