package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSetUserActive(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-user-active-test",
		"members": []map[string]interface{}{
			{"user_id": "user-active-1", "username": "User Active 1", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	req := map[string]interface{}{
		"user_id":   "user-active-1",
		"is_active": false,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
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

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user in response")
	}

	if user["is_active"] != false {
		t.Errorf("Expected is_active false, got %v", user["is_active"])
	}
}

func TestGetUserReviews(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "team-reviews-test",
		"members": []map[string]interface{}{
			{"user_id": "user-reviews-1", "username": "User Reviews 1", "is_active": true},
			{"user_id": "user-reviews-2", "username": "User Reviews 2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	teamResp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	teamResp.Body.Close()

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-reviews-1",
		"pull_request_name": "Test PR Reviews",
		"author_id":         "user-reviews-1",
	}
	prBody, _ := json.Marshal(prReq)
	prResp, err := http.Post(testBaseURL+"/pullRequest/create", "application/json", bytes.NewReader(prBody))
	if err != nil {
		t.Fatalf("Failed to create PR: %v", err)
	}
	prResp.Body.Close()

	resp, err := http.Get(testBaseURL + "/users/getReview?user_id=user-reviews-2")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["user_id"] != "user-reviews-2" {
		t.Errorf("Expected user_id 'user-reviews-2', got %v", result["user_id"])
	}
}
