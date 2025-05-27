package handlers

import (
	"budget-app/internal/db"
	"budget-app/internal/models"
	"budget-app/pkg/config"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().
			Err(err).
			Interface("request_body", req).
			Msg("Ошибка валидации запроса на регистрацию")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат данных: " + err.Error()})
		return
	}

	// Проверка минимальной длины пароля
	if len(req.Password) < 6 {
		log.Warn().
			Str("username", req.Username).
			Msg("Пароль слишком короткий")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен содержать минимум 6 символов"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка хеширования пароля")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при хешировании пароля"})
		return
	}

	// Сохранение пользователя
	_, err = db.DB.Exec("INSERT INTO users (username, password, email) VALUES ($1, $2, $3)",
		req.Username, string(hashedPassword), req.Email)
	if err != nil {
		log.Warn().
			Err(err).
			Str("username", req.Username).
			Str("email", req.Email).
			Msg("Ошибка сохранения пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь с таким именем или email уже существует"})
		return
	}

	log.Info().
		Str("username", req.Username).
		Str("email", req.Email).
		Msg("Пользователь успешно зарегистрирован")
	c.JSON(http.StatusOK, gin.H{"message": "Регистрация успешна"})
}

func Login(c *gin.Context) {
	cfg, _ := c.MustGet("config").(*config.Config)
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Ошибка валидации запроса на вход")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат данных: " + err.Error()})
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, username, password FROM users WHERE username = $1", req.Username).
		Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		log.Warn().Err(err).Str("username", req.Username).Msg("Пользователь не найден")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверное имя пользователя или пароль"})
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Warn().Str("username", req.Username).Msg("Неверный пароль")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверное имя пользователя или пароль"})
		return
	}

	// Создание JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * time.Duration(cfg.JWT.ExpiryHours)).Unix(),
	})
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		log.Error().Err(err).Msg("Ошибка создания токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при создании токена"})
		return
	}

	log.Info().Str("username", req.Username).Msg("Успешный вход")
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Ошибка валидации запроса на сброс пароля")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат данных: " + err.Error()})
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, email FROM users WHERE email = $1", req.Email).
		Scan(&user.ID, &user.Email)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("Пользователь не найден")
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь с таким email не найден"})
		return
	}

	// Здесь должен быть код для отправки письма с ссылкой на сброс пароля
	log.Info().Str("email", req.Email).Msg("Запрос на сброс пароля")
	c.JSON(http.StatusOK, gin.H{"message": "Инструкции по сбросу пароля отправлены на email"})
}

func AddTransaction(c *gin.Context) {
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Ошибка валидации запроса на добавление транзакции")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат данных: " + err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		log.Warn().Err(err).Str("date", req.Date).Msg("Неверный формат даты")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты (ожидается YYYY-MM-DD)"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO transactions (user_id, date, name, category, amount) VALUES ($1, $2, $3, $4, $5)",
		userID, date.Format("2006-01-02"), req.Name, req.Category, req.Amount)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка сохранения транзакции")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения транзакции"})
		return
	}

	log.Info().Int("user_id", userID.(int)).Str("category", req.Category).Float64("amount", req.Amount).Msg("Транзакция добавлена")
	c.JSON(http.StatusOK, gin.H{"message": "Транзакция добавлена"})
}

func GetTransactionsByDay(c *gin.Context) {
	userID, _ := c.Get("user_id")
	date := c.Param("date")

	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Warn().Err(err).Str("date", date).Msg("Неверный формат даты")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты (ожидается YYYY-MM-DD)"})
		return
	}

	rows, err := db.DB.Query("SELECT id, date, name, category, amount FROM transactions WHERE user_id = $1 AND date = $2",
		userID, date)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка получения транзакций")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var dateStr string
		if err := rows.Scan(&t.ID, &dateStr, &t.Name, &t.Category, &t.Amount); err != nil {
			log.Error().Err(err).Msg("Ошибка чтения данных транзакции")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}
		t.Date, _ = time.Parse("2006-01-02", dateStr)
		transactions = append(transactions, t)
	}

	log.Info().Int("user_id", userID.(int)).Str("date", date).Int("count", len(transactions)).Msg("Получены транзакции за день")
	c.JSON(http.StatusOK, transactions)
}

func GetTransactionsByPeriod(c *gin.Context) {
	userID, _ := c.Get("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Warn().Err(err).Str("start_date", startDate).Msg("Неверный формат даты")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат start_date (ожидается YYYY-MM-DD)"})
		return
	}
	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Warn().Err(err).Str("end_date", endDate).Msg("Неверный формат даты")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат end_date (ожидается YYYY-MM-DD)"})
		return
	}

	rows, err := db.DB.Query("SELECT id, date, name, category, amount FROM transactions WHERE user_id = $1 AND date BETWEEN $2 AND $3",
		userID, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка получения транзакций за период")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var dateStr string
		if err := rows.Scan(&t.ID, &dateStr, &t.Name, &t.Category, &t.Amount); err != nil {
			log.Error().Err(err).Msg("Ошибка чтения данных транзакции")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}
		t.Date, _ = time.Parse("2006-01-02", dateStr)
		transactions = append(transactions, t)
	}

	log.Info().Int("user_id", userID.(int)).Str("start_date", startDate).Str("end_date", endDate).Int("count", len(transactions)).Msg("Получены транзакции за период")
	c.JSON(http.StatusOK, transactions)
}
