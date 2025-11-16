package user

import "github.com/exPriceD/pr-reviewer-service/internal/domain/entity"

func ToEntity(m *Model) *entity.User {
	return entity.NewUserFromRepository(
		m.ID,
		m.Username,
		m.TeamName,
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func FromEntity(u *entity.User) *Model {
	return &Model{
		ID:        u.ID(),
		Username:  u.Username(),
		TeamName:  u.TeamName(),
		IsActive:  u.IsActive(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}
