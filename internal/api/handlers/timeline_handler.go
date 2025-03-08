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

// TimelineHandler タイムライン関連のハンドラーを管理する構造体
type TimelineHandler struct {
	postRepo   interfaces.PostRepository
	userRepo   interfaces.UserRepository
	followRepo interfaces.FollowRepository
	likeRepo   interfaces.LikeRepository
	log        logger.Logger
}

// NewTimelineHandler 新しいタイムラインハンドラーを作成する
func NewTimelineHandler(
	postRepo interfaces.PostRepository,
	userRepo interfaces.UserRepository,
	followRepo interfaces.FollowRepository,
	likeRepo interfaces.LikeRepository,
	log logger.Logger,
) *TimelineHandler {
	return &TimelineHandler{
		postRepo:   postRepo,
		userRepo:   userRepo,
		followRepo: followRepo,
		likeRepo:   likeRepo,
		log:        log,
	}
}

// GetHomeTimeline ホームタイムライン取得ハンドラー
// フォローしているユーザーの投稿を時系列順で取得する
func (h *TimelineHandler) GetHomeTimeline(c *gin.Context) {
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

	// フォローしているユーザーのIDを取得
	following, err := h.followRepo.GetFollowingIDs(c, currentUserID)
	if err != nil {
		h.log.Error("フォロー中ユーザーID取得中にエラーが発生しました", "error", err)
		response.ServerError(c, "タイムラインの取得中にエラーが発生しました")
		return
	}

	// 自分の投稿も含めるために自分のIDも追加
	userIDs := append(following, currentUserID)

	// フォローしているユーザーの投稿を取得
	posts, err := h.postRepo.GetByUserIDs(c, userIDs, offset, perPage)
	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.ServerError(c, "タイムラインの取得中にエラーが発生しました")
		return
	}

	// 投稿の総数を取得（概算）
	// Note: 正確な数を取得するよりも、ページネーションの大まかな情報提供が目的
	totalPosts, err := h.postRepo.CountByUserIDs(c, userIDs)
	if err != nil {
		h.log.Error("投稿数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
		totalPosts = int64(len(posts))
	}

	// 投稿のレスポンスを作成
	postsResponse := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		// 投稿ユーザーの情報を取得
		user, err := h.userRepo.GetByID(c, post.UserID)
		if err != nil {
			h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
			continue // このユーザーの情報は取得できないのでスキップ
		}

		// いいね状態の確認
		isLiked, _ := h.likeRepo.HasLiked(c, currentUserID, post.ID)

		// リポスト状態の確認
		// TODO: リポジトリにHasRepostedメソッドを追加する必要があります
		isReposted := false

		// 投稿レスポンスを作成
		postResponse := gin.H{
			"id":            post.ID,
			"user_id":       post.UserID,
			"content":       post.Content,
			"media_urls":    post.MediaURLs,
			"created_at":    post.CreatedAt,
			"likes_count":   post.LikeCount,
			"replies_count": post.ReplyCount,
			"reposts_count": post.RepostCount,
			"is_liked":      isLiked,
			"is_reposted":   isReposted,
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.Name,
				"avatar_url":   user.ProfileImage,
			},
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

		// リポストの場合はリポスト元の情報も追加
		if post.IsRepost && post.RepostID != nil {
			repostPost, err := h.postRepo.GetByID(c, *post.RepostID)
			if err == nil {
				repostUser, err := h.userRepo.GetByID(c, repostPost.UserID)
				if err == nil {
					postResponse["repost"] = gin.H{
						"id":         repostPost.ID,
						"user_id":    repostPost.UserID,
						"content":    repostPost.Content,
						"created_at": repostPost.CreatedAt,
						"user": gin.H{
							"username":     repostUser.Username,
							"display_name": repostUser.Name,
							"avatar_url":   repostUser.ProfileImage,
						},
					}
				}
			}
		}

		postsResponse = append(postsResponse, postResponse)
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

// GetExploreTimeline 探索タイムライン取得ハンドラー
// 人気の投稿や新着投稿を取得する
func (h *TimelineHandler) GetExploreTimeline(c *gin.Context) {
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

	// ソート方法を取得（デフォルトは人気順）
	sortBy := c.DefaultQuery("sort_by", "popular")

	var posts []*models.Post
	var err error

	// ソート方法に応じた投稿を取得
	if sortBy == "latest" {
		// 最新の投稿を取得
		posts, err = h.postRepo.List(c, offset, perPage)
	} else {
		// 人気の投稿を取得（いいねとリポストの合計数でソート）
		posts, err = h.postRepo.ListPopular(c, offset, perPage)
	}

	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.ServerError(c, "探索タイムラインの取得中にエラーが発生しました")
		return
	}

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	if currentUserIDStr, exists := c.Get("userID"); exists {
		currentUserID, _ = uuid.Parse(currentUserIDStr.(string))
	}

	// 投稿の総数を取得（概算）
	totalPosts, err := h.postRepo.Count(c)
	if err != nil {
		h.log.Error("投稿数の取得中にエラーが発生しました", "error", err)
		// エラーがあっても処理は続行
		totalPosts = int64(len(posts))
	}

	// 投稿のレスポンスを作成
	postsResponse := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		// 投稿ユーザーの情報を取得
		user, err := h.userRepo.GetByID(c, post.UserID)
		if err != nil {
			h.log.Error("ユーザー取得中にエラーが発生しました", "error", err)
			continue // このユーザーの情報は取得できないのでスキップ
		}

		// いいね状態の確認
		isLiked := false
		if currentUserID != uuid.Nil {
			isLiked, _ = h.likeRepo.HasLiked(c, currentUserID, post.ID)
		}

		postsResponse = append(postsResponse, gin.H{
			"id":            post.ID,
			"user_id":       post.UserID,
			"content":       post.Content,
			"media_urls":    post.MediaURLs,
			"created_at":    post.CreatedAt,
			"likes_count":   post.LikeCount,
			"replies_count": post.ReplyCount,
			"reposts_count": post.RepostCount,
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
