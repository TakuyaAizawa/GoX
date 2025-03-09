package handlers

import (
	"fmt"
	"sort"
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
	currentUserIDInterface, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}

	// 型変換エラーを防ぐため、安全に型変換を行う
	var currentUserID uuid.UUID
	var err error

	// 型に応じた変換処理
	switch v := currentUserIDInterface.(type) {
	case uuid.UUID:
		currentUserID = v
	case string:
		currentUserID, err = uuid.Parse(v)
		if err != nil {
			h.log.Error("ユーザーIDのパース中にエラーが発生しました", "error", err, "value", v)
			response.InternalServerError(c, "認証処理に問題が発生しました")
			return
		}
	default:
		h.log.Error("ユーザーIDの型変換に失敗しました", "type", fmt.Sprintf("%T", currentUserIDInterface))
		response.InternalServerError(c, "認証処理に問題が発生しました")
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
	following, err := h.followRepo.GetFollowing(c.Request.Context(), currentUserID, 0, 1000) // 一度に取得するフォロー数に制限を設ける
	if err != nil {
		h.log.Error("フォロー中ユーザーID取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "タイムラインの取得中にエラーが発生しました")
		return
	}

	// 自分の投稿も含める
	userIDs := append(following, currentUserID)

	// 各ユーザーの投稿を取得して結合
	var allPosts []*models.Post
	for _, userID := range userIDs {
		userPosts, err := h.postRepo.GetByUserID(c.Request.Context(), userID, offset, perPage)
		if err != nil {
			h.log.Error("投稿取得中にエラーが発生しました", "error", err, "userID", userID)
			continue
		}
		allPosts = append(allPosts, userPosts...)
	}

	// 投稿を時系列順にソート
	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].CreatedAt.After(allPosts[j].CreatedAt)
	})

	// ページネーションの範囲に限定
	var posts []*models.Post
	if len(allPosts) > 0 {
		end := offset + perPage
		if end > len(allPosts) {
			end = len(allPosts)
		}
		if offset < len(allPosts) {
			posts = allPosts[offset:end]
		}
	}

	// 総投稿数は取得した投稿の数をそのまま使用
	totalPosts := int64(len(allPosts))

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
		posts, err = h.postRepo.List(c.Request.Context(), offset, perPage)
	}

	if err != nil {
		h.log.Error("投稿取得中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "探索タイムラインの取得中にエラーが発生しました")
		return
	}

	// 投稿をいいね数+リポスト数の多い順にソート
	sort.Slice(posts, func(i, j int) bool {
		likesAndRepostsI := posts[i].LikeCount + posts[i].RepostCount
		likesAndRepostsJ := posts[j].LikeCount + posts[j].RepostCount
		return likesAndRepostsI > likesAndRepostsJ
	})

	// 現在のユーザーID（認証済みの場合）
	var currentUserID uuid.UUID
	if currentUserIDInterface, exists := c.Get("userID"); exists {
		// 型に応じた安全な変換
		switch v := currentUserIDInterface.(type) {
		case uuid.UUID:
			currentUserID = v
		case string:
			parsedUUID, err := uuid.Parse(v)
			if err != nil {
				h.log.Warn("ユーザーIDのパースに失敗しました", "error", err, "value", v)
				// 認証が必須でないので処理は続行
			} else {
				currentUserID = parsedUUID
			}
		default:
			h.log.Warn("ユーザーIDの型変換に失敗しました", "type", fmt.Sprintf("%T", currentUserIDInterface))
			// 認証が必須でないので処理は続行
		}
	}

	// 投稿の総数を概算
	// 探索タイムラインの場合は簡略化して投稿数をカウント
	var totalPosts int64 = 0
	// 取得した投稿数を総数の概算として使用（ページネーションのために）
	totalPosts = int64(len(posts)) * 10 // 概算値として表示用に調整

	// Note: 正確な数はパフォーマンス上の理由から計算しない

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
