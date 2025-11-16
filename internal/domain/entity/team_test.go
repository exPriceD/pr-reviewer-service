package entity

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestNewTeam проверяет создание новой команды
func TestNewTeam(t *testing.T) {
	tests := []struct {
		name        string
		teamName    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid team",
			teamName: "backend-team",
			wantErr:  false,
		},
		{
			name:     "valid team with underscores",
			teamName: "frontend_team",
			wantErr:  false,
		},
		{
			name:     "valid team short name",
			teamName: "qa",
			wantErr:  false,
		},
		{
			name:     "valid team with spaces trimmed",
			teamName: "  devops-team  ",
			wantErr:  false,
		},
		{
			name:        "empty team name",
			teamName:    "",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
		{
			name:        "team name with only spaces",
			teamName:    "   ",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
		{
			name:        "team name too long",
			teamName:    strings.Repeat("a", 101),
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
		{
			name:        "team name with invalid characters",
			teamName:    "team@name",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
		{
			name:        "team name with spaces inside",
			teamName:    "team name",
			wantErr:     true,
			expectedErr: ErrInvalidTeamName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := NewTeam(tt.teamName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTeam() expected error, got nil")
					return
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("NewTeam() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewTeam() unexpected error: %v", err)
				return
			}

			expectedName := strings.TrimSpace(tt.teamName)
			if team.Name() != expectedName {
				t.Errorf("Name = %v, want %v", team.Name(), expectedName)
			}

			if team.CreatedAt().IsZero() {
				t.Errorf("CreatedAt is zero")
			}

			if team.UpdatedAt().IsZero() {
				t.Errorf("UpdatedAt is zero")
			}

			if !team.CreatedAt().Equal(team.UpdatedAt()) {
				t.Errorf("CreatedAt != UpdatedAt for new team")
			}
		})
	}
}

// TestNewTeamFromRepository проверяет восстановление команды из БД
func TestNewTeamFromRepository(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	team := NewTeamFromRepository("backend-team", createdAt, updatedAt)

	if team.Name() != "backend-team" {
		t.Errorf("Name = %v, want backend-team", team.Name())
	}

	if !team.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", team.CreatedAt(), createdAt)
	}

	if !team.UpdatedAt().Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", team.UpdatedAt(), updatedAt)
	}
}

// TestTeamEquals проверяет сравнение команд
func TestTeamEquals(t *testing.T) {
	team1, _ := NewTeam("backend")
	team2, _ := NewTeam("backend")
	team3, _ := NewTeam("frontend")

	if !team1.Equals(team2) {
		t.Errorf("team1.Equals(team2) = false, want true (same name)")
	}

	if team1.Equals(team3) {
		t.Errorf("team1.Equals(team3) = true, want false (different names)")
	}

	if team1.Equals(nil) {
		t.Errorf("team1.Equals(nil) = true, want false")
	}

	//nolint:gocritic // Намеренное сравнение объекта с самим собой
	if !team1.Equals(team1) {
		t.Errorf("team1.Equals(team1) = false, want true (same object)")
	}
}

// TestTeamGetters проверяет все геттеры
func TestTeamGetters(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	team := NewTeamFromRepository("payments-team", createdAt, updatedAt)

	if got := team.Name(); got != "payments-team" {
		t.Errorf("Name() = %v, want payments-team", got)
	}

	if got := team.CreatedAt(); !got.Equal(createdAt) {
		t.Errorf("CreatedAt() = %v, want %v", got, createdAt)
	}

	if got := team.UpdatedAt(); !got.Equal(updatedAt) {
		t.Errorf("UpdatedAt() = %v, want %v", got, updatedAt)
	}
}

// TestTeamValidNames проверяет различные валидные форматы имён команд
func TestTeamValidNames(t *testing.T) {
	validNames := []string{
		"backend",
		"frontend-team",
		"qa_automation",
		"team-123",
		"DevOps",
		"mobile-iOS",
		"team_1",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			team, err := NewTeam(name)
			if err != nil {
				t.Errorf("NewTeam(%q) failed: %v", name, err)
			}
			if team == nil {
				t.Errorf("NewTeam(%q) returned nil team", name)
			}
		})
	}
}

// TestTeamInvalidNames проверяет различные невалидные форматы имён команд
func TestTeamInvalidNames(t *testing.T) {
	invalidNames := []string{
		"",             // пустое
		"   ",          // только пробелы
		"team name",    // пробел внутри
		"team@company", // @
		"team.name",    // точка
		"team#1",       // #
		"team$name",    // $
		"team%name",    // %
		"team&name",    // &
		"team*name",    // *
		"team(name)",   // скобки
		"team[name]",   // квадратные скобки
		"team{name}",   // фигурные скобки
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			team, err := NewTeam(name)
			if err == nil {
				t.Errorf("NewTeam(%q) expected error, got nil", name)
			}
			if team != nil {
				t.Errorf("NewTeam(%q) expected nil team, got %v", name, team)
			}
			if !errors.Is(err, ErrInvalidTeamName) {
				t.Errorf("NewTeam(%q) error = %v, want ErrInvalidTeamName", name, err)
			}
		})
	}
}
