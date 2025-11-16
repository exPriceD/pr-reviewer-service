package team

import "github.com/exPriceD/pr-reviewer-service/internal/domain/entity"

func ToEntity(m *Model) *entity.Team {
	return entity.NewTeamFromRepository(
		m.Name,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func FromEntity(t *entity.Team) *Model {
	return &Model{
		Name:      t.Name(),
		CreatedAt: t.CreatedAt(),
		UpdatedAt: t.UpdatedAt(),
	}
}
