package integration

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/exPriceD/pr-reviewer-service/internal/app"
	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/config"
)

var (
	testApp     *app.App
	testServer  *httptest.Server
	testBaseURL string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	testCfg := &config.Config{
		Server: config.ServerConfig{
			Port:            "0",
			Host:            "localhost",
			ReadTimeout:     30,
			WriteTimeout:    30,
			IdleTimeout:     60,
			MaxHeaderBytes:  1048576,
			MaxBodySize:     10485760,
			ShutdownTimeout: 10,
		},
		Database: config.DatabaseConfig{
			Host:            getEnv("TEST_DB_HOST", "localhost"),
			Port:            getEnvInt("TEST_DB_PORT", 5433),
			User:            getEnv("TEST_DB_USER", "test_user"),
			Password:        getEnv("TEST_DB_PASSWORD", "test_password"),
			DBName:          getEnv("TEST_DB_NAME", "pr_reviewer_test"),
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5,
			PingTimeout:     5,
		},
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}

	var err error
	testApp, err = buildTestApp(ctx, testCfg)
	if err != nil {
		fmt.Printf("Failed to build test app: %v\n", err)
		os.Exit(1)
	}

	testServer = httptest.NewServer(testApp.HTTPServer.Handler())
	testBaseURL = testServer.URL

	code := m.Run()

	testServer.Close()
	if testApp != nil {
		testApp.Shutdown()
	}

	os.Exit(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
