package e2e

import (
	"fmt"
	"os"
	"testing"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = getBaseURL()

	if err := waitForServer(baseURL); err != nil {
		fmt.Printf("Server is not available: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}
