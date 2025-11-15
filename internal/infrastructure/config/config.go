package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config конфигурация приложения
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port            string `yaml:"port"`
	Host            string `yaml:"host"`
	ReadTimeout     int    `yaml:"read_timeout"`     // в секундах
	WriteTimeout    int    `yaml:"write_timeout"`    // в секундах
	IdleTimeout     int    `yaml:"idle_timeout"`     // в секундах
	MaxHeaderBytes  int    `yaml:"max_header_bytes"` // в байтах
	MaxBodySize     int    `yaml:"max_body_size"`    // в байтах
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // в секундах
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	PingTimeout     int    `yaml:"ping_timeout"` // в секундах
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load загружает конфигурацию из файла и переопределяет значения из переменных окружения
// CONFIG_FILE определяет имя конфиг-файла (например, development для configs/development.yaml)
// По умолчанию используется development
func Load() (*Config, error) {
	configFile := getEnv("CONFIG_FILE", "development")

	configPath := filepath.Join("configs", configFile+".yaml")

	//nolint:gosec // configPath создается из CONFIG_FILE env var, а не из пользовательского ввода
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	applyServerOverrides(cfg)
	applyDatabaseOverrides(cfg)
	applyLoggerOverrides(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func applyServerOverrides(cfg *Config) {
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if readTimeout := os.Getenv("SERVER_READ_TIMEOUT"); readTimeout != "" {
		if t, err := strconv.Atoi(readTimeout); err == nil {
			cfg.Server.ReadTimeout = t
		}
	}
	if writeTimeout := os.Getenv("SERVER_WRITE_TIMEOUT"); writeTimeout != "" {
		if t, err := strconv.Atoi(writeTimeout); err == nil {
			cfg.Server.WriteTimeout = t
		}
	}
	if idleTimeout := os.Getenv("SERVER_IDLE_TIMEOUT"); idleTimeout != "" {
		if t, err := strconv.Atoi(idleTimeout); err == nil {
			cfg.Server.IdleTimeout = t
		}
	}
	if maxHeaderBytes := os.Getenv("SERVER_MAX_HEADER_BYTES"); maxHeaderBytes != "" {
		if m, err := strconv.Atoi(maxHeaderBytes); err == nil {
			cfg.Server.MaxHeaderBytes = m
		}
	}
	if maxBodySize := os.Getenv("SERVER_MAX_BODY_SIZE"); maxBodySize != "" {
		if m, err := strconv.Atoi(maxBodySize); err == nil {
			cfg.Server.MaxBodySize = m
		}
	}
	if shutdownTimeout := os.Getenv("SERVER_SHUTDOWN_TIMEOUT"); shutdownTimeout != "" {
		if t, err := strconv.Atoi(shutdownTimeout); err == nil {
			cfg.Server.ShutdownTimeout = t
		}
	}
}

func applyDatabaseOverrides(cfg *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}
	if maxOpenConns := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpenConns != "" {
		if m, err := strconv.Atoi(maxOpenConns); err == nil {
			cfg.Database.MaxOpenConns = m
		}
	}
	if maxIdleConns := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdleConns != "" {
		if m, err := strconv.Atoi(maxIdleConns); err == nil {
			cfg.Database.MaxIdleConns = m
		}
	}
	if connMaxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME"); connMaxLifetime != "" {
		if m, err := strconv.Atoi(connMaxLifetime); err == nil {
			cfg.Database.ConnMaxLifetime = m
		}
	}
	if pingTimeout := os.Getenv("DB_PING_TIMEOUT"); pingTimeout != "" {
		if t, err := strconv.Atoi(pingTimeout); err == nil {
			cfg.Database.PingTimeout = t
		}
	}
}

func applyLoggerOverrides(cfg *Config) {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Logger.Format = format
	}
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if err := c.validateServer(); err != nil {
		return err
	}
	if err := c.validateDatabase(); err != nil {
		return err
	}
	return c.validateLogger()
}

func (c *Config) validateServer() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 15
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 15
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20
	}
	if c.Server.MaxBodySize == 0 {
		c.Server.MaxBodySize = 10 * 1024 * 1024
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 10
	}

	if c.Server.ReadTimeout < 1 {
		return fmt.Errorf("server read_timeout must be at least 1 second")
	}
	if c.Server.WriteTimeout < 1 {
		return fmt.Errorf("server write_timeout must be at least 1 second")
	}
	if c.Server.IdleTimeout < 1 {
		return fmt.Errorf("server idle_timeout must be at least 1 second")
	}
	if c.Server.MaxHeaderBytes < 1024 {
		return fmt.Errorf("server max_header_bytes must be at least 1024 bytes")
	}
	if c.Server.MaxBodySize < 1024 {
		return fmt.Errorf("server max_body_size must be at least 1024 bytes")
	}
	if c.Server.ShutdownTimeout < 1 {
		return fmt.Errorf("server shutdown_timeout must be at least 1 second")
	}

	return nil
}

func (c *Config) validateDatabase() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 5
	}
	if c.Database.PingTimeout == 0 {
		c.Database.PingTimeout = 5
	}

	if c.Database.MaxOpenConns < 1 {
		return fmt.Errorf("database max_open_conns must be at least 1")
	}
	if c.Database.MaxIdleConns < 1 {
		return fmt.Errorf("database max_idle_conns must be at least 1")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database max_idle_conns cannot be greater than max_open_conns")
	}
	if c.Database.ConnMaxLifetime < 1 {
		return fmt.Errorf("database conn_max_lifetime must be at least 1 minute")
	}
	if c.Database.PingTimeout < 1 {
		return fmt.Errorf("database ping_timeout must be at least 1 second")
	}

	return nil
}

func (c *Config) validateLogger() error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logger.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logger.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", c.Logger.Format)
	}

	return nil
}

// getEnv получает значение из environment или возвращает default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
