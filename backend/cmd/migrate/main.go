package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"university-schedule/internal/config"
	"university-schedule/internal/logger"
	"university-schedule/internal/migrate"
)

var (
	cfg *config.Config
)

func main() {
	cfg = config.Load()
	logger := logger.SetupLogger(cfg.LogFile)

	var rootCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Инструмент миграций для ClickHouse",
	}

	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Printf("Ошибка: %v", err)
		os.Exit(1)
	}
}

// upCmd команда для применения миграций
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Применить все неприменённые миграции",
	Run: func(cmd *cobra.Command, args []string) {
		migrator, err := migrate.New(cfg)
		if err != nil {
			fmt.Printf("Ошибка создания мигратора: %v\n", err)
			os.Exit(1)
		}
		defer migrator.Close()

		if err := migrator.Up(); err != nil {
			fmt.Printf("Ошибка применения миграций: %v\n", err)
			os.Exit(1)
		}
	},
}

// downCmd команда для отката миграций
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Откатить последний batch миграций",
	Run: func(cmd *cobra.Command, args []string) {
		migrator, err := migrate.New(cfg)
		if err != nil {
			fmt.Printf("Ошибка создания мигратора: %v\n", err)
			os.Exit(1)
		}
		defer migrator.Close()

		if err := migrator.Down(); err != nil {
			fmt.Printf("Ошибка отката миграций: %v\n", err)
			os.Exit(1)
		}
	},
}

// statusCmd команда для показа статуса миграций
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Показать статус миграций",
	Run: func(cmd *cobra.Command, args []string) {
		migrator, err := migrate.New(cfg)
		if err != nil {
			fmt.Printf("Ошибка создания мигратора: %v\n", err)
			os.Exit(1)
		}
		defer migrator.Close()

		if err := migrator.Status(); err != nil {
			fmt.Printf("Ошибка получения статуса: %v\n", err)
			os.Exit(1)
		}
	},
}

// createCmd команда для создания новой миграции
var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Создать новую миграцию",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s", timestamp, name)

		if err := os.MkdirAll(cfg.MigrationsPath, 0755); err != nil {
			fmt.Printf("Не удалось создать папку миграций: %v\n", err)
			os.Exit(1)
		}

		upPath := filepath.Join(cfg.MigrationsPath, filename+".up.sql")
		upContent := fmt.Sprintf("-- Миграция: %s\n-- Создано: %s\n\n-- Добавьте ваш SQL код здесь\n", name, time.Now().Format("2006-01-02 15:04:05"))

		if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
			fmt.Printf("Не удалось создать up файл: %v\n", err)
			os.Exit(1)
		}

		downPath := filepath.Join(cfg.MigrationsPath, filename+".down.sql")
		downContent := fmt.Sprintf("-- Откат: %s\n-- Создано: %s\n\n-- Добавьте ваш SQL код отката здесь\n", name, time.Now().Format("2006-01-02 15:04:05"))

		if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
			fmt.Printf("Не удалось создать down файл: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Создана миграция: %s\n", filename)
		fmt.Printf("Up файл: %s\n", upPath)
		fmt.Printf("Down файл: %s\n", downPath)
	},
}
