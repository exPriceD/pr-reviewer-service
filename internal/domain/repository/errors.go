package repository

import "errors"

// Стандартные ошибки репозиториев
var (
	// ErrNotFound возвращается когда сущность не найдена
	ErrNotFound = errors.New("entity not found")
)
