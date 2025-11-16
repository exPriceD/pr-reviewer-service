package presenter

// ErrorResponse представляет ошибку в формате API
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail детали ошибки
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes согласно OpenAPI
const (
	ErrorCodeTeamExists     = "TEAM_EXISTS"
	ErrorCodePRExists       = "PR_EXISTS"
	ErrorCodePRMerged       = "PR_MERGED"
	ErrorCodeNotAssigned    = "NOT_ASSIGNED"
	ErrorCodeNoCandidate    = "NO_CANDIDATE"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeInvalidRequest = "INVALID_REQUEST"
	ErrorCodeInternalError  = "INTERNAL_ERROR"
)
