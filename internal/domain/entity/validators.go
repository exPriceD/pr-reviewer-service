package entity

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// idPattern паттерн для ID: буквы, цифры, дефис, подчёркивание
	idPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// usernamePattern паттерн для имен пользователей
	usernamePattern = regexp.MustCompile(`^[\p{L}\p{N}\s_-]+$`)
)

const (
	minIDLength = 1
	maxIDLength = 255

	minUsernameLength = 1
	maxUsernameLength = 100

	minTeamNameLength = 1
	maxTeamNameLength = 100

	minPRNameLength = 1
	maxPRNameLength = 200
)

// validateAndNormalizeID проверяет и нормализует идентификатор
func validateAndNormalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)

	if len(id) < minIDLength || len(id) > maxIDLength {
		return "", fmt.Errorf("id length must be between %d and %d characters", minIDLength, maxIDLength)
	}

	if !idPattern.MatchString(id) {
		return "", fmt.Errorf("id must contain only letters, numbers, hyphens, and underscores")
	}

	return id, nil
}

// validateAndNormalizeUsername проверяет и нормализует имя пользователя
func validateAndNormalizeUsername(username string) (string, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return "", fmt.Errorf("username length must be between %d and %d characters", minUsernameLength, maxUsernameLength)
	}

	if !usernamePattern.MatchString(username) {
		return "", fmt.Errorf("username contains invalid characters")
	}

	return username, nil
}

// validateAndNormalizeTeamName проверяет и нормализует название команды
func validateAndNormalizeTeamName(teamName string) (string, error) {
	teamName = strings.TrimSpace(teamName)

	if teamName == "" {
		return "", fmt.Errorf("team_name cannot be empty")
	}

	if len(teamName) < minTeamNameLength || len(teamName) > maxTeamNameLength {
		return "", fmt.Errorf("team_name length must be between %d and %d characters", minTeamNameLength, maxTeamNameLength)
	}

	if !idPattern.MatchString(teamName) {
		return "", fmt.Errorf("team_name must contain only letters, numbers, hyphens, and underscores")
	}

	return teamName, nil
}

// validateAndNormalizePRName проверяет и нормализует название Pull Request
func validateAndNormalizePRName(prName string) (string, error) {
	prName = strings.TrimSpace(prName)

	if prName == "" {
		return "", fmt.Errorf("pull_request_name cannot be empty")
	}

	if len(prName) < minPRNameLength || len(prName) > maxPRNameLength {
		return "", fmt.Errorf("pull_request_name length must be between %d and %d characters", minPRNameLength, maxPRNameLength)
	}

	if !usernamePattern.MatchString(prName) {
		return "", fmt.Errorf("pull_request_name contains invalid characters")
	}

	return prName, nil
}
