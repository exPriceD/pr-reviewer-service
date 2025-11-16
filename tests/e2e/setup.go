package e2e

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL = "http://localhost:8081"
	maxRetries     = 60
	retryDelay     = 2 * time.Second
)

func getBaseURL() string {
	if url := os.Getenv("E2E_BASE_URL"); url != "" {
		return url
	}
	return defaultBaseURL
}

func waitForServer(baseURL string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			//nolint:gosec,errcheck // игнорируем в тестах
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			//nolint:gosec,errcheck // игнорируем в тестах
			resp.Body.Close()
		}
		//nolint:forbidigo // игнорируем в тестах
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("server did not become healthy after %d retries", maxRetries)
}
