package dto

// TeamDTO представляет команду с участниками для HTTP ответа
type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

// TeamMemberDTO представляет участника команды для HTTP ответа
type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
