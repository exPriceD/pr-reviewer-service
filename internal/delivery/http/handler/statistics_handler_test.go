package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

type mockStatisticsUseCase struct {
	getStatistics func(ctx context.Context, teamName string) (*dto.StatisticsDTO, error)
}

func (m *mockStatisticsUseCase) GetStatistics(ctx context.Context, teamName string) (*dto.StatisticsDTO, error) {
	return m.getStatistics(ctx, teamName)
}

func TestStatisticsHandler_GetStatistics(t *testing.T) {
	tests := []struct {
		name       string
		teamName   string
		setupMock  func() *mockStatisticsUseCase
		wantStatus int
	}{
		{
			name:     "success - all statistics",
			teamName: "",
			setupMock: func() *mockStatisticsUseCase {
				return &mockStatisticsUseCase{
					getStatistics: func(ctx context.Context, teamName string) (*dto.StatisticsDTO, error) {
						return &dto.StatisticsDTO{
							PRStats: dto.PRStatsDTO{
								Total:  10,
								Open:   5,
								Merged: 5,
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "success - team statistics",
			teamName: "team-1",
			setupMock: func() *mockStatisticsUseCase {
				return &mockStatisticsUseCase{
					getStatistics: func(ctx context.Context, teamName string) (*dto.StatisticsDTO, error) {
						return &dto.StatisticsDTO{
							PRStats: dto.PRStatsDTO{
								Total:  10,
								Open:   5,
								Merged: 5,
							},
							UserStats: []dto.UserStatsDTO{
								{UserID: "user-1", TotalReviews: 3, ActiveReviews: 2},
								{UserID: "user-2", TotalReviews: 2, ActiveReviews: 1},
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "success - team with empty stats",
			teamName: "team-1",
			setupMock: func() *mockStatisticsUseCase {
				return &mockStatisticsUseCase{
					getStatistics: func(ctx context.Context, teamName string) (*dto.StatisticsDTO, error) {
						return &dto.StatisticsDTO{
							PRStats: dto.PRStatsDTO{
								Total:  10,
								Open:   5,
								Merged: 5,
							},
							UserStats: []dto.UserStatsDTO{},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewStatisticsHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
			if tt.teamName != "" {
				q := req.URL.Query()
				q.Add("team_name", tt.teamName)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			handler.GetStatistics(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
