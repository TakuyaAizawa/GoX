package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	testing_helper "github.com/TakuyaAizawa/gox/internal/repository/postgres/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRepository(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	userRepo := NewUserRepository(db.Pool)
	postRepo := NewPostRepository(db.Pool)

	ctx := context.Background()
	testUser := &models.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       "hashedpassword",
		Name:           "Test User",
		Bio:            "Test bio",
		ProfileImage:   "https://example.com/image.jpg",
		FollowerCount:  0,
		FollowingCount: 0,
		PostCount:      0,
		IsVerified:     false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// テストユーザーを作成
	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// テスト投稿の作成
	testPost := &models.Post{
		ID:          uuid.New(),
		UserID:      testUser.ID,
		Content:     "Test content",
		MediaURLs:   []string{"image1.jpg", "image2.jpg"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		LikeCount:   0,
		RepostCount: 0,
		ReplyCount:  0,
	}

	// Create のテスト
	t.Run("Create", func(t *testing.T) {
		err := postRepo.Create(ctx, testPost)
		assert.NoError(t, err)
	})

	// GetByID のテスト
	t.Run("GetByID", func(t *testing.T) {
		post, err := postRepo.GetByID(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, testPost.ID, post.ID)
		assert.Equal(t, testPost.Content, post.Content)
		assert.Equal(t, testPost.MediaURLs, post.MediaURLs)

		// 存在しないIDでの取得を試みる
		_, err = postRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
	})

	// Update のテスト
	t.Run("Update", func(t *testing.T) {
		testPost.Content = "Updated content"
		err := postRepo.Update(ctx, testPost)
		require.NoError(t, err)

		// 更新された情報を確認
		updated, err := postRepo.GetByID(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated content", updated.Content)
	})

	// GetByUserID のテスト
	t.Run("GetByUserID", func(t *testing.T) {
		posts, err := postRepo.GetByUserID(ctx, testUser.ID, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, posts)
		assert.Equal(t, testPost.ID, posts[0].ID)

		// 存在しないユーザーIDでの取得
		posts, err = postRepo.GetByUserID(ctx, uuid.New(), 0, 10)
		require.NoError(t, err)
		assert.Empty(t, posts)
	})

	// Reply機能のテスト
	t.Run("Reply", func(t *testing.T) {
		// 返信の作成
		replyID := uuid.New()
		replyPost := &models.Post{
			ID:        replyID,
			UserID:    testUser.ID,
			Content:   "Reply content",
			ReplyToID: &testPost.ID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err := postRepo.Create(ctx, replyPost)
		require.NoError(t, err)

		// 返信の取得
		replies, err := postRepo.GetReplies(ctx, testPost.ID, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, replies)
		assert.Equal(t, replyID, replies[0].ID)

		// 返信数の確認
		count, err := postRepo.CountReplies(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// Repost機能のテスト
	t.Run("Repost", func(t *testing.T) {
		// リポストの作成
		repostID := uuid.New()
		repostPost := &models.Post{
			ID:        repostID,
			UserID:    testUser.ID,
			Content:   "Repost comment",
			RepostID:  &testPost.ID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err := postRepo.Create(ctx, repostPost)
		require.NoError(t, err)

		// リポストの取得
		reposts, err := postRepo.GetReposts(ctx, testPost.ID, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, reposts)
		assert.Equal(t, repostID, reposts[0].ID)

		// リポスト数の確認
		count, err := postRepo.CountReposts(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	// カウント機能のテスト
	t.Run("Counts", func(t *testing.T) {
		// いいね数の増加
		err := postRepo.IncrementLikeCount(ctx, testPost.ID)
		require.NoError(t, err)

		// リポスト数の増加
		err = postRepo.IncrementRepostCount(ctx, testPost.ID)
		require.NoError(t, err)

		// 返信数の増加
		err = postRepo.IncrementReplyCount(ctx, testPost.ID)
		require.NoError(t, err)

		// 数値の確認
		post, err := postRepo.GetByID(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, post.LikeCount)
		assert.Equal(t, 1, post.RepostCount)
		assert.Equal(t, 1, post.ReplyCount)

		// いいね数の減少
		err = postRepo.DecrementLikeCount(ctx, testPost.ID)
		require.NoError(t, err)

		// リポスト数の減少
		err = postRepo.DecrementRepostCount(ctx, testPost.ID)
		require.NoError(t, err)

		// 返信数の減少
		err = postRepo.DecrementReplyCount(ctx, testPost.ID)
		require.NoError(t, err)

		// 減少後の数値確認
		post, err = postRepo.GetByID(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, post.LikeCount)
		assert.Equal(t, 0, post.RepostCount)
		assert.Equal(t, 0, post.ReplyCount)

		// 負の値にならないことを確認
		err = postRepo.DecrementLikeCount(ctx, testPost.ID)
		require.NoError(t, err)
		post, err = postRepo.GetByID(ctx, testPost.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, post.LikeCount)
	})

	// インクリメント/デクリメントのエラーケースを追加
	t.Run("CountOperationsErrors", func(t *testing.T) {
		nonexistentID := uuid.New()

		// 存在しない投稿へのいいね数操作
		err := postRepo.IncrementLikeCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")

		err = postRepo.DecrementLikeCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")

		// 存在しない投稿へのリポスト数操作
		err = postRepo.IncrementRepostCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")

		err = postRepo.DecrementRepostCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")

		// 存在しない投稿への返信数操作
		err = postRepo.IncrementReplyCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")

		err = postRepo.DecrementReplyCount(ctx, nonexistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")
	})

	// 異常系データのテスト
	t.Run("InvalidData", func(t *testing.T) {
		// 1. 最大長を超えるコンテンツ
		hugeContent := make([]byte, 5000) // PostgreSQLのtext型の一般的な制限よりも大きい
		for i := range hugeContent {
			hugeContent[i] = 'a'
		}

		invalidPost := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   string(hugeContent),
			MediaURLs: []string{"valid.jpg"},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		// 長すぎるコンテンツでの作成を試みる
		err := postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for oversized content")

		// 2. 不正な構造のメディアURL
		invalidPost = &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   "Valid content",
			MediaURLs: make([]string, 1000), // 大量のempty strings
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for too many media URLs")

		// 3. 存在しないユーザーIDでの作成
		invalidPost = &models.Post{
			ID:        uuid.New(),
			UserID:    uuid.New(), // 存在しないユーザーID
			Content:   "Valid content",
			MediaURLs: []string{"valid.jpg"},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for non-existent user ID")
	})

	// Count関数のエラーケース
	t.Run("CountErrors", func(t *testing.T) {
		nonexistentID := uuid.New()

		// 存在しない投稿の返信数
		count, err := postRepo.CountReplies(ctx, nonexistentID)
		assert.NoError(t, err) // エラーは期待されない
		assert.Equal(t, int64(0), count)

		// 存在しない投稿のリポスト数
		count, err = postRepo.CountReposts(ctx, nonexistentID)
		assert.NoError(t, err) // エラーは期待されない
		assert.Equal(t, int64(0), count)

		// 存在しないユーザーの投稿数
		count, err = postRepo.CountByUserID(ctx, nonexistentID)
		assert.NoError(t, err) // エラーは期待されない
		assert.Equal(t, int64(0), count)
	})

	// List のテスト
	t.Run("List", func(t *testing.T) {
		posts, err := postRepo.List(ctx, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, posts)
	})

	// Delete のテスト
	t.Run("Delete", func(t *testing.T) {
		err := postRepo.Delete(ctx, testPost.ID)
		require.NoError(t, err)

		// 削除されたことを確認
		_, err = postRepo.GetByID(ctx, testPost.ID)
		assert.Error(t, err)
	})

	// データ制約のテスト
	t.Run("DataConstraints", func(t *testing.T) {
		// 1. 280文字を超えるコンテンツ
		longContent := make([]byte, 281)
		for i := range longContent {
			longContent[i] = 'a'
		}

		invalidPost := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   string(longContent),
			MediaURLs: []string{"valid.jpg"},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		// 文字数制限違反での作成を試みる
		err := postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for content exceeding 280 characters")
		assert.Contains(t, err.Error(), "280 characters", "Expected content length violation error")

		// 2. 空のコンテンツ
		invalidPost = &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   "",
			MediaURLs: []string{"valid.jpg"},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for empty content")
		assert.Contains(t, err.Error(), "empty", "Expected empty content error")

		// 3. メディアURL制限違反
		mediaURLs := make([]string, 5)
		for i := range mediaURLs {
			mediaURLs[i] = fmt.Sprintf("image%d.jpg", i)
		}

		invalidPost = &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   "Valid content",
			MediaURLs: mediaURLs,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for too many media URLs")
		assert.Contains(t, err.Error(), "media URLs", "Expected media URLs limit error")

		// 4. 外部キー制約違反
		invalidPost = &models.Post{
			ID:        uuid.New(),
			UserID:    uuid.New(), // 存在しないユーザーID
			Content:   "Valid content",
			MediaURLs: []string{"valid.jpg"},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = postRepo.Create(ctx, invalidPost)
		assert.Error(t, err, "Expected error for non-existent user ID")
	})
}

func TestPostRepository_Counts(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	userRepo := NewUserRepository(db.Pool)
	postRepo := NewPostRepository(db.Pool)

	ctx := context.Background()
	testUser := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Name:      "Test User",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// 投稿数のテスト
	t.Run("CountByUserID", func(t *testing.T) {
		// 初期状態のカウント
		count, err := postRepo.CountByUserID(ctx, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// 投稿を追加
		post1 := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   "Post 1",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = postRepo.Create(ctx, post1)
		require.NoError(t, err)

		post2 := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Content:   "Post 2",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = postRepo.Create(ctx, post2)
		require.NoError(t, err)

		// カウントの確認
		count, err = postRepo.CountByUserID(ctx, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 存在しないユーザーIDのカウント
		count, err = postRepo.CountByUserID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
