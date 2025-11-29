package server

import (
	"budget-app/internal/api/handlers"
	"budget-app/internal/db"
	"budget-app/internal/middleware"
	"budget-app/pkg/config"
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Run(cfg *config.Config) error {
	// Инициализация базы данных
	if err := db.InitDB(cfg); err != nil {
		return err
	}
	defer db.CloseDB()

	// Инициализация Gin
	gin.SetMode(gin.ReleaseMode) // Устанавливаем release mode
	r := gin.Default()
	r.SetTrustedProxies(nil) // Отключаем доверие ко всем прокси

	// Middleware для логирования запросов
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		// Чтение тела запроса
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error().Err(err).Msg("Ошибка чтения тела запроса")
		}
		// Восстановление тела запроса для последующих обработчиков
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Логирование тела
		log.Info().Str("body", string(bodyBytes)).Msg("Получен запрос")

		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Msg("Обработан HTTP запрос")
	})

	// Передача конфигурации в контекст
	r.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	})

	// Роуты без аутентификации
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/password/reset", handlers.RequestPasswordReset)

	// Группа роутов с аутентификацией
	authorized := r.Group("/api", middleware.AuthMiddleware(cfg))
	{
		authorized.POST("/transactions", handlers.AddTransaction)
		authorized.GET("/transactions/day/:date", handlers.GetTransactionsByDay)
		authorized.GET("/transactions/period", handlers.GetTransactionsByPeriod)
	}

	// Запуск сервера
	log.Info().Str("port", cfg.Server.Port).Msg("Сервер запущен")
	return r.Run(cfg.Server.Port)
}
