package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"university-schedule/internal/config"
	"university-schedule/internal/model"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Migrator простой класс для работы с миграциями
type Migrator struct {
	config *config.Config
	db     *sql.DB
}

// New создаёт новый мигратор
func New(cfg *config.Config) (*Migrator, error) {
	db, err := sql.Open("clickhouse", cfg.ClickhouseDSN)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к ClickHouse: %v", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось проверить соединение: %v", err)
	}

	return &Migrator{
		config: cfg,
		db:     db,
	}, nil
}

// Close закрывает соединение
func (m *Migrator) Close() error {
	return m.db.Close()
}

// Up применяет все неприменённые миграции
func (m *Migrator) Up() error {
	log.Println("Применяем миграции...")

	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %v", err)
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("не удалось получить файлы миграций: %v", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("не удалось получить применённые миграции: %v", err)
	}

	pending := m.getPendingMigrations(files, applied)

	if len(pending) == 0 {
		log.Println("Нет новых миграций для применения")
		return nil
	}

	batch, err := m.getNextBatch()
	if err != nil {
		return fmt.Errorf("не удалось получить номер batch: %v", err)
	}

	for _, file := range pending {
		if err := m.applyMigration(file, batch); err != nil {
			return fmt.Errorf("не удалось применить миграцию %s: %v", file.Name, err)
		}
		log.Printf("Применена миграция: %s", file.Name)
	}

	log.Printf("Все миграции применены успешно")
	return nil
}

// Down откатывает последний batch миграций
func (m *Migrator) Down() error {
	log.Println("Откатываем миграции...")

	lastBatch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("не удалось получить последний batch: %v", err)
	}

	if lastBatch == 0 {
		log.Println("Нет миграций для отката")
		return nil
	}

	migrations, err := m.getMigrationsByBatch(lastBatch)
	if err != nil {
		return fmt.Errorf("не удалось получить миграции batch: %v", err)
	}

	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if err := m.rollbackMigration(migration); err != nil {
			return fmt.Errorf("не удалось откатить миграцию %s: %v", migration.Name, err)
		}
		log.Printf("Откачена миграция: %s", migration.Name)
	}

	log.Printf("Все миграции откачены успешно")
	return nil
}

// Status показывает статус миграций
func (m *Migrator) Status() error {
	log.Println("Статус миграций:")

	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %v", err)
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("не удалось получить файлы миграций: %v", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("не удалось получить применённые миграции: %v", err)
	}

	pending := m.getPendingMigrations(files, applied)

	lastBatch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("не удалось получить последний batch: %v", err)
	}

	fmt.Printf("\nСтатус миграций\n")
	fmt.Printf("==================\n")
	fmt.Printf("Последний batch: %d\n", lastBatch)
	fmt.Printf("Применено: %d миграций\n", len(applied))
	fmt.Printf("Ожидает: %d миграций\n", len(pending))

	if len(applied) > 0 {
		fmt.Printf("\nПрименённые миграции:\n")
		for _, migration := range applied {
			fmt.Printf("  [Batch %d] %s (%s)\n",
				migration.Batch,
				migration.Name,
				migration.ExecutedAt.Format("2006-01-02 15:04:05"))
		}
	}

	if len(pending) > 0 {
		fmt.Printf("\nОжидающие миграции:\n")
		for _, file := range pending {
			fmt.Printf("  %s\n", file.Name)
		}
	}

	fmt.Println()
	return nil
}

// createMigrationsTable создаёт таблицу миграций
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id UInt32,
			migration String,
			batch UInt32,
			executed_at DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY id
	`
	_, err := m.db.Exec(query)
	return err
}

func (m *Migrator) getMigrationFiles() ([]model.MigrationFile, error) {
	var files []model.MigrationFile

	err := filepath.Walk(m.config.MigrationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".up.sql") {
			name := strings.TrimSuffix(filepath.Base(path), ".up.sql")
			files = append(files, model.MigrationFile{
				Name: name,
				Path: path,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// getAppliedMigrations получает список применённых миграций
func (m *Migrator) getAppliedMigrations() ([]model.Migration, error) {
	query := "SELECT id, migration, batch, executed_at FROM migrations ORDER BY id"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []model.Migration
	for rows.Next() {
		var m model.Migration
		err := rows.Scan(&m.ID, &m.Name, &m.Batch, &m.ExecutedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

// getPendingMigrations определяет неприменённые миграции
func (m *Migrator) getPendingMigrations(files []model.MigrationFile, applied []model.Migration) []model.MigrationFile {
	appliedMap := make(map[string]bool)
	for _, migration := range applied {
		appliedMap[migration.Name] = true
	}

	var pending []model.MigrationFile
	for _, file := range files {
		if !appliedMap[file.Name] {
			pending = append(pending, file)
		}
	}

	return pending
}

// getNextBatch получает следующий номер batch
func (m *Migrator) getNextBatch() (uint32, error) {
	query := "SELECT COALESCE(MAX(batch), 0) + 1 FROM migrations"
	var batch uint32
	err := m.db.QueryRow(query).Scan(&batch)
	return batch, err
}

// getLastBatch получает номер последнего batch
func (m *Migrator) getLastBatch() (uint32, error) {
	query := "SELECT COALESCE(MAX(batch), 0) FROM migrations"
	var batch uint32
	err := m.db.QueryRow(query).Scan(&batch)
	return batch, err
}

// getMigrationsByBatch получает миграции по номеру batch
func (m *Migrator) getMigrationsByBatch(batch uint32) ([]model.Migration, error) {
	query := "SELECT id, migration, batch, executed_at FROM migrations WHERE batch = ? ORDER BY id"
	rows, err := m.db.Query(query, batch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []model.Migration
	for rows.Next() {
		var m model.Migration
		err := rows.Scan(&m.ID, &m.Name, &m.Batch, &m.ExecutedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

// applyMigration применяет одну миграцию
func (m *Migrator) applyMigration(file model.MigrationFile, batch uint32) error {
	// Читаем содержимое файла
	content, err := ioutil.ReadFile(file.Path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %v", err)
	}

	// Выполняем SQL
	if _, err := m.db.Exec(string(content)); err != nil {
		return fmt.Errorf("не удалось выполнить SQL: %v", err)
	}

	// Записываем информацию о миграции
	query := "INSERT INTO migrations (id, migration, batch) VALUES (?, ?, ?)"

	// Получаем следующий ID
	id, err := m.getNextID()
	if err != nil {
		return fmt.Errorf("не удалось получить следующий ID: %v", err)
	}

	_, err = m.db.Exec(query, id, file.Name, batch)
	return err
}

// rollbackMigration откатывает одну миграцию
func (m *Migrator) rollbackMigration(migration model.Migration) error {
	// Читаем содержимое down файла
	downPath := filepath.Join(m.config.MigrationsPath, migration.Name+".down.sql")
	content, err := ioutil.ReadFile(downPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл отката: %v", err)
	}

	// Выполняем SQL отката
	if _, err := m.db.Exec(string(content)); err != nil {
		return fmt.Errorf("не удалось выполнить SQL отката: %v", err)
	}

	// Удаляем запись о миграции
	query := "DELETE FROM migrations WHERE id = ?"
	_, err = m.db.Exec(query, migration.ID)
	return err
}

// getNextID получает следующий ID для миграции
func (m *Migrator) getNextID() (uint32, error) {
	query := "SELECT COALESCE(MAX(id), 0) + 1 FROM migrations"
	var id uint32
	err := m.db.QueryRow(query).Scan(&id)
	return id, err
}
