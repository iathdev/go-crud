package dto

import "time"

type AuthMeResponse struct {
	ID                  string    `json:"id"`
	PrepUserID          int64     `json:"prep_user_id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	IsFirstLogin        bool      `json:"is_first_login"`
	ForceUpdatePassword bool      `json:"force_update_password"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
