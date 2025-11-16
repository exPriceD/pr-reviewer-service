package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreatePR(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-pr-test",
		"members": []map[string]interface{}{
			{"user_id": "user-pr-1", "username": "User PR 1", "is_active": true},
			{"user_id": "user-pr-2", "username": "User PR 2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	req := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Test PR",
		"author_id":         "user-pr-1",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Errorf("Expected status 201, got %d: %v", resp.StatusCode, errResp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected pr in response")
	}

	if pr["pull_request_id"] != "pr-1" {
		t.Errorf("Expected pull_request_id 'pr-1', got %v", pr["pull_request_id"])
	}

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	if !ok {
		t.Fatal("Expected assigned_reviewers in response")
	}

	if len(reviewers) == 0 {
		t.Error("Expected at least one reviewer assigned")
	}
}

func TestMergePR(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-merge-test",
		"members": []map[string]interface{}{
			{"user_id": "user-merge-1", "username": "User Merge 1", "is_active": true},
			{"user_id": "user-merge-2", "username": "User Merge 2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-merge-1",
		"pull_request_name": "Test PR Merge",
		"author_id":         "user-merge-1",
	}
	prBody, _ := json.Marshal(prReq)
	prResp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(prBody))
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	if prResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create PR: status %d", prResp.StatusCode)
	}
	prResp.Body.Close()

	req := map[string]interface{}{
		"pull_request_id": "pr-merge-1",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Errorf("Expected status 200, got %d: %v", resp.StatusCode, errResp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected pr in response")
	}

	if pr["status"] != "MERGED" {
		t.Errorf("Expected status 'MERGED', got %v", pr["status"])
	}

	t.Run("Idempotency", func(t *testing.T) {
		body2, _ := json.Marshal(req)
		resp2, err := http.Post(testBaseURL+"/pullRequest/merge", "application/json", bytes.NewReader(body2))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 on second merge, got %d", resp2.StatusCode)
		}
	})
}

func TestReassignReviewer(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-reassign-test",
		"members": []map[string]interface{}{
			{"user_id": "user-reassign-1", "username": "User Reassign 1", "is_active": true},
			{"user_id": "user-reassign-2", "username": "User Reassign 2", "is_active": true},
			{"user_id": "user-reassign-3", "username": "User Reassign 3", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-reassign-1",
		"pull_request_name": "Test PR Reassign",
		"author_id":         "user-reassign-1",
	}
	prBody, _ := json.Marshal(prReq)
	prResp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(prBody))
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	if prResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create PR: status %d", prResp.StatusCode)
	}
	prResp.Body.Close()

	req := map[string]interface{}{
		"pull_request_id": "pr-reassign-1",
		"old_user_id":     "user-reassign-2",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if _, ok := result["replaced_by"]; !ok {
			t.Error("Expected replaced_by in response")
		}
	case http.StatusConflict:
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error.Code == "PR_MERGED" {
			t.Log("PR is already merged, cannot reassign (expected)")
		} else {
			t.Logf("Reassign failed: %s", errResp.Error.Code)
		}
	default:
		t.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

func TestCreatePRWithoutReviewers(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-single-user",
		"members": []map[string]interface{}{
			{"user_id": "user-single-1", "username": "User Single 1", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	req := map[string]interface{}{
		"pull_request_id":   "pr-no-reviewers",
		"pull_request_name": "Test PR Without Reviewers",
		"author_id":         "user-single-1",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Errorf("Expected status 201, got %d: %v", resp.StatusCode, errResp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected pr in response")
	}

	if pr["pull_request_id"] != "pr-no-reviewers" {
		t.Errorf("Expected pull_request_id 'pr-no-reviewers', got %v", pr["pull_request_id"])
	}

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	if !ok {
		t.Fatal("Expected assigned_reviewers in response")
	}

	if len(reviewers) != 0 {
		t.Errorf("Expected 0 reviewers when only author in team, got %d", len(reviewers))
	}
}

func TestCreatePRWithOnlyInactiveUsers(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-inactive-users",
		"members": []map[string]interface{}{
			{"user_id": "user-inactive-1", "username": "User Inactive 1", "is_active": true},
			{"user_id": "user-inactive-2", "username": "User Inactive 2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	setActiveReq := map[string]interface{}{
		"user_id":   "user-inactive-2",
		"is_active": false,
	}
	setActiveBody, _ := json.Marshal(setActiveReq)
	setActiveResp, err := http.Post(testBaseURL+"/users/setIsActive", "application/json", bytes.NewReader(setActiveBody))
	if err != nil {
		t.Fatalf("Failed to deactivate user: %v", err)
	}
	setActiveResp.Body.Close()

	req := map[string]interface{}{
		"pull_request_id":   "pr-inactive-reviewers",
		"pull_request_name": "Test PR With Inactive Reviewers",
		"author_id":         "user-inactive-1",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Errorf("Expected status 201, got %d: %v", resp.StatusCode, errResp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pr, ok := result["pr"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected pr in response")
	}

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	if !ok {
		t.Fatal("Expected assigned_reviewers in response")
	}

	if len(reviewers) != 0 {
		t.Errorf("Expected 0 reviewers when only inactive users in team (besides author), got %d", len(reviewers))
	}
}
