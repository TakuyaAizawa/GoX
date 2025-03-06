package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TakuyaAizawa/gox/pkg/database"
	"github.com/joho/godotenv"
)

// User はユーザー情報を表す構造体
type User struct {
	ID        int
	Username  string
	Email     string
	CreatedAt time.Time
}

func main() {
	// メモ: .envファイルがUTF-16エンコーディングで保存されています
	// 以下のようにコマンドを実行して変換できます:
	// 1. PowerShellで:
	//    Get-Content -Path .\.env -Encoding Unicode | Set-Content -Path .\.env -Encoding utf8
	// 2. または直接メモ帳などでファイルを開き、UTF-8形式で保存し直す
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("警告: カレントディレクトリの取得に失敗しました: %v", err)
	}
	log.Printf("現在の作業ディレクトリ: %s", dir)

	// 正確なパスで.envファイルを探す (C:\share\git\GoX\.env)
	rootDir := filepath.Dir(filepath.Dir(dir)) // カレントディレクトリの親の親 (プロジェクトルート)
	
	// .envファイルのパスを明示的に指定して環境変数を読み込む
	envPaths := []string{
		".env",                            // カレントディレクトリ
		filepath.Join(dir, ".env"),        // 絶対パス
		filepath.Join(dir, "../.env"),     // 親ディレクトリ
		filepath.Join(rootDir, ".env"),    // プロジェクトルート
	}

	// 各パスをログに出力（デバッグ用）
	log.Println("探索する.envファイルのパス:")
	for i, path := range envPaths {
		log.Printf("  パス %d: %s", i+1, path)
	}
	
	envLoaded := false
	var loadErr error
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("ファイルが存在します: %s", path)
			if err := godotenv.Load(path); err == nil {
				log.Printf("情報: .envファイルを読み込みました: %s", path)
				envLoaded = true
				break
			} else {
				loadErr = err
				log.Printf("警告: ファイルは存在しますが読み込めませんでした: %s, エラー: %v", path, err)
			}
		}
	}
	
	if !envLoaded {
		log.Printf("警告: .envファイルが見つかりませんでした。最後のエラー: %v", loadErr)
		log.Println("システム環境変数を使用します。")
	}

	// 環境変数からデータベースURLを取得
	dbURL := os.Getenv("DATABASE_URL")
	
	// 環境変数の内容をログに出力（セキュリティ上の理由から本番環境では避けてください）
	if dbURL != "" {
		log.Printf("DATABASE_URL環境変数が設定されています")
	} else {
		// .envファイルの内容を直接読み込んでみる（最終手段）
	envPath := filepath.Join(rootDir, ".env")
	if envContent, err := os.ReadFile(envPath); err == nil {
		log.Printf(".envファイルを直接読み込みました: %s", envPath)
		
		// ファイル内容をテキストとして解析
		fileStr := string(envContent)
		
		// BOMがあれば削除 (UTF-8のBOM: \xef\xbb\xbf)
		if strings.HasPrefix(fileStr, "\ufeff") {
			fileStr = strings.TrimPrefix(fileStr, "\ufeff")
			log.Printf("BOMマーカーを削除しました")
		}
		
		// 行ごとに処理
		lines := strings.Split(fileStr, "\n")
		for _, line := range lines {
			// 空行やコメント行はスキップ
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
				continue
			}
			
			// DATABASE_URLを探す
			if strings.HasPrefix(trimmedLine, "DATABASE_URL=") {
				// クォーテーションとキャリッジリターンを削除
				dbURLRaw := strings.TrimPrefix(trimmedLine, "DATABASE_URL=")
				dbURLRaw = strings.Trim(dbURLRaw, "\"'")
				dbURLRaw = strings.TrimRight(dbURLRaw, "\r")
				
				// 結果をログに出力（センシティブな情報なので本番環境では注意）
				log.Printf("DATABASE_URLを.envファイルから抽出しました: %s", dbURLRaw[:10] + "...")
				dbURL = dbURLRaw
				break
			}
		}
	} else {
		log.Printf(".envファイルを直接読み込めませんでした: %v", err)
	}
	}
	
	// 環境変数が設定されていない場合はエラー
	if dbURL == "" {
		log.Fatal("エラー: DATABASE_URL環境変数が設定されていません。.envファイルを確認してください。")
	}

	log.Println("Supabaseに接続しています...")

	// データベースに接続
	db, err := database.NewPostgresDB(dbURL)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// 接続テストを実行
	if err := database.TestConnection(db); err != nil {
		log.Fatalf("接続テストに失敗しました: %v", err)
	}

	log.Println("Supabaseへの接続に成功し、接続を確認しました！")

	// サンプルユーザーテーブルを作成
	// 注意: 本番環境では通常マイグレーションツールを使用します
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS sample_users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("テーブル作成に失敗しました: %v", err)
	}
	log.Println("テーブル 'sample_users' の準備ができました")

	// ユーザーを挿入
	username := "testuser"
	email := "test@example.com"
	
	// 既存のユーザーを削除（テスト用）
	_, err = db.Exec("DELETE FROM sample_users WHERE username = $1", username)
	if err != nil {
		log.Printf("既存ユーザーの削除中にエラーが発生しました: %v", err)
	}
	
	// 新しいユーザーを挿入
	var userID int
	err = db.QueryRow(
		"INSERT INTO sample_users (username, email) VALUES ($1, $2) RETURNING id",
		username, email,
	).Scan(&userID)
	
	if err != nil {
		log.Fatalf("ユーザー挿入に失敗しました: %v", err)
	}
	log.Printf("ID: %d のユーザーを挿入しました", userID)

	// ユーザーを取得
	var user User
	err = db.QueryRow(
		"SELECT id, username, email, created_at FROM sample_users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	
	if err != nil {
		log.Fatalf("ユーザー取得に失敗しました: %v", err)
	}
	log.Printf("ユーザーを取得しました: ID=%d, ユーザー名=%s, メール=%s, 作成日時=%v",
		user.ID, user.Username, user.Email, user.CreatedAt)

	// ユーザーを更新
	newEmail := "updated@example.com"
	_, err = db.Exec(
		"UPDATE sample_users SET email = $1 WHERE id = $2",
		newEmail, userID,
	)
	
	if err != nil {
		log.Fatalf("ユーザー更新に失敗しました: %v", err)
	}
	log.Printf("ID %d のユーザーを更新しました、新しいメール: %s", userID, newEmail)

	// 更新を確認
	var updatedEmail string
	err = db.QueryRow(
		"SELECT email FROM sample_users WHERE id = $1", userID,
	).Scan(&updatedEmail)
	
	if err != nil {
		log.Fatalf("更新されたユーザーの取得に失敗しました: %v", err)
	}
	log.Printf("ユーザーのメールが次のように更新されたことを確認しました: %s", updatedEmail)

	// ユーザーの削除はコメントアウト（必要に応じて使用）
	/*
	_, err = db.Exec("DELETE FROM sample_users WHERE id = $1", userID)
	if err != nil {
		log.Fatalf("ユーザー削除に失敗しました: %v", err)
	}
	log.Printf("ID %d のユーザーを削除しました", userID)
	*/

	// ユーザー一覧を取得
	rows, err := db.Query("SELECT id, username, email, created_at FROM sample_users")
	if err != nil {
		log.Fatalf("ユーザー一覧の取得に失敗しました: %v", err)
	}
	defer rows.Close()

	log.Println("全ユーザー:")
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			log.Printf("ユーザーデータの読み取り中にエラーが発生しました: %v", err)
			continue
		}
		log.Printf("- ID: %d, ユーザー名: %s, メール: %s, 作成日時: %v", 
			u.ID, u.Username, u.Email, u.CreatedAt)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ユーザー一覧の反復処理中にエラーが発生しました: %v", err)
	}

	fmt.Println("\nSupabase接続とCRUD操作が正常に完了しました！")
}