package dto

// CreateTeamRequest входные данные для создания команды
type CreateTeamRequest struct {
	TeamName string              `json:"team_name"`
	Members  []TeamMemberRequest `json:"members"`
}

// TeamMemberRequest данные участника команды
type TeamMemberRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// DeactivateTeamMembersRequest входные данные для массовой деактивации пользователей команды
type DeactivateTeamMembersRequest struct {
	TeamName string `json:"team_name"`
}
