package config

import (
	"os"
	"strings"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

type Config struct {
	AppEnv              string
	CORSAllowedOrigins  []string
	DBConnectionString  string
	JobsEnabled         bool
	DailyBillingJobTime string
	Port                string
	TrustedProxies      []string
}

func Load() Config {
	return Config{
		AppEnv:              envOrDefault("APP_ENV", EnvironmentDevelopment),
		CORSAllowedOrigins:  splitList(envOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		DBConnectionString:  os.Getenv("DB_CONNECTION_STRING"),
		JobsEnabled:         boolEnvOrDefault("JOBS_ENABLED", true),
		DailyBillingJobTime: envOrDefault("DAILY_BILLING_JOB_TIME", "01:00"),
		Port:                os.Getenv("PORT"),
		TrustedProxies:      splitList(envOrDefault("TRUSTED_PROXIES", "127.0.0.1,::1")),
	}
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), EnvironmentProduction)
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func boolEnvOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return strings.EqualFold(value, "true") || value == "1"
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
