package handlers

import (
	"net/http"
	"strconv"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler ユーザー関連のハンドラーを管理する構造体
type UserHandler struct {
	userRepo   interfaces.UserRepository
	followRepo interfaces.FollowRepository
	postRepo   interfaces.PostRepository
	log        logger.Logger
}

// NewUserHandler 新しいユーザーハンドラーを作成する
func NewUserHandler(
	userRepo interfaces.UserRepository,
	followRepo interfaces.FollowRepository,
	postRepo interfaces.PostRepository,
	log logger.Logger,
) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		followRepo: followRepo,
		postRepo:   postRepo,
		log:        log,
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
		response.ServerError(c, "ユーザー情報の取得中にエラーが発生しました")
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
			response.ServerError(c, "プロフィールの更新中にエラーが発生しました")
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
	followers, err := h.followRepo.GetFollowers(c, user.ID, offset, perPage)
	if err != nil {
		h.log.Error("フォロワー取得中にエラーが発生しました", "error", err)
		response.ServerError(c, "フォロワーの取得中にエラーが発生しました")
		return
	}

	// フォロワーの総数を取得
	totalFollowers, err := h.followRepo.CountFollowers(c, user.ID)
	if err != nil {
		h.log.Error("フォロワー数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, _ = uuid.Parse(currentUserIDStr.(string))
	}

	// フォロワーのレスポンスを作成
	followersResponse := make([]gin.H, 0, len(followers))
	for _, follower := range followers {
		// 現在のユーザーがフォロワーをフォローしているかを確認
		isFollowing := false
		if currentUserID != uuid.Nil && currentUserID != follower.ID {
			isFollowing, _ = h.followRepo.IsFollowing(c, currentUserID, follower.ID)
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

	// ユーザーのフォローしているユーザーを取得
	following, err := h.followRepo.GetFollowing(c, user.ID, offset, perPage)
	if err != nil {
		h.log.Error("フォロー中ユーザー取得中にエラーが発生しました", "error", err)
		response.ServerError(c, "フォロー中ユーザーの取得中にエラーが発生しました")
		return
	}

	// フォロー中ユーザーの総数を取得
	totalFollowing, err := h.followRepo.CountFollowing(c, user.ID)
	if err != nil {
		h.log.Error("フォロー中ユーザー数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, _ = uuid.Parse(currentUserIDStr.(string))
	}

	// フォロー中ユーザーのレスポンスを作成
	followingResponse := make([]gin.H, 0, len(following))
	for _, followed := range following {
		// 現在のユーザーがこのユーザーをフォローしているかを確認
		isFollowing := false
		if currentUserID != uuid.Nil && currentUserID != followed.ID {
			isFollowing, _ = h.followRepo.IsFollowing(c, currentUserID, followed.ID)
		}

		followingResponse = append(followingResponse, gin.H{
			"id":           followed.ID,
			"username":     followed.Username,
			"display_name": followed.Name,
			"avatar_url":   followed.ProfileImage,
			"bio":          followed.Bio,
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
		response.ServerError(c, "ユーザー情報の取得中にエラーが発生しました")
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
		response.ServerError(c, "フォロー情報の確認中にエラーが発生しました")
		return
	}

	// 既にフォローしている場合
	if isFollowing {
		response.BadRequest(c, "既にフォローしています", nil)
		return
	}

	// フォローの作成
	follow := models.NewFollow(currentUserID, targetUser.ID)
	if err := h.followRepo.Create(c, follow); err != nil {
		h.log.Error("フォロー作成中にエラーが発生しました", "error", err)
		response.ServerError(c, "フォロー処理中にエラーが発生しました")
		return
	}

	// フォロワー数を取得
	followerCount, err := h.followRepo.CountFollowers(c, targetUser.ID)
	if err != nil {
		h.log.Error("フォロワー数の取得中にエラーが発生しました", "error", err)
		// エラーがあってもレスポンスは返す
		followerCount = targetUser.FollowerCount + 1
	}

	response.Success(c, gin.H{
		"following":       true,
		"followers_count": followerCount,
	})
}

// UnfollowUser ユーザーのフォロー解除ハンドラー
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
		response.ServerError(c, "ユーザー情報の取得中にエラーが発生しました")
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

	// フォローの削除
	if err := h.followRepo.Delete(c, currentUserID, targetUser.ID); err != nil {
		h.log.Error("フォロー解除中にエラーが発生しました", "error", err)
		response.ServerError(c, "フォロー解除処理中にエラーが発生しました")
		return
	}

	// フォロワー数を取得
	followerCount, err := h.followRepo.CountFollowers(c, targetUser.ID)
	if err != nil {
		h.log.Error("フォロワー数の取得中にエラーが発生しました", "error", err)
		// エラーがあってもレスポンスは返す
		if targetUser.FollowerCount > 0 {
			followerCount = targetUser.FollowerCount - 1
		}
	}

	response.Success(c, gin.H{
		"following":       false,
		"followers_count": followerCount,
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
		response.ServerError(c, "投稿の取得中にエラーが発生しました")
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