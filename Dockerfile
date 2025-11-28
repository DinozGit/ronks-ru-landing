# ---------- Stage 1: build ----------
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Сначала только go.mod — кешируем зависимости (если появятся)
COPY go.mod ./

# Если будут дополнительные Go-файлы/папки — они подтянутся следующим COPY
COPY . .

# Собираем статически линкованный бинарь
RUN CGO_ENABLED=0 GOOS=linux go build -o ronks-server ./server.go

# ---------- Stage 2: minimal runtime ----------
FROM alpine:3.20

# Небольшой непривилегированный пользователь
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Копируем бинарь и index.html
COPY --from=builder /app/ronks-server /app/ronks-server
COPY --from=builder /app/index.html /app/index.html

# Порт внутри контейнера
EXPOSE 8008

USER app

CMD ["/app/ronks-server"]
