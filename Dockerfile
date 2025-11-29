# Базовый образ для сборки
FROM golang:1.24-alpine AS builder

# Установка рабочей директории
WORKDIR /app

# Копирование go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod tidy

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o budget-app cmd/budget-app/main.go

# Финальный образ
FROM alpine:latest

# Установка рабочей директории
WORKDIR /app

# Копирование собранного бинарного файла
COPY --from=builder /app/budget-app .
COPY config/config.yaml ./config/config.yaml

# Установка переменной окружения для отключения буферизации логов
ENV GIN_MODE=release

# Открытие порта
EXPOSE 8080

# Команда для запуска приложения
CMD ["./budget-app"]