package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`
	JWT struct {
		Secret      string `mapstructure:"secret"`
		ExpiryHours int    `mapstructure:"expiry_hours"`
	} `mapstructure:"jwt"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Ошибка чтения конфигурации")
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error().Err(err).Msg("Ошибка разбора конфигурации")
		return nil, err
	}

	log.Info().Msg("Конфигурация успешно загружена")
	return &cfg, nil
}
