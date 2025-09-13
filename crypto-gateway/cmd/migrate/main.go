package main

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/appmetrics"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	Name     string `json:"name"`
	Filepath string `json:"filepath"`
}

func main() {
	config.LoadConfig()

	ctx := context.Background()

	if err := EnsureDatabaseExists(ctx); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "Could not ensure DB exists", err)
		log.Fatalf("Could not ensure DB: %v", err)
	}

	if err := MakeMigrations(); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "Migration failed", err)
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully.")
}

func MakeMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, config.Database_Url)
	if err != nil {
		return fmt.Errorf("failed to connect to target DB: %w", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping target DB: %w", err)
	}

	var schemaExists bool
	err = dbpool.QueryRow(ctx, `SELECT EXISTS (
        SELECT 1 FROM information_schema.schemata WHERE schema_name = 'public'
    )`).Scan(&schemaExists)

	if err != nil || !schemaExists {
		return fmt.Errorf("schema 'public' not found or not ready")
	}

	initialized, err := modelsInitialized(ctx, dbpool)
	if err != nil {
		return fmt.Errorf("failed to check model initialization: %w", err)
	}

	if initialized {
		log.Println("Models already initialized. Skipping.")
		return nil
	}

	// основные миграции
	if err := applySQLFiles(ctx, dbpool); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// кастомные
	if err := applyModuleSQLFiles(ctx, dbpool); err != nil {
		return fmt.Errorf("failed to apply module migrations: %w", err)
	}

	return nil
}

func modelsInitialized(ctx context.Context, dbpool *pgxpool.Pool) (bool, error) {
	var count int
	err := dbpool.QueryRow(ctx, `
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
    `).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyModuleSQLFiles(ctx context.Context, dbpool *pgxpool.Pool) error {
	data, err := os.ReadFile(filepath.Join("cmd", "migrate", "modules.json"))
	if err != nil {
		return fmt.Errorf("failed to read modules.json: %w", err)
	}

	var modules []Module
	if err := json.Unmarshal(data, &modules); err != nil {
		return fmt.Errorf("failed to parse modules.json: %w", err)
	}

	for _, m := range modules {
		absPath := filepath.Join("internal", "modules", m.Filepath)
		log.Printf("Applying migration for module %s from %s", m.Name, absPath)
		if err := applySQLFile(ctx, dbpool, absPath); err != nil {
			return fmt.Errorf("error applying module %s (%s): %w", m.Name, absPath, err)
		}
	}

	return nil
}

func applySQLFiles(ctx context.Context, dbpool *pgxpool.Pool) error {
	files, err := filepath.Glob(filepath.Join("internal", "migrations", "*.sql"))
	fmt.Println(files)
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		if err := applySQLFile(ctx, dbpool, file); err != nil {
			return fmt.Errorf("error applying file %s: %w", file, err)
		}
	}
	return nil
}

func applySQLFile(ctx context.Context, dbpool *pgxpool.Pool, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	parts := strings.Split(string(content), ";")

	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}

		_, err := dbpool.Exec(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %s\nerror: %w", stmt, err)
		}
	}

	return nil
}

func EnsureDatabaseExists(ctx context.Context) error {
	adminUrl := "postgresql://postgres:root@localhost:5432/postgres?sslmode=disable"
	if config.IsInDocker {
		adminUrl = "postgresql://postgres:root@postgres:5432/postgres?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, adminUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.Database_Name).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check db existence: %w", err)
	}

	if exists {
		return nil
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, config.Database_Name))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database %s created.", config.Database_Name)
	return nil
}
