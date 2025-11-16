package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateTeam(t *testing.T) {
	req := map[string]interface{}{
		"team_name": "test-team-1",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "User 1", "is_active": true},
			{"user_id": "user-2", "username": "User 2", "is_active": true},
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(body))
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

	team, ok := result["team"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected team in response")
	}

	if team["team_name"] != "test-team-1" {
		t.Errorf("Expected team_name 'test-team-1', got %v", team["team_name"])
	}
}

func TestGetTeam(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/team/get?team_name=test-team-1")
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

	if result["team_name"] != "test-team-1" {
		t.Errorf("Expected team_name 'test-team-1', got %v", result["team_name"])
	}
}

func TestCreateTeamDuplicate(t *testing.T) {
	req := map[string]interface{}{
		"team_name": "test-team-1",
		"members": []map[string]interface{}{
			{"user_id": "user-3", "username": "User 3", "is_active": true},
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(testBaseURL+"/team/add", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for duplicate team, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error.Code != "TEAM_EXISTS" {
		t.Errorf("Expected error code TEAM_EXISTS, got %s", errResp.Error.Code)
	}
}

func TestGetTeamNotFound(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/team/get?team_name=nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeactivateTeamMembers(t *testing.T) {
	req := map[string]interface{}{
		"team_name": "test-team-1",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", testBaseURL+"/team/deactivateMembers", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
	}
}
