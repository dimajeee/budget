package main

import (
	"budget-app/internal/server"
	"budget-app/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	// Настройка логгера
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Создание директории logs/, если она не существует
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal().Err(err).Msg("Не удалось создать директорию логов")
	}

	logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось открыть файл логов")
	}
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось загрузить конфигурацию")
	}

	// Запуск сервера
	if err := server.Run(cfg); err != nil {
		log.Fatal().Err(err).Msg("Ошибка запуска сервера")
	}
}
