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

type mockPullRequestUseCase struct {
	createPR         func(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error)
	mergePR          func(ctx context.Context, prID string) (*dto.PullRequestDTO, error)
	reassignReviewer func(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error)
}

func (m *mockPullRequestUseCase) CreatePR(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error) {
	return m.createPR(ctx, req)
}

func (m *mockPullRequestUseCase) MergePR(ctx context.Context, prID string) (*dto.PullRequestDTO, error) {
	return m.mergePR(ctx, prID)
}

func (m *mockPullRequestUseCase) ReassignReviewer(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error) {
	return m.reassignReviewer(ctx, req)
}

func TestPullRequestHandler_CreatePR(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.CreatePRRequest
		setupMock  func() *mockPullRequestUseCase
		wantStatus int
	}{
		{
			name: "success",
			body: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        "user-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					createPR: func(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error) {
						return &dto.PullRequestDTO{
							PullRequestID:     req.PullRequestID,
							PullRequestName:   req.PullRequestName,
							AuthorID:          req.AuthorID,
							Status:            string(entity.PRStatusOpen),
							AssignedReviewers: []string{},
						}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid body",
			body: dto.CreatePRRequest{},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: dto.CreatePRRequest{
				PullRequestID:   "",
				PullRequestName: "Test PR",
				AuthorID:        "user-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error",
			body: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR",
				AuthorID:        "user-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					createPR: func(ctx context.Context, req dto.CreatePRRequest) (*dto.PullRequestDTO, error) {
						return nil, usecase.ErrPRAlreadyExists
					},
				}
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewPullRequestHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.CreatePR(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestPullRequestHandler_MergePR(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.MergePRRequest
		setupMock  func() *mockPullRequestUseCase
		wantStatus int
	}{
		{
			name: "success",
			body: dto.MergePRRequest{
				PullRequestID: "pr-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					mergePR: func(ctx context.Context, prID string) (*dto.PullRequestDTO, error) {
						return &dto.PullRequestDTO{
							PullRequestID: prID,
							Status:        string(entity.PRStatusMerged),
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid body",
			body: dto.MergePRRequest{},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error",
			body: dto.MergePRRequest{
				PullRequestID: "pr-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					mergePR: func(ctx context.Context, prID string) (*dto.PullRequestDTO, error) {
						return nil, usecase.ErrPRNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewPullRequestHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.MergePR(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestPullRequestHandler_ReassignReviewer(t *testing.T) {
	tests := []struct {
		name       string
		body       dto.ReassignReviewerRequest
		setupMock  func() *mockPullRequestUseCase
		wantStatus int
	}{
		{
			name: "success",
			body: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					reassignReviewer: func(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error) {
						return &dto.PullRequestDTO{
							PullRequestID: req.PullRequestID,
							Status:        string(entity.PRStatusOpen),
						}, "reviewer-2", nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid body",
			body: dto.ReassignReviewerRequest{},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "use case error",
			body: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMock: func() *mockPullRequestUseCase {
				return &mockPullRequestUseCase{
					reassignReviewer: func(ctx context.Context, req dto.ReassignReviewerRequest) (*dto.PullRequestDTO, string, error) {
						return nil, "", usecase.ErrPRNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := tt.setupMock()
			handler := NewPullRequestHandler(mockUseCase)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.ReassignReviewer(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
