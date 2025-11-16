package usecase

import "errors"

// Доменные ошибки Use Cases
var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")

	ErrPRAlreadyExists     = errors.New("pull request already exists")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRAlreadyMerged     = errors.New("pull request already merged")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoActiveCandidates  = errors.New("no active replacement candidate in team")
)
