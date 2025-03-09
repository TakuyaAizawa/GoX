package handlers

import (
	"strconv"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandler 通知関連のハンドラーを管理する構造体
type NotificationHandler struct {
	notificationRepo interfaces.NotificationRepository
	userRepo         interfaces.UserRepository
	postRepo         interfaces.PostRepository
	log              logger.Logger
}

// NewNotificationHandler 新しい通知ハンドラーを作成する
func NewNotificationHandler(
	notificationRepo interfaces.NotificationRepository,
	userRepo interfaces.UserRepository,
	postRepo interfaces.PostRepository,
	log logger.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		postRepo:         postRepo,
		log:              log,
	}
}

// GetNotifications ユーザーの通知一覧を取得する
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// ユーザーIDを取得
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}

	// クエリパラメータを取得
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	// typeFilterは使用していないので削除

	// パラメータの変換
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// ページネーション用のオフセットを計算
	offset := (page - 1) * limit
	perPage := limit

	// 通知の取得
	notifications, err := h.notificationRepo.GetByUserID(c.Request.Context(), currentUserID.(uuid.UUID), offset, perPage)
	if err != nil {
		h.log.Error("通知取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "通知の取得中にエラーが発生しました")
		return
	}

	// 通知の総数を取得
	totalNotifications, err := h.notificationRepo.CountUnreadByUserID(c.Request.Context(), currentUserID.(uuid.UUID))
	if err != nil {
		h.log.Error("通知数の取得中にエラーが発生しました", "error", err)
		totalNotifications = int64(len(notifications))
	}

	// 未読の通知を既読にマーク
	if len(notifications) > 0 {
		err = h.notificationRepo.MarkAllAsRead(c.Request.Context(), currentUserID.(uuid.UUID))
		if err != nil {
			h.log.Error("通知の既読マーク中にエラーが発生しました", "error", err)
		}
	}

	// 通知レスポンスの作成
	notificationsResponse := make([]gin.H, 0, len(notifications))
	for _, notification := range notifications {
		// アクション実行者の情報を取得
		actor, err := h.userRepo.GetByID(c, notification.ActorID)
		if err != nil {
			h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
			continue
		}

		notificationResponse := gin.H{
			"id":         notification.ID,
			"type":       notification.Type,
			"created_at": notification.CreatedAt,
			"read":       notification.IsRead,
			"actor": gin.H{
				"id":           actor.ID,
				"username":     actor.Username,
				"display_name": actor.Name,
				"avatar_url":   actor.ProfileImage,
			},
		}

		// 通知タイプに応じて追加情報を取得
		switch notification.Type {
		case models.NotificationTypeLike, models.NotificationTypeReply, models.NotificationTypeRepost:
			if notification.PostID != nil {
				post, err := h.postRepo.GetByID(c, *notification.PostID)
				if err == nil {
					notificationResponse["post"] = gin.H{
						"id":         post.ID,
						"content":    post.Content,
						"created_at": post.CreatedAt,
					}
				}
			}
		}

		notificationsResponse = append(notificationsResponse, notificationResponse)
	}

	// ページネーション情報を含むレスポンスを返す
	totalPages := int(totalNotifications) / perPage
	if int(totalNotifications)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"notifications": notificationsResponse,
		"pagination": gin.H{
			"total":       totalNotifications,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}

// GetUnreadCount 未読通知の数を取得する
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// 現在のユーザーIDを取得
	currentUserIDStr, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}

	currentUserID, err := uuid.Parse(currentUserIDStr.(string))
	if err != nil {
		h.log.Error("ユーザーIDのパース中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "ユーザー情報の取得中にエラーが発生しました")
		return
	}

	// 未読通知数の取得
	unreadCount, err := h.notificationRepo.CountUnreadByUserID(c, currentUserID)
	if err != nil {
		h.log.Error("未読通知数の取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "通知情報の取得中にエラーが発生しました")
		return
	}

	response.Success(c, gin.H{
		"unread_count": unreadCount,
	})
}

// MarkAsRead 通知を既読にする
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// 現在のユーザーIDを取得
	currentUserIDStr, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}

	currentUserID, err := uuid.Parse(currentUserIDStr.(string))
	if err != nil {
		h.log.Error("ユーザーIDのパース中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "ユーザー情報の取得中にエラーが発生しました")
		return
	}

	// リクエストからパラメータを取得
	var req struct {
		NotificationID *uuid.UUID `json:"notification_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "リクエストの形式が正しくありません", nil)
		return
	}

	// 特定の通知IDが指定されている場合はその通知のみを既読に
	// 指定されていない場合はすべての通知を既読にする
	if req.NotificationID != nil {
		err = h.notificationRepo.MarkAsRead(c.Request.Context(), *req.NotificationID)
	} else {
		err = h.notificationRepo.MarkAllAsRead(c.Request.Context(), currentUserID)
	}
	if err != nil {
		h.log.Error("通知の既読マーク中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "通知の更新中にエラーが発生しました")
		return
	}

	response.Success(c, gin.H{
		"message": "通知を既読にしました",
	})
}
