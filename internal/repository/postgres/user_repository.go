package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL implementation of UserRepository
func NewUserRepository(db *pgxpool.Pool) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.Password, user.Name,
		user.Bio, user.ProfileImage, user.FollowerCount, user.FollowingCount,
		user.PostCount, user.IsVerified, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		// Unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("user with this username or email already exists")
		}
		return err
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		FROM users WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Name,
		&user.Bio, &user.ProfileImage, &user.FollowerCount, &user.FollowingCount,
		&user.PostCount, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		FROM users WHERE username = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Name,
		&user.Bio, &user.ProfileImage, &user.FollowerCount, &user.FollowingCount,
		&user.PostCount, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Name,
		&user.Bio, &user.ProfileImage, &user.FollowerCount, &user.FollowingCount,
		&user.PostCount, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			username = $1, email = $2, name = $3, bio = $4,
			profile_image = $5, follower_count = $6, following_count = $7,
			post_count = $8, is_verified = $9, updated_at = $10
		WHERE id = $11
	`

	result, err := r.db.Exec(ctx, query,
		user.Username, user.Email, user.Name, user.Bio,
		user.ProfileImage, user.FollowerCount, user.FollowingCount,
		user.PostCount, user.IsVerified, user.UpdatedAt, user.ID,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("user with this username or email already exists")
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Name,
			&user.Bio, &user.ProfileImage, &user.FollowerCount, &user.FollowingCount,
			&user.PostCount, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Search(ctx context.Context, query string, offset, limit int) ([]*models.User, error) {
	sqlQuery := `
		SELECT id, username, email, password, name, bio, profile_image,
			follower_count, following_count, post_count, is_verified,
			created_at, updated_at
		FROM users
		WHERE username ILIKE $1 OR name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, sqlQuery, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Name,
			&user.Bio, &user.ProfileImage, &user.FollowerCount, &user.FollowingCount,
			&user.PostCount, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func (r *userRepository) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateAvatar updates the avatar URL for a user
func (r *userRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `
		UPDATE users 
		SET profile_image = $1, updated_at = NOW() 
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, avatarURL, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateBanner updates the banner URL for a user
func (r *userRepository) UpdateBanner(ctx context.Context, userID uuid.UUID, bannerURL string) error {
	query := `
		UPDATE users 
		SET banner_image = $1, updated_at = NOW() 
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, bannerURL, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
