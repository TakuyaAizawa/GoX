package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Password       string    `json:"-"` // パスワードはJSONにシリアライズしない
	Name           string    `json:"name"`
	Bio            string    `json:"bio"`
	ProfileImage   string    `json:"profile_image"`
	BannerImage    string    `json:"banner_image"`
	Location       string    `json:"location"`
	WebsiteURL     string    `json:"website_url"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	PostCount      int       `json:"post_count"`
	IsVerified     bool      `json:"is_verified"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewUser creates a new user with default values
func NewUser(username, email, password, name string) *User {
	now := time.Now()
	return &User{
		ID:             uuid.New(),
		Username:       username,
		Email:          email,
		Password:       password, // 注意: この段階ではハッシュ化されていない
		Name:           name,
		Bio:            "",
		ProfileImage:   "",
		BannerImage:    "",
		Location:       "",
		WebsiteURL:     "",
		FollowerCount:  0,
		FollowingCount: 0,
		PostCount:      0,
		IsVerified:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UserResponse represents the user data sent to clients
type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Bio            string    `json:"bio"`
	ProfileImage   string    `json:"profile_image"`
	BannerImage    string    `json:"banner_image"`
	Location       string    `json:"location"`
	WebsiteURL     string    `json:"website_url"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	PostCount      int       `json:"post_count"`
	IsVerified     bool      `json:"is_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Name:           u.Name,
		Bio:            u.Bio,
		ProfileImage:   u.ProfileImage,
		BannerImage:    u.BannerImage,
		Location:       u.Location,
		WebsiteURL:     u.WebsiteURL,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		PostCount:      u.PostCount,
		IsVerified:     u.IsVerified,
		CreatedAt:      u.CreatedAt,
	}
}
