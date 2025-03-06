package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// クライアントのレート制限データを表す構造体
type RateLimitClient struct {
	Count      int       // リクエスト数
	ResetTime  time.Time // リセット時刻
}

// リクエスト数を制限するミドルウェアを返す
func RateLimit(limit int, duration time.Duration) gin.HandlerFunc {
	// IPアドレスごとのリクエスト数を保持するマップ
	clients := make(map[string]*RateLimitClient)
	var mutex sync.Mutex
	
	return func(c *gin.Context) {
		// クライアントIPを取得
		clientIP := c.ClientIP()
		
		mutex.Lock()
		defer mutex.Unlock()
		
		// 新しいクライアントの場合は初期化
		if _, exists := clients[clientIP]; !exists {
			clients[clientIP] = &RateLimitClient{
				Count:      0,
				ResetTime:  time.Now().Add(duration),
			}
		}
		
		client := clients[clientIP]
		now := time.Now()
		
		// リセット時間を過ぎていれば、カウンターをリセット
		if now.After(client.ResetTime) {
			client.Count = 0
			client.ResetTime = now.Add(duration)
		}
		
		// レート制限チェック
		if client.Count >= limit {
			// レスポンスヘッダーを設定
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", client.ResetTime.Unix()))
			c.Header("Retry-After", fmt.Sprintf("%d", int(client.ResetTime.Sub(now).Seconds())))
			
			// リクエスト過多エラーを返す
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "レート制限を超過しました",
			})
			return
		}
		
		// リクエストカウンターをインクリメント
		client.Count++
		
		// レスポンスヘッダーを設定
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-client.Count))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", client.ResetTime.Unix()))
		
		c.Next()
	}
} 