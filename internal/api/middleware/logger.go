package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/TakuyaAizawa/gox/pkg/logger"
)

// リクエスト詳細をログに記録するミドルウェアを返す
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエスト開始時間
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// ハンドラー処理の前にログ
		log.Info("リクエスト開始",
			"method", method,
			"path", path,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// 次のミドルウェアを呼び出し
		c.Next()

		// ハンドラー処理後
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		// レスポンスのログレベルはステータスコードに基づく
		if statusCode >= 500 {
			log.Error("リクエスト完了",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
				"error", c.Errors.String(),
			)
		} else if statusCode >= 400 {
			log.Warn("リクエスト完了",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
				"error", c.Errors.String(),
			)
		} else {
			log.Info("リクエスト完了",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
			)
		}
	}
} 