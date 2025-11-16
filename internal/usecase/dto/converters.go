package dto

import (
	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
)

// ToPullRequestDTO конвертирует entity.PullRequest в PullRequestDTO
func ToPullRequestDTO(pr *entity.PullRequest) PullRequestDTO {
	return PullRequestDTO{
		PullRequestID:     pr.ID(),
		PullRequestName:   pr.Name(),
		AuthorID:          pr.AuthorID(),
		Status:            string(pr.Status()),
		AssignedReviewers: pr.AssignedReviewers(),
		CreatedAt:         pr.CreatedAt(),
		MergedAt:          pr.MergedAt(),
	}
}

// ToPullRequestShortDTO конвертирует entity.PullRequest в PullRequestShortDTO
func ToPullRequestShortDTO(pr *entity.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.ID(),
		PullRequestName: pr.Name(),
		AuthorID:        pr.AuthorID(),
		Status:          string(pr.Status()),
	}
}

// ToPullRequestShortDTOs конвертирует слайс entity.PullRequest в слайс PullRequestShortDTO
func ToPullRequestShortDTOs(prs []*entity.PullRequest) []PullRequestShortDTO {
	result := make([]PullRequestShortDTO, len(prs))
	for i, pr := range prs {
		result[i] = ToPullRequestShortDTO(pr)
	}
	return result
}

// ToUserDTO конвертирует entity.User в UserDTO
func ToUserDTO(user *entity.User) UserDTO {
	return UserDTO{
		UserID:   user.ID(),
		Username: user.Username(),
		TeamName: user.TeamName(),
		IsActive: user.IsActive(),
	}
}

// ToTeamMemberDTO конвертирует entity.User в TeamMemberDTO
func ToTeamMemberDTO(user *entity.User) TeamMemberDTO {
	return TeamMemberDTO{
		UserID:   user.ID(),
		Username: user.Username(),
		IsActive: user.IsActive(),
	}
}

// ToTeamMemberDTOs конвертирует слайс entity.User в слайс TeamMemberDTO
func ToTeamMemberDTOs(users []*entity.User) []TeamMemberDTO {
	result := make([]TeamMemberDTO, len(users))
	for i, user := range users {
		result[i] = ToTeamMemberDTO(user)
	}
	return result
}

// ToTeamDTO конвертирует entity.Team и слайс entity.User в TeamDTO
func ToTeamDTO(team *entity.Team, members []*entity.User) TeamDTO {
	return TeamDTO{
		TeamName: team.Name(),
		Members:  ToTeamMemberDTOs(members),
	}
}
