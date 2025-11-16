package entity

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestNewUser проверяет создание нового пользователя
func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		username    string
		teamName    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid user",
			id:       "user-123",
			username: "John Doe",
			teamName: "backend-team",
			wantErr:  false,
		},
		{
			name:     "valid user with underscores",
			id:       "user_456",
			username: "Jane_Smith",
			teamName: "frontend_team",
			wantErr:  false,
		},
		{
			name:     "valid user with unicode in username",
			id:       "u1",
			username: "Иван Иванов",
			teamName: "payments",
			wantErr:  false,
		},
		{
			name:     "valid user with spaces trimmed",
			id:       "  user-789  ",
			username: "  Bob Johnson  ",
			teamName: "  devops  ",
			wantErr:  false,
		},
		{
			name:        "empty id",
			id:          "",
			username:    "John",
			teamName:    "team1",
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "empty username",
			id:          "u1",
			username:    "",
			teamName:    "team1",
			wantErr:     true,
			expectedErr: ErrInvalidUsername,
		},
		{
			name:        "empty team name",
			id:          "u1",
			username:    "John",
			teamName:    "",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
		{
			name:        "id too long",
			id:          strings.Repeat("a", 256),
			username:    "John",
			teamName:    "team1",
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "username too long",
			id:          "u1",
			username:    strings.Repeat("a", 101),
			teamName:    "team1",
			wantErr:     true,
			expectedErr: ErrInvalidUsername,
		},
		{
			name:        "id with invalid characters",
			id:          "user@123",
			username:    "John",
			teamName:    "team1",
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "team name with invalid characters",
			id:          "u1",
			username:    "John",
			teamName:    "team@name",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.id, tt.username, tt.teamName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() expected error, got nil")
					return
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("NewUser() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error: %v", err)
				return
			}

			expectedID := strings.TrimSpace(tt.id)
			expectedUsername := strings.TrimSpace(tt.username)
			expectedTeamName := strings.TrimSpace(tt.teamName)

			if user.ID() != expectedID {
				t.Errorf("ID = %v, want %v", user.ID(), expectedID)
			}
			if user.Username() != expectedUsername {
				t.Errorf("Username = %v, want %v", user.Username(), expectedUsername)
			}
			if user.TeamName() != expectedTeamName {
				t.Errorf("TeamName = %v, want %v", user.TeamName(), expectedTeamName)
			}
			if !user.IsActive() {
				t.Errorf("IsActive = false, want true")
			}
			if user.CreatedAt().IsZero() {
				t.Errorf("CreatedAt is zero")
			}
			if user.UpdatedAt().IsZero() {
				t.Errorf("UpdatedAt is zero")
			}
			if !user.CreatedAt().Equal(user.UpdatedAt()) {
				t.Errorf("CreatedAt != UpdatedAt for new user")
			}
		})
	}
}

// TestNewUserFromRepository проверяет восстановление пользователя из БД
func TestNewUserFromRepository(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	user := NewUserFromRepository(
		"user-1",
		"John Doe",
		"backend",
		false,
		createdAt,
		updatedAt,
	)

	if user.ID() != "user-1" {
		t.Errorf("ID = %v, want user-1", user.ID())
	}
	if user.Username() != "John Doe" {
		t.Errorf("Username = %v, want John Doe", user.Username())
	}
	if user.TeamName() != "backend" {
		t.Errorf("TeamName = %v, want backend", user.TeamName())
	}
	if user.IsActive() {
		t.Errorf("IsActive = true, want false")
	}
	if !user.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", user.CreatedAt(), createdAt)
	}
	if !user.UpdatedAt().Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", user.UpdatedAt(), updatedAt)
	}
}

// TestUserDeactivate проверяет деактивацию пользователя
func TestUserDeactivate(t *testing.T) {
	user, err := NewUser("u1", "John", "team1")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	changed := user.Deactivate()
	if !changed {
		t.Errorf("Deactivate() = false, want true on first call")
	}
	if user.IsActive() {
		t.Errorf("IsActive = true after Deactivate(), want false")
	}

	oldUpdatedAt := user.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	changed = user.Deactivate()
	if changed {
		t.Errorf("Deactivate() = true, want false on second call (idempotent)")
	}
	if user.IsActive() {
		t.Errorf("IsActive = true, want false")
	}

	if !user.UpdatedAt().Equal(oldUpdatedAt) {
		t.Errorf("UpdatedAt changed on idempotent Deactivate()")
	}
}

// TestUserActivate проверяет активацию пользователя
func TestUserActivate(t *testing.T) {
	user, _ := NewUser("u1", "John", "team1")
	user.Deactivate()

	changed := user.Activate()
	if !changed {
		t.Errorf("Activate() = false, want true when activating inactive user")
	}
	if !user.IsActive() {
		t.Errorf("IsActive = false after Activate(), want true")
	}

	oldUpdatedAt := user.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	changed = user.Activate()
	if changed {
		t.Errorf("Activate() = true, want false on second call (idempotent)")
	}
	if !user.IsActive() {
		t.Errorf("IsActive = false, want true")
	}
	if !user.UpdatedAt().Equal(oldUpdatedAt) {
		t.Errorf("UpdatedAt changed on idempotent Activate()")
	}
}

// TestUserChangeTeam проверяет смену команды
func TestUserChangeTeam(t *testing.T) {
	user, _ := NewUser("u1", "John", "team1")
	oldUpdatedAt := user.UpdatedAt()

	time.Sleep(10 * time.Millisecond)

	err := user.ChangeTeam("team2")
	if err != nil {
		t.Errorf("ChangeTeam() error = %v, want nil", err)
	}
	if user.TeamName() != "team2" {
		t.Errorf("TeamName = %v, want team2", user.TeamName())
	}
	if !user.UpdatedAt().After(oldUpdatedAt) {
		t.Errorf("UpdatedAt was not updated")
	}

	err = user.ChangeTeam("team2")
	if !errors.Is(err, ErrNoChange) {
		t.Errorf("ChangeTeam() with same team error = %v, want ErrNoChange", err)
	}

	err = user.ChangeTeam("team@invalid")
	if !errors.Is(err, ErrInvalidTeamName) {
		t.Errorf("ChangeTeam() with invalid name error = %v, want ErrInvalidTeamName", err)
	}

	err = user.ChangeTeam("")
	if !errors.Is(err, ErrInvalidTeamName) {
		t.Errorf("ChangeTeam() with empty name error = %v, want ErrInvalidTeamName", err)
	}

	err = user.ChangeTeam("  team3  ")
	if err != nil {
		t.Errorf("ChangeTeam() with spaces error = %v, want nil", err)
	}
	if user.TeamName() != "team3" {
		t.Errorf("TeamName = %v, want team3 (spaces should be trimmed)", user.TeamName())
	}
}

// TestUserChangeUsername проверяет смену имени пользователя
func TestUserChangeUsername(t *testing.T) {
	user, _ := NewUser("u1", "John", "team1")
	oldUpdatedAt := user.UpdatedAt()

	time.Sleep(10 * time.Millisecond)

	err := user.ChangeUsername("Jane Doe")
	if err != nil {
		t.Errorf("ChangeUsername() error = %v, want nil", err)
	}
	if user.Username() != "Jane Doe" {
		t.Errorf("Username = %v, want Jane Doe", user.Username())
	}
	if !user.UpdatedAt().After(oldUpdatedAt) {
		t.Errorf("UpdatedAt was not updated")
	}

	err = user.ChangeUsername("Jane Doe")
	if !errors.Is(err, ErrNoChange) {
		t.Errorf("ChangeUsername() with same name error = %v, want ErrNoChange", err)
	}

	err = user.ChangeUsername("")
	if !errors.Is(err, ErrInvalidUsername) {
		t.Errorf("ChangeUsername() with empty name error = %v, want ErrInvalidUsername", err)
	}

	err = user.ChangeUsername(strings.Repeat("a", 101))
	if !errors.Is(err, ErrInvalidUsername) {
		t.Errorf("ChangeUsername() with long name error = %v, want ErrInvalidUsername", err)
	}

	err = user.ChangeUsername("  Bob Smith  ")
	if err != nil {
		t.Errorf("ChangeUsername() with spaces error = %v, want nil", err)
	}
	if user.Username() != "Bob Smith" {
		t.Errorf("Username = %v, want Bob Smith (spaces should be trimmed)", user.Username())
	}
}

// TestUserEquals проверяет сравнение пользователей
func TestUserEquals(t *testing.T) {
	user1, _ := NewUser("u1", "John", "team1")
	user2, _ := NewUser("u1", "Jane", "team2") // тот же ID, другие поля
	user3, _ := NewUser("u2", "John", "team1")

	if !user1.Equals(user2) {
		t.Errorf("user1.Equals(user2) = false, want true (same ID)")
	}

	if user1.Equals(user3) {
		t.Errorf("user1.Equals(user3) = true, want false (different ID)")
	}

	if user1.Equals(nil) {
		t.Errorf("user1.Equals(nil) = true, want false")
	}

	//nolint:gocritic // Намеренное сравнение объекта с самим собой
	if !user1.Equals(user1) {
		t.Errorf("user1.Equals(user1) = false, want true (same object)")
	}
}

// TestUserGetters проверяет все геттеры
func TestUserGetters(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	user := NewUserFromRepository(
		"user-123",
		"John Doe",
		"backend-team",
		true,
		createdAt,
		updatedAt,
	)

	if got := user.ID(); got != "user-123" {
		t.Errorf("ID() = %v, want user-123", got)
	}
	if got := user.Username(); got != "John Doe" {
		t.Errorf("Username() = %v, want John Doe", got)
	}
	if got := user.TeamName(); got != "backend-team" {
		t.Errorf("TeamName() = %v, want backend-team", got)
	}
	if got := user.IsActive(); got != true {
		t.Errorf("IsActive() = %v, want true", got)
	}
	if got := user.CreatedAt(); !got.Equal(createdAt) {
		t.Errorf("CreatedAt() = %v, want %v", got, createdAt)
	}
	if got := user.UpdatedAt(); !got.Equal(updatedAt) {
		t.Errorf("UpdatedAt() = %v, want %v", got, updatedAt)
	}
}

// TestUserLifecycle проверяет полный жизненный цикл пользователя
func TestUserLifecycle(t *testing.T) {
	user, err := NewUser("u1", "John Doe", "backend")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if !user.IsActive() {
		t.Errorf("New user should be active")
	}

	if !user.Deactivate() {
		t.Errorf("First deactivation should return true")
	}

	if err := user.ChangeTeam("frontend"); err != nil {
		t.Errorf("Failed to change team: %v", err)
	}

	if err := user.ChangeUsername("Jane Doe"); err != nil {
		t.Errorf("Failed to change username: %v", err)
	}

	if !user.Activate() {
		t.Errorf("Activation should return true")
	}

	if user.ID() != "u1" {
		t.Errorf("ID changed unexpectedly")
	}
	if user.Username() != "Jane Doe" {
		t.Errorf("Username = %v, want Jane Doe", user.Username())
	}
	if user.TeamName() != "frontend" {
		t.Errorf("TeamName = %v, want frontend", user.TeamName())
	}
	if !user.IsActive() {
		t.Errorf("User should be active")
	}
}
