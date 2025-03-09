package handlers

import (
	"strconv"

	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/service"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler ユーザー関連のハンドラーを管理する構造体
type UserHandler struct {
	userRepo            interfaces.UserRepository
	followRepo          interfaces.FollowRepository
	postRepo            interfaces.PostRepository
	notificationService *service.NotificationService
	log                 logger.Logger
}

// NewUserHandler 新しいユーザーハンドラーを作成する
func NewUserHandler(
	userRepo interfaces.UserRepository,
	followRepo interfaces.FollowRepository,
	postRepo interfaces.PostRepository,
	notificationService *service.NotificationService,
	log logger.Logger,
) *UserHandler {
	return &UserHandler{
		userRepo:            userRepo,
		followRepo:          followRepo,
		postRepo:            postRepo,
		notificationService: notificationService,
		log:                 log,
	}
}

// GetUserProfile ユーザープロフィール取得ハンドラー
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
		return
	}

	// ユーザーをユーザー名で検索
	user, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// 現在のユーザーがフォローしているかどうかを確認
	isFollowing := false
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, err := uuid.Parse(currentUserIDStr.(string))
		if err == nil && currentUserID != user.ID {
			isFollowing, err = h.followRepo.IsFollowing(c, currentUserID, user.ID)
			if err != nil {
				h.log.Error("フォロー状態の確認中にエラーが発生しました", "error", err)
				// エラーがあってもプロフィール表示は続行
			}
		}
	}

	// レスポンスを組み立てて返す
	response.Success(c, gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"display_name":    user.Name,
		"bio":             user.Bio,
		"avatar_url":      user.ProfileImage,
		"banner_url":      user.BannerImage,
		"location":        user.Location,
		"website_url":     user.WebsiteURL,
		"verified":        user.IsVerified,
		"created_at":      user.CreatedAt,
		"followers_count": user.FollowerCount,
		"following_count": user.FollowingCount,
		"posts_count":     user.PostCount,
		"is_following":    isFollowing,
	})
}

// UpdateProfileRequest プロフィール更新リクエストの構造体
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,min=1,max=50"`
	Bio         string `json:"bio" binding:"omitempty,max=160"`
	Location    string `json:"location" binding:"omitempty,max=30"`
	WebsiteURL  string `json:"website_url" binding:"omitempty,max=100,url"`
}

// UpdateProfile プロフィール更新ハンドラー
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
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

	// 現在のユーザー情報を取得
	user, err := h.userRepo.GetByID(c, currentUserID)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// 変更があるフィールドのみ更新
	updated := false

	if req.DisplayName != "" && req.DisplayName != user.Name {
		user.Name = req.DisplayName
		updated = true
	}

	if req.Bio != user.Bio {
		user.Bio = req.Bio
		updated = true
	}

	if req.Location != user.Location {
		user.Location = req.Location
		updated = true
	}

	if req.WebsiteURL != user.WebsiteURL {
		user.WebsiteURL = req.WebsiteURL
		updated = true
	}

	// 変更があれば更新
	if updated {
		if err := h.userRepo.Update(c, user); err != nil {
			h.log.Error("ユーザー更新中にエラーが発生しました", "error", err)
			response.InternalServerError(c, "プロフィールの更新中にエラーが発生しました")
			return
		}
	}

	// 更新後のユーザー情報を返す
	response.Success(c, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.Name,
		"bio":          user.Bio,
		"avatar_url":   user.ProfileImage,
		"banner_url":   user.BannerImage,
		"location":     user.Location,
		"website_url":  user.WebsiteURL,
		"verified":     user.IsVerified,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	})
}

// GetFollowers フォロワー一覧取得ハンドラー
func (h *UserHandler) GetFollowers(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
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

	// ユーザーをユーザー名で検索
	user, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// ユーザーのフォロワーを取得
	followerIDs, err := h.followRepo.GetFollowers(c.Request.Context(), user.ID, offset, perPage)
	if err != nil {
		h.log.Error("フォロワー取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロワーの取得中にエラーが発生しました")
		return
	}

	// フォロワーの総数を取得
	totalFollowers, err := h.followRepo.CountFollowers(c.Request.Context(), user.ID)
	if err != nil {
		h.log.Error("フォロワー数取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロワーの取得中にエラーが発生しました")
		return
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	currentUserIDInterface, exists := c.Get("userID")
	if exists {
		currentUserID = currentUserIDInterface.(uuid.UUID)
	}

	// フォロワーのレスポンスを作成
	followersResponse := make([]gin.H, 0, len(followerIDs))
	for _, followerID := range followerIDs {
		// ユーザー情報を取得
		follower, err := h.userRepo.GetByID(c.Request.Context(), followerID)
		if err != nil {
			h.log.Error("フォロワー情報取得中にエラーが発生しました", "error", err, "followerID", followerID)
			continue
		}

		// 現在のユーザーがフォロワーをフォローしているかを確認
		isFollowing := false
		if currentUserID != uuid.Nil && currentUserID != follower.ID {
			isFollowing, _ = h.followRepo.IsFollowing(c.Request.Context(), currentUserID, follower.ID)
		}

		followersResponse = append(followersResponse, gin.H{
			"id":           follower.ID,
			"username":     follower.Username,
			"display_name": follower.Name,
			"avatar_url":   follower.ProfileImage,
			"bio":          follower.Bio,
			"is_following": isFollowing,
		})
	}

	// ページネーション情報を含むレスポンスを返す
	totalPages := int(totalFollowers) / perPage
	if int(totalFollowers)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"users": followersResponse,
		"pagination": gin.H{
			"total":       totalFollowers,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}

// GetFollowing フォロー中ユーザー一覧取得ハンドラー
func (h *UserHandler) GetFollowing(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
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

	// ユーザーをユーザー名で検索
	user, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// ユーザーがフォローしているユーザーを取得
	followingIDs, err := h.followRepo.GetFollowing(c.Request.Context(), user.ID, offset, perPage)
	if err != nil {
		h.log.Error("フォロー中ユーザー取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロー中ユーザーの取得中にエラーが発生しました")
		return
	}

	// フォロー中ユーザーの総数を取得
	totalFollowing, err := h.followRepo.CountFollowing(c.Request.Context(), user.ID)
	if err != nil {
		h.log.Error("フォロー中ユーザー数取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロー中ユーザーの取得中にエラーが発生しました")
		return
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	currentUserIDInterface, exists := c.Get("userID")
	if exists {
		currentUserID = currentUserIDInterface.(uuid.UUID)
	}

	// フォロー中ユーザーのレスポンスを作成
	followingResponse := make([]gin.H, 0, len(followingIDs))
	for _, followingID := range followingIDs {
		// ユーザー情報を取得
		followedUser, err := h.userRepo.GetByID(c.Request.Context(), followingID)
		if err != nil {
			h.log.Error("フォロー中ユーザー情報取得中にエラーが発生しました", "error", err, "followingID", followingID)
			continue
		}

		// 現在のユーザーがフォローしているかを確認
		isFollowing := false
		if currentUserID != uuid.Nil && currentUserID != followedUser.ID {
			isFollowing, _ = h.followRepo.IsFollowing(c.Request.Context(), currentUserID, followedUser.ID)
		}

		followingResponse = append(followingResponse, gin.H{
			"id":           followedUser.ID,
			"username":     followedUser.Username,
			"display_name": followedUser.Name,
			"avatar_url":   followedUser.ProfileImage,
			"bio":          followedUser.Bio,
			"is_following": isFollowing,
		})
	}

	// ページネーション情報を含むレスポンスを返す
	totalPages := int(totalFollowing) / perPage
	if int(totalFollowing)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"users": followingResponse,
		"pagination": gin.H{
			"total":       totalFollowing,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}

// FollowUser ユーザーをフォローするハンドラー
func (h *UserHandler) FollowUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
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

	// フォローするユーザーを取得
	targetUser, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// 自分自身をフォローしようとしている場合
	if currentUserID == targetUser.ID {
		response.BadRequest(c, "自分自身をフォローすることはできません", nil)
		return
	}

	// 既にフォローしているかどうかを確認
	isFollowing, err := h.followRepo.IsFollowing(c, currentUserID, targetUser.ID)
	if err != nil {
		h.log.Error("フォロー状態の確認中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロー情報の確認中にエラーが発生しました")
		return
	}

	// 既にフォローしている場合
	if isFollowing {
		response.BadRequest(c, "既にフォローしています", nil)
		return
	}

	// フォロー関係を作成
	err = h.followRepo.Follow(c.Request.Context(), currentUserID, targetUser.ID)
	if err != nil {
		h.log.Error("フォロー作成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロー処理中にエラーが発生しました")
		return
	}

	// フォロワー数を更新
	targetUser.FollowerCount++
	err = h.userRepo.Update(c.Request.Context(), targetUser)
	if err != nil {
		h.log.Error("ユーザー更新中にエラーが発生しました", "error", err)
		// エラーがあってもレスポンスは返す
	}

	// 通知の作成
	if h.notificationService != nil {
		err = h.notificationService.CreateFollowNotification(
			c.Request.Context(),
			currentUserID, // フォローした人
			targetUser.ID, // フォローされた人
		)
		if err != nil {
			h.log.Error("フォロー通知の作成中にエラーが発生しました", "error", err)
			// 通知作成のエラーはレスポンスには影響させない
		}
	}

	response.Success(c, gin.H{
		"following":       true,
		"followers_count": targetUser.FollowerCount,
	})
}

// UnfollowUser ユーザーのフォローを解除するハンドラー
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
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

	// フォロー解除するユーザーを取得
	targetUser, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// 自分自身をフォロー解除しようとしている場合
	if currentUserID == targetUser.ID {
		response.BadRequest(c, "自分自身のフォローを解除することはできません", nil)
		return
	}

	// フォロー関係を削除
	err = h.followRepo.Unfollow(c.Request.Context(), currentUserID, targetUser.ID)
	if err != nil {
		h.log.Error("フォロー解除中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "フォロー解除処理中にエラーが発生しました")
		return
	}

	// フォロワー数を更新
	if targetUser.FollowerCount > 0 {
		targetUser.FollowerCount--
		err = h.userRepo.Update(c.Request.Context(), targetUser)
		if err != nil {
			h.log.Error("ユーザー更新中にエラーが発生しました", "error", err)
			// エラーがあってもレスポンスは返す
		}
	}

	response.Success(c, gin.H{
		"following":       false,
		"followers_count": targetUser.FollowerCount,
	})
}

// GetUserPosts ユーザーの投稿一覧取得ハンドラー
func (h *UserHandler) GetUserPosts(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "ユーザー名が必要です", nil)
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

	// ユーザーをユーザー名で検索
	user, err := h.userRepo.GetByUsername(c, username)
	if err != nil {
		h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
		response.NotFound(c, "ユーザーが見つかりません")
		return
	}

	// ユーザーの投稿を取得
	posts, err := h.postRepo.GetByUserID(c, user.ID, offset, perPage)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "投稿の取得中にエラーが発生しました")
		return
	}

	// 投稿の総数を取得
	totalPosts, err := h.postRepo.CountByUserID(c, user.ID)
	if err != nil {
		h.log.Error("投稿数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
		totalPosts = int64(len(posts))
	}

	// 投稿のレスポンスを作成
	postsResponse := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		postsResponse = append(postsResponse, gin.H{
			"id":            post.ID,
			"user_id":       post.UserID,
			"content":       post.Content,
			"media_urls":    post.MediaURLs,
			"created_at":    post.CreatedAt,
			"likes_count":   post.LikeCount,
			"replies_count": post.ReplyCount,
			"reposts_count": post.RepostCount,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.Name,
				"avatar_url":   user.ProfileImage,
			},
			"is_liked":    false, // TODO: 現在のユーザーがいいねしているかどうかを確認
			"is_reposted": false, // TODO: 現在のユーザーがリポストしているかどうかを確認
		})
	}

	// ページネーション情報を含むレスポンスを返す
	totalPages := int(totalPosts) / perPage
	if int(totalPosts)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"posts": postsResponse,
		"pagination": gin.H{
			"total":       totalPosts,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}
