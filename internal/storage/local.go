package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/TakuyaAizawa/gox/internal/interfaces"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/google/uuid"
)

// LocalStorage はローカルファイルシステムを使用したストレージプロバイダーです
type LocalStorage struct {
	baseDir string
	baseURL string
	log     logger.Logger
}

// NewLocalStorage は新しいLocalStorageインスタンスを作成します
func NewLocalStorage(baseDir, baseURL string, log logger.Logger) interfaces.StorageProvider {
	// ベースディレクトリが存在するか確認し、存在しない場合は作成
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			log.Error("ストレージディレクトリの作成に失敗しました", "error", err)
		}
	}

	return &LocalStorage{
		baseDir: baseDir,
		baseURL: baseURL,
		log:     log,
	}
}

// SaveFile はファイルをローカルファイルシステムに保存します
func (s *LocalStorage) SaveFile(ctx context.Context, path string, filename string, fileContent io.Reader, fileSize int64) (string, error) {
	// ディレクトリが存在するか確認
	fullDirPath := filepath.Join(s.baseDir, path)
	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイル名の拡張子を取得
	ext := filepath.Ext(filename)

	// ユニークなファイル名を生成
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(fullDirPath, uniqueFilename)

	// ファイルを作成
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer dst.Close()

	// ファイルの内容をコピー
	if _, err = io.Copy(dst, fileContent); err != nil {
		return "", fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	// 公開URL
	publicURL := fmt.Sprintf("%s/%s/%s", s.baseURL, path, uniqueFilename)

	s.log.Info("ファイルを保存しました", "path", fullPath, "url", publicURL)

	return publicURL, nil
}

// DeleteFile はローカルファイルシステムからファイルを削除します
func (s *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			s.log.Warn("削除対象のファイルが存在しません", "path", path)
			return nil
		}
		return fmt.Errorf("ファイルの削除に失敗しました: %w", err)
	}

	s.log.Info("ファイルを削除しました", "path", path)

	return nil
}

// GetSignedURL はローカルストレージでは実際に署名URLは使用しないため、単純にURLを返します
func (s *LocalStorage) GetSignedURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	// ローカルストレージでは署名URLは不要のため、通常のURLを返す
	return fmt.Sprintf("%s/%s", s.baseURL, path), nil
}
