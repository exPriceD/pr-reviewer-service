package validator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/usecase/dto"
)

func TestValidateCreateTeamRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      dto.CreateTeamRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "User 1", IsActive: true},
				},
			},
			wantErrs: 0,
		},
		{
			name: "empty team name",
			req: dto.CreateTeamRequest{
				TeamName: "",
				Members:  []dto.TeamMemberRequest{},
			},
			wantErrs: 2,
		},
		{
			name: "empty members",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members:  []dto.TeamMemberRequest{},
			},
			wantErrs: 1,
		},
		{
			name: "member without user_id",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "", Username: "User 1", IsActive: true},
				},
			},
			wantErrs: 1,
		},
		{
			name: "member without username",
			req: dto.CreateTeamRequest{
				TeamName: "team-1",
				Members: []dto.TeamMemberRequest{
					{UserID: "user-1", Username: "", IsActive: true},
				},
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateTeamRequest(tt.req)
			if len(errs) != tt.wantErrs {
				t.Errorf("expected %d errors, got %d", tt.wantErrs, len(errs))
			}
		})
	}
}

func TestValidateCreatePRRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      dto.CreatePRRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "PR 1",
				AuthorID:        "user-1",
			},
			wantErrs: 0,
		},
		{
			name: "empty pull_request_id",
			req: dto.CreatePRRequest{
				PullRequestID:   "",
				PullRequestName: "PR 1",
				AuthorID:        "user-1",
			},
			wantErrs: 1,
		},
		{
			name: "empty pull_request_name",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "",
				AuthorID:        "user-1",
			},
			wantErrs: 1,
		},
		{
			name: "empty author_id",
			req: dto.CreatePRRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "PR 1",
				AuthorID:        "",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreatePRRequest(tt.req)
			if len(errs) != tt.wantErrs {
				t.Errorf("expected %d errors, got %d", tt.wantErrs, len(errs))
			}
		})
	}
}

func TestValidateSetUserActiveRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      dto.SetUserActiveRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: dto.SetUserActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			wantErrs: 0,
		},
		{
			name: "empty user_id",
			req: dto.SetUserActiveRequest{
				UserID:   "",
				IsActive: true,
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateSetUserActiveRequest(tt.req)
			if len(errs) != tt.wantErrs {
				t.Errorf("expected %d errors, got %d", tt.wantErrs, len(errs))
			}
		})
	}
}

func TestValidateReassignReviewerRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      dto.ReassignReviewerRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			wantErrs: 0,
		},
		{
			name: "empty pull_request_id",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "",
				OldUserID:     "reviewer-1",
			},
			wantErrs: 1,
		},
		{
			name: "empty old_user_id",
			req: dto.ReassignReviewerRequest{
				PullRequestID: "pr-1",
				OldUserID:     "",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateReassignReviewerRequest(tt.req)
			if len(errs) != tt.wantErrs {
				t.Errorf("expected %d errors, got %d", tt.wantErrs, len(errs))
			}
		})
	}
}

func TestRespondValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		errors     []ValidationError
		wantStatus int
	}{
		{
			name:       "single error",
			errors:     []ValidationError{{Field: "field1", Message: "error1"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "multiple errors",
			errors: []ValidationError{
				{Field: "field1", Message: "error1"},
				{Field: "field2", Message: "error2"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty errors",
			errors:     []ValidationError{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondValidationErrors(w, tt.errors)

			if w.Code != tt.wantStatus && len(tt.errors) > 0 {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
