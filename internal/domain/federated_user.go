package domain

type FederatedUser struct {
	UserID           string `json:"user_id" db:"user_id"`
	IdentityProvider string `json:"identity_provider" db:"identity_provider"`
	ExternalUserID   string `json:"external_user_id" db:"external_user_id"`
}
