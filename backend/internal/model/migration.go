package model

import "time"

// Migration запись о миграции в базе данных
type Migration struct {
	ID         uint32    `json:"id"`
	Name       string    `json:"name"`
	Batch      uint32    `json:"batch"`
	ExecutedAt time.Time `json:"executed_at"`
}

// MigrationFile файл миграции на диске
type MigrationFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
