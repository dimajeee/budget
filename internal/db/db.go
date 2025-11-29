package db

import (
	"budget-app/pkg/config"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка проверки подключения к БД: %v", err)
	}

	// Создание таблиц
	err = createTables()
	if err != nil {
		return fmt.Errorf("ошибка создания таблиц: %v", err)
	}

	log.Info().Msg("База данных успешно инициализирована")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Info().Msg("Соединение с базой данных закрыто")
	}
}

func createTables() error {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL
	);`
	transactionTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		date DATE NOT NULL,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(50) NOT NULL,
		amount DOUBLE PRECISION NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	_, err := DB.Exec(userTable)
	if err != nil {
		return err
	}
	_, err = DB.Exec(transactionTable)
	if err != nil {
		return err
	}
	return nil
}
