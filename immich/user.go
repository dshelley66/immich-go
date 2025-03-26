package immich

import "time"

type User struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	Name                 string    `json:"name"`
	ProfileImagePath     string    `json:"profileImagePath"`
	AvatarColor          string    `json:"avatarColor"`
	ProfileChangedAt     time.Time `json:"profileChangedAt"`
	StorageLabel         string    `json:"storageLabel"`
	ShouldChangePassword bool      `json:"shouldChangePassword"`
	IsAdmin              bool      `json:"isAdmin"`
	CreatedAt            time.Time `json:"createdAt"`
	DeletedAt            time.Time `json:"deletedAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	OauthID              string    `json:"oauthId"`
	QuotaSizeInBytes     int64     `json:"quotaSizeInBytes"`
	QuotaUsageInBytes    int64     `json:"quotaUsageInBytes"`
	Status               string    `json:"status"`
	License              string    `json:"license"`
}
