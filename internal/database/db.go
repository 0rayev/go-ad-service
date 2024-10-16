package database

import (
	"ad_service/internal/config"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(cfg config.MySQLConfig) (*sql.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	migrationsPath := filepath.Join("internal", "database", "migrations", "init.sql")
	if err := runInitSqlScript(db, migrationsPath); err != nil {
		return nil, err
	}

	return db, nil
}

func runInitSqlScript(db *sql.DB, filePath string) error {
	// Check if the file exists and read the SQL file
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Execute the SQL commands
	_, err = db.Exec(string(sqlBytes))
	return err
}
