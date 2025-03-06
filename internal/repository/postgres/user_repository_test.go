package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	testing_helper "github.com/TakuyaAizawa/gox/internal/repository/postgres/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	repo := NewUserRepository(db.Pool)
	ctx := context.Background()

	// テストユーザーの作成
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

	// Create のテスト
	t.Run("Create", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)

		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		// 重複ユーザーの作成を試みる
		duplicateUser := &models.User{
			ID:       uuid.New(),
			Username: testUser.Username,
			Email:    testUser.Email,
		}
		err = repo.Create(ctx, duplicateUser)
		assert.Error(t, err)
	})

	// GetByID のテスト
	t.Run("GetByID", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)
		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		user, err := repo.GetByID(ctx, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)

		// 存在しないIDでの取得を試みる
		_, err = repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
	})

	// GetByUsername のテスト
	t.Run("GetByUsername", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)
		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		user, err := repo.GetByUsername(ctx, testUser.Username)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)

		// 存在しないユーザー名での取得を試みる
		_, err = repo.GetByUsername(ctx, "nonexistent")
		assert.Error(t, err)
	})

	// GetByEmail のテスト
	t.Run("GetByEmail", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)
		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		user, err := repo.GetByEmail(ctx, testUser.Email)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Email, user.Email)

		// 存在しないメールアドレスでの取得を試みる
		_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})

	// Update のテスト
	t.Run("Update", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)
		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		testUser.Bio = "Updated bio"
		testUser.Name = "Updated Name"
		err = repo.Update(ctx, testUser)
		require.NoError(t, err)

		// 更新された情報を確認
		updated, err := repo.GetByID(ctx, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated bio", updated.Bio)
		assert.Equal(t, "Updated Name", updated.Name)

		// 存在しないユーザーの更新を試みる
		nonexistentUser := &models.User{
			ID:       uuid.New(),
			Username: "nonexistent",
			Email:    "nonexistent@example.com",
		}
		err = repo.Update(ctx, nonexistentUser)
		assert.Error(t, err)

		// 重複するユーザー名での更新を試みる
		duplicateUser := &models.User{
			ID:        uuid.New(),
			Username:  "uniqueuser",
			Email:     "unique@example.com",
			Password:  "hashedpassword",
			Name:      "Unique User",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = repo.Create(ctx, duplicateUser)
		require.NoError(t, err)

		duplicateUser.Username = testUser.Username
		err = repo.Update(ctx, duplicateUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	// List のテスト
	t.Run("List", func(t *testing.T) {
		// テスト前にクリーンアップ
		db.CleanupAllTables(t)

		// 単一のユーザーを作成
		err := repo.Create(ctx, testUser)
		require.NoError(t, err)

		users, err := repo.List(ctx, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Len(t, users, 1)
	})

	// Search のテスト
	t.Run("Search", func(t *testing.T) {
		// ユーザー名で検索
		users, err := repo.Search(ctx, "test", 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, users)

		// 存在しない検索語で検索
		users, err = repo.Search(ctx, "nonexistent", 0, 10)
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	// IsUsernameAvailable のテスト
	t.Run("IsUsernameAvailable", func(t *testing.T) {
		// 既存のユーザー名をチェック
		available, err := repo.IsUsernameAvailable(ctx, testUser.Username)
		require.NoError(t, err)
		assert.False(t, available)

		// 利用可能なユーザー名をチェック
		available, err = repo.IsUsernameAvailable(ctx, "availableusername")
		require.NoError(t, err)
		assert.True(t, available)
	})

	// IsEmailAvailable のテスト
	t.Run("IsEmailAvailable", func(t *testing.T) {
		// 既存のメールアドレスをチェック
		available, err := repo.IsEmailAvailable(ctx, testUser.Email)
		require.NoError(t, err)
		assert.False(t, available)

		// 利用可能なメールアドレスをチェック
		available, err = repo.IsEmailAvailable(ctx, "available@example.com")
		require.NoError(t, err)
		assert.True(t, available)
	})

	// Delete のテスト
	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, testUser.ID)
		require.NoError(t, err)

		// 削除されたことを確認
		_, err = repo.GetByID(ctx, testUser.ID)
		assert.Error(t, err)
	})
}

func TestUserRepository_Count(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	repo := NewUserRepository(db.Pool)
	ctx := context.Background()

	// 初期状態でのカウント
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// ユーザーを追加
	user1 := &models.User{
		ID:        uuid.New(),
		Username:  "user1",
		Email:     "user1@example.com",
		Password:  "hashedpassword",
		Name:      "User 1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = repo.Create(ctx, user1)
	require.NoError(t, err)

	// カウントの確認
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
