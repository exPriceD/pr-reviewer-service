package entity

import (
	"fmt"
	"time"
)

// User представляет участника команды в доменной модели
type User struct {
	id        string
	username  string
	teamName  string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

// NewUser создаёт нового пользователя с валидацией всех полей
func NewUser(id, username, teamName string) (*User, error) {
	normalizedID, err := validateAndNormalizeID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidID, err)
	}

	normalizedUsername, err := validateAndNormalizeUsername(username)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidUsername, err)
	}

	normalizedTeamName, err := validateAndNormalizeTeamName(teamName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTeamName, err)
	}

	now := time.Now().UTC()

	return &User{
		id:        normalizedID,
		username:  normalizedUsername,
		teamName:  normalizedTeamName,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// NewUserFromRepository восстанавливает пользователя из хранилища без валидации
func NewUserFromRepository(
	id string,
	username string,
	teamName string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:        id,
		username:  username,
		teamName:  teamName,
		isActive:  isActive,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) TeamName() string {
	return u.teamName
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Deactivate деактивирует пользователя.
func (u *User) Deactivate() bool {
	if !u.isActive {
		return false
	}
	u.isActive = false
	u.updatedAt = time.Now().UTC()
	return true
}

// Activate активирует пользователя.
func (u *User) Activate() bool {
	if u.isActive {
		return false
	}
	u.isActive = true
	u.updatedAt = time.Now().UTC()
	return true
}

// ChangeTeam переносит пользователя в другую команду
func (u *User) ChangeTeam(newTeamName string) error {
	normalizedTeamName, err := validateAndNormalizeTeamName(newTeamName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTeamName, err)
	}

	if u.teamName == normalizedTeamName {
		return ErrNoChange
	}

	u.teamName = normalizedTeamName
	u.updatedAt = time.Now().UTC()
	return nil
}

// ChangeUsername обновляет имя пользователя
func (u *User) ChangeUsername(newUsername string) error {
	normalizedUsername, err := validateAndNormalizeUsername(newUsername)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidUsername, err)
	}

	if u.username == normalizedUsername {
		return ErrNoChange
	}

	u.username = normalizedUsername
	u.updatedAt = time.Now().UTC()
	return nil
}

// Equals сравнивает двух пользователей по идентификатору
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.id == other.id
}
