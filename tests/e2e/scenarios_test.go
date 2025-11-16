package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type TeamResponse struct {
	Team struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	} `json:"team"`
}

type PRResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
	} `json:"pr"`
}

func TestCompletePRWorkflow(t *testing.T) {
	teamName := fmt.Sprintf("team-e2e-%d", time.Now().Unix())
	userID1 := fmt.Sprintf("user-e2e-1-%d", time.Now().Unix())
	userID2 := fmt.Sprintf("user-e2e-2-%d", time.Now().Unix())
	userID3 := fmt.Sprintf("user-e2e-3-%d", time.Now().Unix())
	userID4 := fmt.Sprintf("user-e2e-4-%d", time.Now().Unix())
	prID := fmt.Sprintf("pr-e2e-%d", time.Now().Unix())

	teamReq := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": userID1, "username": "User 1", "is_active": true},
			{"user_id": userID2, "username": "User 2", "is_active": true},
			{"user_id": userID3, "username": "User 3", "is_active": true},
			{"user_id": userID4, "username": "User 4", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to create team: status %d, error: %v", resp.StatusCode, errResp)
	}

	var teamResp TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&teamResp); err != nil {
		t.Fatalf("Failed to decode team response: %v", err)
	}

	if teamResp.Team.TeamName != teamName {
		t.Errorf("Expected team_name %s, got %s", teamName, teamResp.Team.TeamName)
	}

	if len(teamResp.Team.Members) != 4 {
		t.Errorf("Expected 4 members, got %d", len(teamResp.Team.Members))
	}

	prReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Test PR",
		"author_id":         userID1,
	}

	body, _ = json.Marshal(prReq)
	resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to create PR: status %d, error: %v", resp.StatusCode, errResp)
	}

	var prResp PRResponse
	if err := json.NewDecoder(resp.Body).Decode(&prResp); err != nil {
		t.Fatalf("Failed to decode PR response: %v", err)
	}

	if prResp.PR.PullRequestID != prID {
		t.Errorf("Expected pull_request_id %s, got %s", prID, prResp.PR.PullRequestID)
	}

	if prResp.PR.Status != "OPEN" {
		t.Errorf("Expected status OPEN, got %s", prResp.PR.Status)
	}

	if len(prResp.PR.AssignedReviewers) == 0 {
		t.Error("Expected at least one reviewer assigned")
	}

	if len(prResp.PR.AssignedReviewers) > 2 {
		t.Errorf("Expected at most 2 reviewers, got %d", len(prResp.PR.AssignedReviewers))
	}

	if contains(prResp.PR.AssignedReviewers, userID1) {
		t.Error("Author should not be assigned as reviewer")
	}

	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     prResp.PR.AssignedReviewers[0],
	}

	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to reassign reviewer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to reassign reviewer: status %d, error: %v", resp.StatusCode, errResp)
	}

	var reassignResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&reassignResp); err != nil {
		t.Fatalf("Failed to decode reassign response: %v", err)
	}

	if _, ok := reassignResp["replaced_by"]; !ok {
		t.Error("Expected replaced_by in response")
	}

	mergeReq := map[string]interface{}{
		"pull_request_id": prID,
	}

	body, _ = json.Marshal(mergeReq)
	resp, err = http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to merge PR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to merge PR: status %d, error: %v", resp.StatusCode, errResp)
	}

	var mergeResp PRResponse
	if err := json.NewDecoder(resp.Body).Decode(&mergeResp); err != nil {
		t.Fatalf("Failed to decode merge response: %v", err)
	}

	if mergeResp.PR.Status != "MERGED" {
		t.Errorf("Expected status MERGED, got %s", mergeResp.PR.Status)
	}

	body, _ = json.Marshal(mergeReq)
	resp, err = http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to merge PR again: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 on idempotent merge, got %d", resp.StatusCode)
	}

	reassignAfterMerge := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     prResp.PR.AssignedReviewers[0],
	}

	body, _ = json.Marshal(reassignAfterMerge)
	resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to reassign reviewer after merge: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 when reassigning after merge, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error.Code != "PR_MERGED" {
		t.Errorf("Expected error code PR_MERGED, got %s", errResp.Error.Code)
	}
}

func TestDeactivateUserWorkflow(t *testing.T) {
	teamName := fmt.Sprintf("team-e2e-deactivate-%d", time.Now().Unix())
	userID1 := fmt.Sprintf("user-e2e-deactivate-1-%d", time.Now().Unix())
	userID2 := fmt.Sprintf("user-e2e-deactivate-2-%d", time.Now().Unix())
	prID := fmt.Sprintf("pr-e2e-deactivate-%d", time.Now().Unix())

	teamReq := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": userID1, "username": "User 1", "is_active": true},
			{"user_id": userID2, "username": "User 2", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	prReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Test PR Deactivate",
		"author_id":         userID1,
	}

	body, _ = json.Marshal(prReq)
	resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	resp.Body.Close()

	deactivateReq := map[string]interface{}{
		"user_id":   userID2,
		"is_active": false,
	}

	body, _ = json.Marshal(deactivateReq)
	resp, err = http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to deactivate user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to deactivate user: status %d, error: %v", resp.StatusCode, errResp)
	}

	var userResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}

	user, ok := userResp["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user in response")
	}

	if user["is_active"] != false {
		t.Errorf("Expected is_active false, got %v", user["is_active"])
	}

	prID2 := fmt.Sprintf("pr-e2e-deactivate-2-%d", time.Now().Unix())
	prReq2 := map[string]interface{}{
		"pull_request_id":   prID2,
		"pull_request_name": "E2E Test PR After Deactivate",
		"author_id":         userID1,
	}

	body, _ = json.Marshal(prReq2)
	resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create PR after deactivation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to create PR: status %d, error: %v", resp.StatusCode, errResp)
	}

	var prResp2 PRResponse
	if err := json.NewDecoder(resp.Body).Decode(&prResp2); err != nil {
		t.Fatalf("Failed to decode PR response: %v", err)
	}

	if contains(prResp2.PR.AssignedReviewers, userID2) {
		t.Error("Deactivated user should not be assigned as reviewer")
	}
}

func TestStatisticsWorkflow(t *testing.T) {
	teamName := fmt.Sprintf("team-e2e-stats-%d", time.Now().Unix())
	userID1 := fmt.Sprintf("user-e2e-stats-1-%d", time.Now().Unix())
	userID2 := fmt.Sprintf("user-e2e-stats-2-%d", time.Now().Unix())

	teamReq := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": userID1, "username": "User 1", "is_active": true},
			{"user_id": userID2, "username": "User 2", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	resp.Body.Close()

	for i := 0; i < 3; i++ {
		prID := fmt.Sprintf("pr-e2e-stats-%d-%d", time.Now().Unix(), i)
		prReq := map[string]interface{}{
			"pull_request_id":   prID,
			"pull_request_name": fmt.Sprintf("E2E Test PR %d", i),
			"author_id":         userID1,
		}

		body, _ = json.Marshal(prReq)
		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		resp.Body.Close()
	}

	resp, err = http.Get(baseURL + "/statistics?team_name=" + teamName)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var statsResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
		t.Fatalf("Failed to decode statistics response: %v", err)
	}

	if _, ok := statsResp["pr_stats"]; !ok {
		t.Fatal("Expected pr_stats in response")
	}

	prStats, ok := statsResp["pr_stats"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected pr_stats to be an object")
	}

	if total, ok := prStats["total"].(float64); ok && total < 3 {
		t.Errorf("Expected at least 3 total PRs, got %.0f", total)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
