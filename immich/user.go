package immich

import (
	"context"
	"time"
)

type User struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	Name                 string    `json:"name"`
	ProfileImagePath     string    `json:"profileImagePath,omitempty"`
	AvatarColor          string    `json:"avatarColor"`
	ProfileChangedAt     time.Time `json:"profileChangedAt"`
	StorageLabel         string    `json:"storageLabel,omitempty"`
	ShouldChangePassword bool      `json:"shouldChangePassword"`
	IsAdmin              bool      `json:"isAdmin"`
	CreatedAt            time.Time `json:"createdAt,omitempty"`
	DeletedAt            time.Time `json:"deletedAt,omitempty"`
	UpdatedAt            time.Time `json:"updatedAt,omitempty"`
	OauthID              string    `json:"oauthId,omitempty"`
	QuotaSizeInBytes     int64     `json:"quotaSizeInBytes,omitempty"`
	QuotaUsageInBytes    int64     `json:"quotaUsageInBytes,omitempty"`
	Status               string    `json:"status,omitempty"`
	License              string    `json:"license,omitempty"`
}

func (ic *ImmichClient) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := ic.newServerCall(ctx, EndPointGetAllUsers).
		do(
			getRequest("/admin/users", setAcceptJSON()),
			responseJSON(&users),
		)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (ic *ImmichClient) GetUserInfo(ctx context.Context, userID string) (User, error) {
	var user User
	err := ic.newServerCall(ctx, EndPointGetUserInfo).
		do(
			getRequest("/admin/users/"+userID, setAcceptJSON()),
			responseJSON(&user),
		)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
