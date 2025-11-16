package entity

import (
	"fmt"
	"time"
)

// Team представляет команду разработчиков в доменной модели
type Team struct {
	name      string
	createdAt time.Time
	updatedAt time.Time
}

// NewTeam создаёт новую команду с валидацией
func NewTeam(name string) (*Team, error) {
	normalizedName, err := validateAndNormalizeTeamName(name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTeamName, err)
	}

	now := time.Now().UTC()

	return &Team{
		name:      normalizedName,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// NewTeamFromRepository восстанавливает команду из хранилища без валидации
func NewTeamFromRepository(
	name string,
	createdAt time.Time,
	updatedAt time.Time,
) *Team {
	return &Team{
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t *Team) Name() string {
	return t.name
}

func (t *Team) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Team) UpdatedAt() time.Time {
	return t.updatedAt
}

// Equals сравнивает две команды по имени
func (t *Team) Equals(other *Team) bool {
	if other == nil {
		return false
	}
	return t.name == other.name
}
