package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/entity"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

type mockUserUseCase struct {
	setUserActive  func(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error)
	getUserReviews func(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error)
}

func (m *mockUserUseCase) SetUserActive(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error) {
	return m.setUserActive(ctx, req)
}

func (m *mockUserUseCase) GetUserReviews(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
	return m.getUserReviews(ctx, userID)
}

func TestUserHandler_SetUserActive(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.SetUserActiveRequest
		setupMock  func() *mockUserUseCase
		wantStatus int
	}{
		{
			name: "success - activate user",
			body: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					setUserActive: func(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error) {
						return &dto.UserDTO{
							UserID:   req.UserID,
							Username: "User 1",
							TeamName: "team-1",
							IsActive: req.IsActive,
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success - deactivate user",
			body: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: false,
			},
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					setUserActive: func(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error) {
						return &dto.UserDTO{
							UserID:   req.UserID,
							Username: "User 1",
							TeamName: "team-1",
							IsActive: req.IsActive,
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid body",
			body: dto.SetUserActiveRequest{},
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - empty user_id",
			body: dto.SetUserActiveRequest{
				UserID:   "",
				IsActive: true,
			},
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error - user not found",
			body: dto.SetUserActiveRequest{
				UserID:   "nonexistent",
				IsActive: true,
			},
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					setUserActive: func(ctx context.Context, req dto.SetUserActiveRequest) (*dto.UserDTO, error) {
						return nil, usecase.ErrUserNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewUserHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.SetUserActive(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestUserHandler_GetUserReviews(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMock  func() *mockUserUseCase
		wantStatus int
	}{
		{
			name:   "success",
			userID: "user-1",
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					getUserReviews: func(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
						return []dto.PullRequestShortDTO{
							{
								PullRequestID:   "pr-1",
								PullRequestName: "Test PR 1",
								AuthorID:        "author-1",
								Status:          string(entity.PRStatusOpen),
							},
							{
								PullRequestID:   "pr-2",
								PullRequestName: "Test PR 2",
								AuthorID:        "author-2",
								Status:          string(entity.PRStatusMerged),
							},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "success - no reviews",
			userID: "user-1",
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					getUserReviews: func(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
						return []dto.PullRequestShortDTO{}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "missing user_id parameter",
			userID: "",
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "use case error - user not found",
			userID: "nonexistent",
			setupMock: func() *mockUserUseCase {
				return &mockUserUseCase{
					getUserReviews: func(ctx context.Context, userID string) ([]dto.PullRequestShortDTO, error) {
						return nil, usecase.ErrUserNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewUserHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
			if tt.userID != "" {
				q := req.URL.Query()
				q.Add("user_id", tt.userID)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			handler.GetUserReviews(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
