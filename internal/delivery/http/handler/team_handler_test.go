package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

type mockTeamUseCase struct {
	createTeam            func(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error)
	getTeam               func(ctx context.Context, teamName string) (*dto.TeamDTO, error)
	deactivateTeamMembers func(ctx context.Context, teamName string) (*dto.TeamDTO, error)
}

func (m *mockTeamUseCase) CreateTeam(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error) {
	return m.createTeam(ctx, req)
}

func (m *mockTeamUseCase) GetTeam(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
	return m.getTeam(ctx, teamName)
}

func (m *mockTeamUseCase) DeactivateTeamMembers(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
	return m.deactivateTeamMembers(ctx, teamName)
}

func TestTeamHandler_CreateTeam(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.CreateTeamRequest
		setupMock  func() *mockTeamUseCase
		wantStatus int
	}{
		{
			name: "success",
			body: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
					{UserID: "user-2", Username: "User 2", IsActive: true},
				},
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					createTeam: func(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error) {
						return &dto.TeamDTO{
							TeamName: req.TeamName,
							Members: []dto.TeamMemberDTO{
								{UserID: "user-1", Username: "User 1", IsActive: true},
								{UserID: "user-2", Username: "User 2", IsActive: true},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid body",
			body: dto.CreateTeamRequest{},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty team_name",
			body: dto.CreateTeamRequest{
				TeamName: "",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
				},
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error - team already exists",
			body: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
				},
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					createTeam: func(ctx context.Context, req dto.CreateTeamRequest) (*dto.TeamDTO, error) {
						return nil, usecase.ErrTeamAlreadyExists
					},
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewTeamHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.CreateTeam(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestTeamHandler_GetTeam(t *testing.T) {
	tests := []struct {
		name       string
		teamName   string
		setupMock  func() *mockTeamUseCase
		wantStatus int
	}{
		{
			name:     "success",
			teamName: "team-1",
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					getTeam: func(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
						return &dto.TeamDTO{
							TeamName: teamName,
							Members: []dto.TeamMemberDTO{
								{UserID: "user-1", Username: "User 1", IsActive: true},
								{UserID: "user-2", Username: "User 2", IsActive: true},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "missing team_name parameter",
			teamName: "",
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "use case error - team not found",
			teamName: "nonexistent",
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					getTeam: func(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
						return nil, usecase.ErrTeamNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewTeamHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
			if tt.teamName != "" {
				q := req.URL.Query()
				q.Add("team_name", tt.teamName)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			handler.GetTeam(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestTeamHandler_DeactivateTeamMembers(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.DeactivateTeamMembersRequest
		setupMock  func() *mockTeamUseCase
		wantStatus int
	}{
		{
			name: "success",
			body: dto.DeactivateTeamMembersRequest{
				TeamName: "team-1",
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					deactivateTeamMembers: func(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
						return &dto.TeamDTO{
							TeamName: teamName,
							Members: []dto.TeamMemberDTO{
								{UserID: "user-1", Username: "User 1", IsActive: false},
								{UserID: "user-2", Username: "User 2", IsActive: false},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid body",
			body: dto.DeactivateTeamMembersRequest{},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty team_name",
			body: dto.DeactivateTeamMembersRequest{
				TeamName: "",
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error - team not found",
			body: dto.DeactivateTeamMembersRequest{
				TeamName: "nonexistent",
			},
			setupMock: func() *mockTeamUseCase {
				return &mockTeamUseCase{
					deactivateTeamMembers: func(ctx context.Context, teamName string) (*dto.TeamDTO, error) {
						return nil, usecase.ErrTeamNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewTeamHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/team/deactivateMembers", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.DeactivateTeamMembers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
