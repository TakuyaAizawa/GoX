package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/TakuyaAizawa/gox/pkg/logger"
)

// パニックから回復するミドルウェア
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// エラーとスタックトレースをログに記録
				stackTrace := string(debug.Stack())
				log.Error("パニックから回復しました",
					"error", fmt.Sprintf("%v", err),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", stackTrace,
				)

				// クライアントにサーバーエラーを返す
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "内部サーバーエラー",
				})
			}
		}()
		
		c.Next()
	}
} 