package config

import "os"

// Config содержит основные настройки
type Config struct {
	// ClickHouse
	ClickhouseDSN string

	// Миграции
	MigrationsPath string

	// Логирование
	LogFile string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		ClickhouseDSN:  getEnv("CLICKHOUSE_DSN", "clickhouse://default:@clickhouse:9000/schedule"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "./migrations"),
		LogFile:        getEnv("LOG_FILE", "./logs/migrate.log"),
	}
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
