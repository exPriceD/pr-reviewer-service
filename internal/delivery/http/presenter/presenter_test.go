package presenter

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	RespondJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("expected message 'test', got %s", result["message"])
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	RespondError(w, http.StatusBadRequest, ErrorCodeInvalidRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var result ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Error.Code != ErrorCodeInvalidRequest {
		t.Errorf("expected error code %s, got %s", ErrorCodeInvalidRequest, result.Error.Code)
	}

	if result.Error.Message != "test error" {
		t.Errorf("expected error message 'test error', got %s", result.Error.Message)
	}
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"data": "test"}

	RespondSuccess(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["data"] != "test" {
		t.Errorf("expected data 'test', got %s", result["data"])
	}
}

func TestMapUseCaseError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantCode       string
		wantMessage    string
	}{
		{
			name:           "team already exists",
			err:            usecase.ErrTeamAlreadyExists,
			wantStatusCode: http.StatusBadRequest,
			wantCode:       ErrorCodeTeamExists,
			wantMessage:    "team_name already exists",
		},
		{
			name:           "team not found",
			err:            usecase.ErrTeamNotFound,
			wantStatusCode: http.StatusNotFound,
			wantCode:       ErrorCodeNotFound,
			wantMessage:    "team not found",
		},
		{
			name:           "user not found",
			err:            usecase.ErrUserNotFound,
			wantStatusCode: http.StatusNotFound,
			wantCode:       ErrorCodeNotFound,
			wantMessage:    "user not found",
		},
		{
			name:           "PR already exists",
			err:            usecase.ErrPRAlreadyExists,
			wantStatusCode: http.StatusConflict,
			wantCode:       ErrorCodePRExists,
			wantMessage:    "PR id already exists",
		},
		{
			name:           "PR not found",
			err:            usecase.ErrPRNotFound,
			wantStatusCode: http.StatusNotFound,
			wantCode:       ErrorCodeNotFound,
			wantMessage:    "pull request not found",
		},
		{
			name:           "PR already merged",
			err:            usecase.ErrPRAlreadyMerged,
			wantStatusCode: http.StatusConflict,
			wantCode:       ErrorCodePRMerged,
			wantMessage:    "cannot reassign on merged PR",
		},
		{
			name:           "reviewer not assigned",
			err:            usecase.ErrReviewerNotAssigned,
			wantStatusCode: http.StatusConflict,
			wantCode:       ErrorCodeNotAssigned,
			wantMessage:    "reviewer is not assigned to this PR",
		},
		{
			name:           "no active candidates",
			err:            usecase.ErrNoActiveCandidates,
			wantStatusCode: http.StatusConflict,
			wantCode:       ErrorCodeNoCandidate,
			wantMessage:    "no active replacement candidate in team",
		},
		{
			name:           "nil error",
			err:            nil,
			wantStatusCode: http.StatusInternalServerError,
			wantCode:       ErrorCodeInternalError,
			wantMessage:    "internal server error",
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown error"),
			wantStatusCode: http.StatusInternalServerError,
			wantCode:       ErrorCodeInternalError,
			wantMessage:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, code, message := MapUseCaseError(tt.err)

			if statusCode != tt.wantStatusCode {
				t.Errorf("expected status code %d, got %d", tt.wantStatusCode, statusCode)
			}
			if code != tt.wantCode {
				t.Errorf("expected error code %s, got %s", tt.wantCode, code)
			}
			if message != tt.wantMessage {
				t.Errorf("expected message %s, got %s", tt.wantMessage, message)
			}
		})
	}
}

func TestRespondPullRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		pr := &dto.PullRequestDTO{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            "OPEN",
			AssignedReviewers: []string{"reviewer-1"},
		}

		RespondPullRequest(w, http.StatusCreated, pr)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var result map[string]*dto.PullRequestDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["pr"] == nil {
			t.Fatal("expected pr in response, got nil")
		}

		if result["pr"].PullRequestID != "pr-1" {
			t.Errorf("expected PR ID 'pr-1', got %s", result["pr"].PullRequestID)
		}
	})

	t.Run("nil pr", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondPullRequest(w, http.StatusCreated, nil)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "pull request data is nil" {
			t.Errorf("expected error message 'pull request data is nil', got %s", result.Error.Message)
		}
	})
}

func TestRespondPullRequestReassign(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		pr := &dto.PullRequestDTO{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            "OPEN",
			AssignedReviewers: []string{"reviewer-2"},
		}
		replacedBy := "reviewer-2"

		RespondPullRequestReassign(w, http.StatusOK, pr, replacedBy)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["pr"] == nil {
			t.Fatal("expected pr in response, got nil")
		}

		if result["replaced_by"] != "reviewer-2" {
			t.Errorf("expected replaced_by 'reviewer-2', got %v", result["replaced_by"])
		}
	})

	t.Run("nil pr", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondPullRequestReassign(w, http.StatusOK, nil, "reviewer-2")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "pull request data is nil" {
			t.Errorf("expected error message 'pull request data is nil', got %s", result.Error.Message)
		}
	})

	t.Run("empty replaced_by", func(t *testing.T) {
		w := httptest.NewRecorder()
		pr := &dto.PullRequestDTO{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            "OPEN",
			AssignedReviewers: []string{"reviewer-1"},
		}

		RespondPullRequestReassign(w, http.StatusOK, pr, "")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "replaced_by is empty" {
			t.Errorf("expected error message 'replaced_by is empty', got %s", result.Error.Message)
		}
	})
}

func TestRespondTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		team := &dto.TeamDTO{
			TeamName: "team-1",
			Members: []dto.TeamMemberDTO{
				{UserID: "user-1", Username: "User 1", IsActive: true},
				{UserID: "user-2", Username: "User 2", IsActive: true},
			},
		}

		RespondTeam(w, http.StatusOK, team)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]*dto.TeamDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["team"] == nil {
			t.Fatal("expected team in response, got nil")
		}

		if result["team"].TeamName != "team-1" {
			t.Errorf("expected team name 'team-1', got %s", result["team"].TeamName)
		}

		if len(result["team"].Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(result["team"].Members))
		}
	})

	t.Run("nil team", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTeam(w, http.StatusOK, nil)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "team data is nil" {
			t.Errorf("expected error message 'team data is nil', got %s", result.Error.Message)
		}
	})
}

func TestRespondUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		user := &dto.UserDTO{
			UserID:   "user-1",
			Username: "User 1",
			TeamName: "team-1",
			IsActive: true,
		}

		RespondUser(w, http.StatusOK, user)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]*dto.UserDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["user"] == nil {
			t.Fatal("expected user in response, got nil")
		}

		if result["user"].UserID != "user-1" {
			t.Errorf("expected user ID 'user-1', got %s", result["user"].UserID)
		}

		if result["user"].Username != "User 1" {
			t.Errorf("expected username 'User 1', got %s", result["user"].Username)
		}
	})

	t.Run("nil user", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondUser(w, http.StatusOK, nil)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "user data is nil" {
			t.Errorf("expected error message 'user data is nil', got %s", result.Error.Message)
		}
	})
}

func TestRespondUserReviews(t *testing.T) {
	t.Run("success - with reviews", func(t *testing.T) {
		w := httptest.NewRecorder()
		prs := []dto.PullRequestShortDTO{
			{
				PullRequestID:   "pr-1",
				PullRequestName: "Test PR 1",
				AuthorID:        "author-1",
				Status:          "OPEN",
			},
			{
				PullRequestID:   "pr-2",
				PullRequestName: "Test PR 2",
				AuthorID:        "author-2",
				Status:          "MERGED",
			},
		}
		userID := "user-1"

		RespondUserReviews(w, http.StatusOK, userID, prs)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["user_id"] != "user-1" {
			t.Errorf("expected user_id 'user-1', got %v", result["user_id"])
		}

		pullRequests, ok := result["pull_requests"].([]interface{})
		if !ok {
			t.Fatalf("expected pull_requests to be array, got %T", result["pull_requests"])
		}

		if len(pullRequests) != 2 {
			t.Errorf("expected 2 pull requests, got %d", len(pullRequests))
		}
	})

	t.Run("success - empty reviews", func(t *testing.T) {
		w := httptest.NewRecorder()
		var prs []dto.PullRequestShortDTO
		userID := "user-1"

		RespondUserReviews(w, http.StatusOK, userID, prs)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["user_id"] != "user-1" {
			t.Errorf("expected user_id 'user-1', got %v", result["user_id"])
		}

		pullRequests, ok := result["pull_requests"].([]interface{})
		if !ok {
			t.Fatalf("expected pull_requests to be array, got %T", result["pull_requests"])
		}

		if len(pullRequests) != 0 {
			t.Errorf("expected 0 pull requests, got %d", len(pullRequests))
		}
	})
}

func TestRespondStatistics(t *testing.T) {
	t.Run("success - with user stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		stats := &dto.StatisticsDTO{
			PRStats: dto.PRStatsDTO{
				Total:  10,
				Open:   5,
				Merged: 5,
			},
			UserStats: []dto.UserStatsDTO{
				{UserID: "user-1", TotalReviews: 3, ActiveReviews: 2},
				{UserID: "user-2", TotalReviews: 2, ActiveReviews: 1},
			},
		}

		RespondStatistics(w, http.StatusOK, stats)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result dto.StatisticsDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.PRStats.Total != 10 {
			t.Errorf("expected total PRs 10, got %d", result.PRStats.Total)
		}

		if len(result.UserStats) != 2 {
			t.Errorf("expected 2 user stats, got %d", len(result.UserStats))
		}
	})

	t.Run("success - without user stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		stats := &dto.StatisticsDTO{
			PRStats: dto.PRStatsDTO{
				Total:  10,
				Open:   5,
				Merged: 5,
			},
			UserStats: []dto.UserStatsDTO{},
		}

		RespondStatistics(w, http.StatusOK, stats)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result dto.StatisticsDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.PRStats.Total != 10 {
			t.Errorf("expected total PRs 10, got %d", result.PRStats.Total)
		}

		if len(result.UserStats) != 0 {
			t.Errorf("expected 0 user stats, got %d", len(result.UserStats))
		}
	})

	t.Run("nil stats", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondStatistics(w, http.StatusOK, nil)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var result ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Error.Code != ErrorCodeInternalError {
			t.Errorf("expected error code %s, got %s", ErrorCodeInternalError, result.Error.Code)
		}

		if result.Error.Message != "statistics data is nil" {
			t.Errorf("expected error message 'statistics data is nil', got %s", result.Error.Message)
		}
	})
}
