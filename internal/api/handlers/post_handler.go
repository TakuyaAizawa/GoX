package handlers

import (
	"strconv"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/service"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostHandler 投稿関連のハンドラーを管理する構造体
type PostHandler struct {
	postRepo            interfaces.PostRepository
	userRepo            interfaces.UserRepository
	likeRepo            interfaces.LikeRepository
	notificationRepo    interfaces.NotificationRepository
	notificationService *service.NotificationService
	log                 logger.Logger
}

// NewPostHandler 新しい投稿ハンドラーを作成する
func NewPostHandler(
	postRepo interfaces.PostRepository,
	userRepo interfaces.UserRepository,
	likeRepo interfaces.LikeRepository,
	notificationRepo interfaces.NotificationRepository,
	notificationService *service.NotificationService,
	log logger.Logger,
) *PostHandler {
	return &PostHandler{
		postRepo:            postRepo,
		userRepo:            userRepo,
		likeRepo:            likeRepo,
		notificationRepo:    notificationRepo,
		notificationService: notificationService,
		log:                 log,
	}
}

// CreatePostRequest 投稿作成リクエストの構造体
type CreatePostRequest struct {
	Content   string   `json:"content" binding:"required,max=280"`
	MediaURLs []string `json:"media_urls" binding:"omitempty,dive,url"`
	ReplyToID *string  `json:"reply_to_id" binding:"omitempty,uuid"`
}

// CreatePost 投稿作成ハンドラー
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

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

	var post *models.Post

	// 返信の場合
	if req.ReplyToID != nil {
		replyToID, err := uuid.Parse(*req.ReplyToID)
		if err != nil {
			response.BadRequest(c, "無効な返信先IDです", nil)
			return
		}

		// 返信先の投稿が存在するか確認
		replyToPost, err := h.postRepo.GetByID(c, replyToID)
		if err != nil {
			h.log.Error("返信先投稿の取得中にエラーが発生しました", "error", err)
			response.NotFound(c, "返信先の投稿が見つかりません")
			return
		}

		post = models.NewReply(currentUserID, replyToID, req.Content, req.MediaURLs)

		// 返信先の返信数をインクリメント
		if err := h.postRepo.IncrementReplyCount(c, replyToID); err != nil {
			h.log.Error("返信カウント更新中にエラーが発生しました", "error", err)
			// 処理は続行
		}

		// 通知の作成（自分自身の投稿への返信でない場合）
		if currentUserID != replyToPost.UserID {
			// TODO: 通知を作成
			notification := models.NewNotification(
				replyToPost.UserID,
				currentUserID,
				models.NotificationTypeReply,
				&post.ID,
			)
			if err := h.notificationRepo.Create(c, notification); err != nil {
				h.log.Error("通知の作成中にエラーが発生しました", "error", err)
				// 処理は続行
			}
		}
	} else {
		// 通常の投稿
		post = models.NewPost(currentUserID, req.Content, req.MediaURLs)
	}

	// 投稿の保存
	if err := h.postRepo.Create(c, post); err != nil {
		h.log.Error("投稿の作成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "投稿の作成中にエラーが発生しました")
		return
	}

	// ユーザー情報を取得
	user, err := h.userRepo.GetByID(c, currentUserID)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		// 投稿は作成されたのでエラーがあっても処理は続行
	}

	// レスポンスを作成
	postResponse := gin.H{
		"id":            post.ID,
		"user_id":       post.UserID,
		"content":       post.Content,
		"media_urls":    post.MediaURLs,
		"reply_to_id":   post.ReplyToID,
		"created_at":    post.CreatedAt,
		"likes_count":   0,
		"replies_count": 0,
		"reposts_count": 0,
	}

	// ユーザー情報があれば追加
	if user != nil {
		postResponse["user"] = gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.Name,
			"avatar_url":   user.ProfileImage,
		}
	}

	response.Created(c, postResponse)
}

// GetPost 投稿取得ハンドラー
func (h *PostHandler) GetPost(c *gin.Context) {
	// 投稿IDの取得とバリデーション
	idParam := c.Param("id")
	if idParam == "" {
		response.BadRequest(c, "投稿IDが必要です", nil)
		return
	}

	postID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "無効な投稿IDです", nil)
		return
	}

	// 投稿の取得
	post, err := h.postRepo.GetByID(c, postID)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "投稿が見つかりません")
		return
	}

	// 投稿ユーザーの情報を取得
	user, err := h.userRepo.GetByID(c, post.UserID)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		// 投稿は取得できたのでユーザー情報がなくても処理は続行
	}

	// 現在のユーザーがいいね・リポストしているか確認
	var isLiked, isReposted bool
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, err := uuid.Parse(currentUserIDStr.(string))
		if err == nil {
			// いいね状態の確認
			isLiked, err = h.likeRepo.HasLiked(c, currentUserID, post.ID)
			if err != nil {
				h.log.Error("いいね状態の確認中にエラーが発生しました", "error", err)
				// 処理は続行
			}

			// リポスト状態の確認
			// TODO: リポジトリにHasRepostedメソッドを追加する必要があります
			// isReposted, err = h.postRepo.HasReposted(c, currentUserID, post.ID)
		}
	}

	// レスポンスを作成
	postResponse := gin.H{
		"id":            post.ID,
		"user_id":       post.UserID,
		"content":       post.Content,
		"media_urls":    post.MediaURLs,
		"reply_to_id":   post.ReplyToID,
		"created_at":    post.CreatedAt,
		"likes_count":   post.LikeCount,
		"replies_count": post.ReplyCount,
		"reposts_count": post.RepostCount,
		"is_liked":      isLiked,
		"is_reposted":   isReposted,
	}

	// ユーザー情報があれば追加
	if user != nil {
		postResponse["user"] = gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.Name,
			"avatar_url":   user.ProfileImage,
		}
	}

	// 返信の場合は返信先の情報も追加
	if post.IsReply && post.ReplyToID != nil {
		replyToPost, err := h.postRepo.GetByID(c, *post.ReplyToID)
		if err == nil {
			replyToUser, err := h.userRepo.GetByID(c, replyToPost.UserID)
			if err == nil {
				postResponse["reply_to"] = gin.H{
					"id":         replyToPost.ID,
					"user_id":    replyToPost.UserID,
					"content":    replyToPost.Content,
					"created_at": replyToPost.CreatedAt,
					"user": gin.H{
						"username":     replyToUser.Username,
						"display_name": replyToUser.Name,
						"avatar_url":   replyToUser.ProfileImage,
					},
				}
			}
		}
	}

	response.Success(c, postResponse)
}

// DeletePost 投稿削除ハンドラー
func (h *PostHandler) DeletePost(c *gin.Context) {
	// 投稿IDの取得とバリデーション
	idParam := c.Param("id")
	if idParam == "" {
		response.BadRequest(c, "投稿IDが必要です", nil)
		return
	}

	postID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "無効な投稿IDです", nil)
		return
	}

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

	// 投稿の取得
	post, err := h.postRepo.GetByID(c, postID)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "投稿が見つかりません")
		return
	}

	// 投稿のオーナーかどうか確認
	if post.UserID != currentUserID {
		response.Forbidden(c, "この操作を行う権限がありません")
		return
	}

	// 投稿の削除
	if err := h.postRepo.Delete(c, postID); err != nil {
		h.log.Error("投稿の削除中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "投稿の削除中にエラーが発生しました")
		return
	}

	// 返信の場合は返信先の返信数をデクリメント
	if post.IsReply && post.ReplyToID != nil {
		if err := h.postRepo.DecrementReplyCount(c, *post.ReplyToID); err != nil {
			h.log.Error("返信カウント更新中にエラーが発生しました", "error", err)
			// 処理は続行
		}
	}

	response.NoContent(c)
}

// GetPostReplies 投稿への返信一覧取得ハンドラー
func (h *PostHandler) GetPostReplies(c *gin.Context) {
	// 投稿IDの取得とバリデーション
	idParam := c.Param("id")
	if idParam == "" {
		response.BadRequest(c, "投稿IDが必要です", nil)
		return
	}

	postID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "無効な投稿IDです", nil)
		return
	}

	// ページネーションパラメータの取得
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// 投稿が存在するか確認
	_, err = h.postRepo.GetByID(c, postID)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "投稿が見つかりません")
		return
	}

	// 返信の取得
	replies, err := h.postRepo.GetReplies(c, postID, offset, perPage)
	if err != nil {
		h.log.Error("返信取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "返信の取得中にエラーが発生しました")
		return
	}

	// 返信の総数を取得
	totalReplies, err := h.postRepo.CountReplies(c, postID)
	if err != nil {
		h.log.Error("返信数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
		totalReplies = int64(len(replies))
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, _ = uuid.Parse(currentUserIDStr.(string))
	}

	// 返信のレスポンスを作成
	repliesResponse := make([]gin.H, 0, len(replies))
	for _, reply := range replies {
		// ユーザー情報を取得
		user, err := h.userRepo.GetByID(c, reply.UserID)
		if err != nil {
			h.log.Error("ユーザー取得中にエラーが発生しました", "error", err, "userID", reply.UserID)
			continue // このユーザーの情報は取得できないのでスキップ
		}

		// いいね状態の確認
		isLiked := false
		if currentUserID != uuid.Nil {
			isLiked, _ = h.likeRepo.HasLiked(c, currentUserID, reply.ID)
		}

		repliesResponse = append(repliesResponse, gin.H{
			"id":            reply.ID,
			"user_id":       reply.UserID,
			"content":       reply.Content,
			"media_urls":    reply.MediaURLs,
			"reply_to_id":   reply.ReplyToID,
			"created_at":    reply.CreatedAt,
			"likes_count":   reply.LikeCount,
			"replies_count": reply.ReplyCount,
			"is_liked":      isLiked,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.Name,
				"avatar_url":   user.ProfileImage,
			},
		})
	}

	// ページネーション情報を含むレスポンスを返す
	totalPages := int(totalReplies) / perPage
	if int(totalReplies)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"replies": repliesResponse,
		"pagination": gin.H{
			"total":       totalReplies,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}

// LikePost 投稿にいいねをするハンドラー
func (h *PostHandler) LikePost(c *gin.Context) {
	// 投稿IDのパラメータ取得
	postIDStr := c.Param("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		h.log.Error("投稿IDのパース中にエラーが発生しました", "error", err)
		response.BadRequest(c, "無効な投稿IDです", nil)
		return
	}

	// 現在のユーザーID（リクエスト処理の前に認証ミドルウェアで設定済み）
	currentUserIDInterface, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}
	currentUserID := currentUserIDInterface.(uuid.UUID)

	// 投稿の存在確認
	post, err := h.postRepo.GetByID(c.Request.Context(), postID)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "投稿が見つかりません")
		return
	}

	// 既にいいね済みかのチェック
	hasLiked, err := h.likeRepo.HasLiked(c.Request.Context(), currentUserID, postID)
	if err != nil {
		h.log.Error("いいね状態確認中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "いいね処理中にエラーが発生しました")
		return
	}

	if hasLiked {
		response.BadRequest(c, "既にいいねしています", nil)
		return
	}

	// いいねの作成
	like := models.NewLike(currentUserID, postID)
	if err := h.likeRepo.Like(c.Request.Context(), like); err != nil {
		h.log.Error("いいね作成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "いいね処理中にエラーが発生しました")
		return
	}

	// 通知サービスが設定されていれば通知を作成
	if h.notificationService != nil {
		// 投稿の所有者への通知
		err = h.notificationService.CreateLikeNotification(
			c.Request.Context(),
			currentUserID, // いいねした人
			post.UserID,   // 投稿主
			post.ID,       // いいねされた投稿
		)
		if err != nil {
			h.log.Error("いいね通知の作成中にエラーが発生しました", "error", err)
			// 通知作成のエラーはレスポンスには影響させない
		}
	}

	// 成功レスポンス
	response.Success(c, gin.H{
		"liked": true,
	})
}

// UnlikePost 投稿へのいいね解除ハンドラー
func (h *PostHandler) UnlikePost(c *gin.Context) {
	// 投稿IDの取得とバリデーション
	idParam := c.Param("id")
	if idParam == "" {
		response.BadRequest(c, "投稿IDが必要です", nil)
		return
	}

	postID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "無効な投稿IDです", nil)
		return
	}

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

	// 投稿が存在するか確認
	post, err := h.postRepo.GetByID(c, postID)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "投稿が見つかりません")
		return
	}

	// いいねしているかどうか確認
	hasLiked, err := h.likeRepo.HasLiked(c, currentUserID, postID)
	if err != nil {
		h.log.Error("いいね状態の確認中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "いいね情報の確認中にエラーが発生しました")
		return
	}

	// いいねしていない場合
	if !hasLiked {
		response.BadRequest(c, "いいねしていません", nil)
		return
	}

	// いいねの削除
	if err := h.likeRepo.Unlike(c.Request.Context(), currentUserID, postID); err != nil {
		h.log.Error("いいね削除中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "いいね解除処理中にエラーが発生しました")
		return
	}

	// いいね数を減らす
	if err := h.postRepo.DecrementLikeCount(c, postID); err != nil {
		h.log.Error("いいねカウント更新中にエラーが発生しました", "error", err)
		// 処理は続行
	}

	// いいね数を確認（0未満にならないように）
	likeCount := post.LikeCount - 1
	if likeCount < 0 {
		likeCount = 0
	}

	response.Success(c, gin.H{
		"liked":       false,
		"likes_count": likeCount,
	})
}

// TODO: RepostPost と CancelRepost の実装
