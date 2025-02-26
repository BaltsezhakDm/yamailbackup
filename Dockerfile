FROM golang:1.24-alpine

WORKDIR /app

# Устанавливаем crond
RUN apk add --no-cache tzdata curl bash busybox-extras

# Копируем приложение
COPY . .

RUN go build -o app cmd/main.go

RUN chmod +x /app/app

# Копируем crontab и устанавливаем его
COPY crontab /etc/crontabs/root

# Запускаем crond в foreground с логированием
CMD ["crond", "-f", "-l", "8"]
