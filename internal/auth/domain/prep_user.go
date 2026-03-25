package domain

// PrepUser is a value object representing user data from Prep User Service (/me endpoint).
// This app does NOT manage Prep users — Prep platform owns this data.
type PrepUser struct {
	PrepUserID          int64
	Email               string
	Name                string
	IsFirstLogin        bool
	ForceUpdatePassword bool
}

func NewPrepUser(prepUserID int64, email, name string, isFirstLogin, forceUpdatePassword bool) *PrepUser {
	return &PrepUser{
		PrepUserID:          prepUserID,
		Email:               email,
		Name:                name,
		IsFirstLogin:        isFirstLogin,
		ForceUpdatePassword: forceUpdatePassword,
	}
}
