package dto

// CreatePRRequest входные данные для создания PR
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// ReassignReviewerRequest входные данные для переназначения ревьювера
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// MergePRRequest входные данные для мерджа PR
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}
