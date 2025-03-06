package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// アプリケーションのロガーインターフェース
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	With(keysAndValues ...interface{}) Logger
	Sync() error
}

// zapLoggerはzap.SugaredLoggerのラッパーでありLoggerインターフェースを実装
type zapLogger struct {
	*zap.SugaredLogger
}

// 新しいロガーインスタンスを作成
func NewLogger(level, format string) (Logger, error) {
	config := zap.NewProductionConfig()

	// ログレベルを設定
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("無効なログレベル: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// ログフォーマットを設定
	switch format {
	case "json":
		config.Encoding = "json"
	case "console":
		config.Encoding = "console"
	default:
		return nil, fmt.Errorf("サポートされていないログフォーマット: %s", format)
	}

	// 出力パスを設定
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// ロガーを構築
	baseLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("ロガーの構築に失敗しました: %w", err)
	}

	return &zapLogger{
		SugaredLogger: baseLogger.Sugar(),
	}, nil
}

// Debugレベルでのログ記録
func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

// Infoレベルでのログ記録
func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

// Warnレベルでのログ記録
func (l *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

// Errorレベルでのログ記録
func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

// 指定されたコンテキストを持つ子ロガーを作成
func (l *zapLogger) With(keysAndValues ...interface{}) Logger {
	return &zapLogger{
		SugaredLogger: l.SugaredLogger.With(keysAndValues...),
	}
}

// 致命的なメッセージをログに記録し、ステータスコード1で終了
func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Fatalw(msg, keysAndValues...)
	os.Exit(1)
} 