package config

import (
	"os"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	yamlContent := `
server:
  host: "localhost"
  port: 8080
  maxCouponCodeCacheSize: 1000
  reqLimitPerIP: 5
  reqBurstPerIP: 10
  reqRateWindow: 1m
  gracefulTimeout: 30s
database:
  host: "dbhost"
  port: 5432
  name: "testdb"
  user: "testuser"
  password: "testpass"
  sslMode: "disable"
  debug: false
  type: "postgres"
  migrationsFolderPath: "db"
  maxOpenConnections: 1  
  maxIdleConnections: 1
  connMaxIdleTime: 1m
  maxConnMaxLifeTime: 1m
logging:
  level: debug
  jsonFormat: true
couponCode:
  unzipped: false
  filePaths:
    - "file1.gz"
    - "file2.gz"
    - "file3.gz"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig(%q) failed: %v", tmpFile.Name(), err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.GracefulTimeout.String() != "30s" {
		t.Errorf("expected server greacefulTimeout 30s, got %d", cfg.Server.Port)
	}
	if cfg.Database.Type != "postgres" {
		t.Errorf("expected database type postgres, got %s", cfg.Database.Type)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected logging debug true")
	}
	if !cfg.Logging.JsonFormat {
		t.Errorf("expected logging jsonFormat true, got false")
	}
	if cfg.CouponCode.Unzipped != false {
		t.Errorf("expected unzipped false, got true")
	}
	if len(cfg.CouponCode.FilePaths) != 3 {
		t.Errorf("expected 3 coupon code files, got %d", len(cfg.CouponCode.FilePaths))
	}
	if cfg.CouponCode.FilePaths[0] != "file1.gz" {
		t.Errorf("expected first coupon code file to be 'file1.gz', got %s", cfg.CouponCode.FilePaths[0])
	}
	if cfg.CouponCode.FilePaths[1] != "file2.gz" {
		t.Errorf("expected second coupon code file to be 'file2.gz', got %s", cfg.CouponCode.FilePaths[1])
	}
	if cfg.CouponCode.FilePaths[2] != "file3.gz" {
		t.Errorf("expected third coupon code file to be 'file3.gz', got %s", cfg.CouponCode.FilePaths[2])
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	invalidPath := "nonexistent.yaml"
	cfg, err := LoadConfig(invalidPath)
	if err == nil {
		t.Errorf("LoadConfig(%q) expected to fail, but succeeded with config: %+v", invalidPath, cfg)
	}
}
