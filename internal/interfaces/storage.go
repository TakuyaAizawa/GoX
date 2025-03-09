package interfaces

import (
	"context"
	"io"
	"time"
)

// StorageProvider はメディアファイルのストレージ操作を定義するインターフェース
type StorageProvider interface {
	// SaveFile はファイルを保存し、そのURLを返します
	SaveFile(ctx context.Context, path string, filename string, fileContent io.Reader, fileSize int64) (string, error)

	// DeleteFile は指定されたパスのファイルを削除します
	DeleteFile(ctx context.Context, path string) error

	// GetSignedURL は期限付きの署名付きURLを生成します（第三者ストレージ用）
	GetSignedURL(ctx context.Context, path string, expires time.Duration) (string, error)
}
