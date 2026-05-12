package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3" 
)

// Config — основная структура конфигурации приложения
type Config struct {
	Server   ServerConfig   `yaml:"server"`   // Настройки сервера
	Database DatabaseConfig `yaml:"database"` // Настройки базы данных
}

// ServerConfig — параметры веб-сервера
type ServerConfig struct {
	Addr string `yaml:"addr"` // Адрес и порт сервера (например, ":8080")
	Mode string `yaml:"mode"` // Режим работы (debug/release)
}

// DatabaseConfig — параметры подключения к БД
type DatabaseConfig struct {
	Host     string `yaml:"host"`     // Хост БД
	Port     int    `yaml:"port"`     // Порт БД
	User     string `yaml:"user"`     // Пользователь БД
	Password string `yaml:"password"` // Пароль пользователя
	DBName   string `yaml:"dbname"`   // Название базы данных
	Migrate  bool   `yaml:"migrate"`  // Флаг запуска миграций
}

// Формируем строку подключения (DSN) для PostgreSQL
func (d DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// Загружаем и парсим конфигурационный файл
// 1. Определяем путь к файлу (из переменной окружения или по умолчанию)
// 2. Читаем файл
// 3. Парсим YAML в структуру Config
// 4. Устанавливаем значение по умолчанию для Server.Addr, если оно не задано
// Возвращаем конфигурацию и ошибку (если есть)
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml" // Путь по умолчанию
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Значение по умолчанию для адреса сервера
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	return &cfg, nil
}
