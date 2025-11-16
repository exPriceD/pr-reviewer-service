package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetStatistics(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/statistics")
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

	if _, ok := result["pr_stats"]; !ok {
		t.Fatal("Expected pr_stats in response")
	}
}

func TestGetStatisticsWithTeam(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/statistics?team_name=test-team-1")
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

	if _, ok := result["pr_stats"]; !ok {
		t.Fatal("Expected pr_stats in response")
	}
}
