package config

import (
	"flag"
	"os"
)

// Config Структура со всеми конфигурациями сервера
type Config struct {
	Addr      string
	DB        string
	SecretKey string
}

// MustLoadConfig загрузка конфигурации
func MustLoadConfig() *Config {
	var config Config
	flag.StringVar(&config.Addr, "a", "localhost:8080", "Адрес порт сервера")
	flag.StringVar(
		&config.DB,
		"d",
		"user=postgres password=postgres host=localhost port=5432 database=postgres sslmode=disable",
		"DSN строка для соединения с базой данных",
	)
	flag.StringVar(&config.SecretKey, "k", "22gwiT5#eQxdh89OJZM-9af=LDB^EIJsW7Bbv90s1L^U.O7jNu8OrEhWLM.zJFUk", "Секретный ключ для хеширования пароля")
	flag.Parse()

	envAddr := os.Getenv("RUN_ADDRESS")
	if envAddr != "" {
		config.Addr = envAddr
	}
	envDB := os.Getenv("DATABASE_URI")
	if envDB != "" {
		config.DB = envDB
	}

	return &config
}
