package entity

import (
	"errors"
)

// Доменные ошибки
var (
	// ErrInvalidID возвращается при невалидном идентификаторе
	ErrInvalidID = errors.New("invalid id")

	// ErrInvalidUsername возвращается при невалидном имени пользователя
	ErrInvalidUsername = errors.New("invalid username")

	// ErrInvalidTeamName возвращается при невалидном названии команды
	ErrInvalidTeamName = errors.New("invalid team name")

	// ErrInvalidPRName возвращается при невалидном названии PR
	ErrInvalidPRName = errors.New("invalid pull request name")

	// ErrNoChange возвращается когда попытка изменить значение на то же самое
	ErrNoChange = errors.New("no change detected")

	// ErrPRMerged возвращается при попытке изменить merged PR
	ErrPRMerged = errors.New("pull request already merged")

	// ErrAuthorCannotReview возвращается при попытке назначить автора ревьювером
	ErrAuthorCannotReview = errors.New("author cannot review their own PR")

	// ErrReviewerAlreadyAssigned возвращается при попытке назначить уже назначенного ревьювера
	ErrReviewerAlreadyAssigned = errors.New("reviewer already assigned")

	// ErrReviewerNotAssigned возвращается при попытке удалить не назначенного ревьювера
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")

	// ErrTooManyReviewers возвращается при попытке назначить больше MaxReviewersCount ревьюверов
	ErrTooManyReviewers = errors.New("too many reviewers")
)
