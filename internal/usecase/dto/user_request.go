package dto

// SetUserActiveRequest входные данные для установки активности пользователя
type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}
