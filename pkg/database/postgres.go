package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresDB は新しいPostgreSQLデータベース接続を作成します
func NewPostgresDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// データベース接続のテスト
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Successfully connected to database")

	return db, nil
}

// TestConnection はデータベース接続が正常に機能しているかテストします
func TestConnection(db *sql.DB) error {
	// 簡単なクエリを実行
	var one int
	err := db.QueryRow("SELECT 1").Scan(&one)
	if err != nil {
		return err
	}
	if one != 1 {
		return fmt.Errorf("expected 1, got %d", one)
	}
	return nil
} 