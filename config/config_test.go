package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173, https://front.internal, ")
	t.Setenv("DB_CONNECTION_STRING", "postgres-dsn")
	t.Setenv("JOBS_ENABLED", "false")
	t.Setenv("DAILY_BILLING_JOB_TIME", "02:30")
	t.Setenv("PORT", "3000")
	t.Setenv("TRUSTED_PROXIES", "127.0.0.1, 10.0.0.0/8, ")

	cfg := Load()

	if cfg.AppEnv != "production" {
		t.Fatalf("Expected app env production, got %q", cfg.AppEnv)
	}
	if cfg.DBConnectionString != "postgres-dsn" {
		t.Fatalf("Expected DB connection string to be loaded")
	}
	wantOrigins := []string{"http://localhost:5173", "https://front.internal"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("Expected CORS allowed origins %v, got %v", wantOrigins, cfg.CORSAllowedOrigins)
	}
	if cfg.Port != "3000" {
		t.Fatalf("Expected port 3000, got %q", cfg.Port)
	}
	if cfg.JobsEnabled {
		t.Fatal("Expected jobs to be disabled")
	}
	if cfg.DailyBillingJobTime != "02:30" {
		t.Fatalf("Expected daily billing job time 02:30, got %q", cfg.DailyBillingJobTime)
	}

	wantProxies := []string{"127.0.0.1", "10.0.0.0/8"}
	if !reflect.DeepEqual(cfg.TrustedProxies, wantProxies) {
		t.Fatalf("Expected trusted proxies %v, got %v", wantProxies, cfg.TrustedProxies)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.AppEnv != EnvironmentDevelopment {
		t.Fatalf("Expected default app env %q, got %q", EnvironmentDevelopment, cfg.AppEnv)
	}

	wantProxies := []string{"127.0.0.1", "::1"}
	if !reflect.DeepEqual(cfg.TrustedProxies, wantProxies) {
		t.Fatalf("Expected default trusted proxies %v, got %v", wantProxies, cfg.TrustedProxies)
	}

	wantOrigins := []string{"http://localhost:5173", "http://localhost:3000"}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, wantOrigins) {
		t.Fatalf("Expected default CORS origins %v, got %v", wantOrigins, cfg.CORSAllowedOrigins)
	}
	if !cfg.JobsEnabled {
		t.Fatal("Expected jobs to be enabled by default")
	}
	if cfg.DailyBillingJobTime != "01:00" {
		t.Fatalf("Expected default daily billing job time 01:00, got %q", cfg.DailyBillingJobTime)
	}
}

func TestIsProduction(t *testing.T) {
	if !(Config{AppEnv: " Production "}).IsProduction() {
		t.Fatal("Expected production env to be detected case-insensitively")
	}

	if (Config{AppEnv: EnvironmentDevelopment}).IsProduction() {
		t.Fatal("Expected development env not to be production")
	}
}
