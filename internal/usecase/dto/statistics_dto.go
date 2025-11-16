package dto

// StatisticsDTO статистика по назначениям
type StatisticsDTO struct {
	PRStats   PRStatsDTO     `json:"pr_stats"`
	UserStats []UserStatsDTO `json:"user_stats,omitempty"`
}

// PRStatsDTO статистика по Pull Requests
type PRStatsDTO struct {
	Total  int `json:"total"`
	Open   int `json:"open"`
	Merged int `json:"merged"`
}

// UserStatsDTO статистика по пользователю
type UserStatsDTO struct {
	UserID        string `json:"user_id"`
	TotalReviews  int    `json:"total_reviews"`
	ActiveReviews int    `json:"active_reviews"`
}
