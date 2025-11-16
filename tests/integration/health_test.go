package integration

import (
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
