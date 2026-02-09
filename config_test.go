package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("screenshot dir is not empty", func(t *testing.T) {
		if cfg.ScreenshotDir == "" {
			t.Error("ScreenshotDir should not be empty")
		}
	})

	t.Run("screenshot dir exists", func(t *testing.T) {
		info, err := os.Stat(cfg.ScreenshotDir)
		if err != nil {
			t.Errorf("ScreenshotDir %q should exist: %v", cfg.ScreenshotDir, err)
		} else if !info.IsDir() {
			t.Errorf("ScreenshotDir %q should be a directory", cfg.ScreenshotDir)
		}
	})

	t.Run("default provider is claude", func(t *testing.T) {
		if cfg.Provider != ProviderClaude {
			t.Errorf("default Provider = %q, want %q", cfg.Provider, ProviderClaude)
		}
	})

	t.Run("default enabled", func(t *testing.T) {
		if !cfg.Enabled {
			t.Error("default Enabled should be true")
		}
	})

	t.Run("max filename length positive", func(t *testing.T) {
		if cfg.MaxFileNameLen <= 0 {
			t.Errorf("MaxFileNameLen = %d, should be positive", cfg.MaxFileNameLen)
		}
	})
}

func TestSaveAndLoadConfig(t *testing.T) {
	// 임시 디렉토리에서 테스트
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	original := Config{
		ScreenshotDir:  "/tmp/screenshots",
		OCRHelperPath:  "/usr/local/bin/ocr-helper",
		Provider:       ProviderCodex,
		ClaudePath:     "/usr/local/bin/claude",
		CodexPath:      "/usr/local/bin/codex",
		MaxFileNameLen: 50,
		Enabled:        false,
	}

	// 저장
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// 읽기
	readData, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// 검증
	if loaded.Provider != ProviderCodex {
		t.Errorf("Provider = %q, want %q", loaded.Provider, ProviderCodex)
	}
	if loaded.MaxFileNameLen != 50 {
		t.Errorf("MaxFileNameLen = %d, want 50", loaded.MaxFileNameLen)
	}
	if loaded.Enabled != false {
		t.Error("Enabled should be false")
	}
	if loaded.CodexPath != "/usr/local/bin/codex" {
		t.Errorf("CodexPath = %q, want /usr/local/bin/codex", loaded.CodexPath)
	}
}

func TestConfigJSON_ProviderValues(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		wantJSON string
	}{
		{"claude", ProviderClaude, `"claude"`},
		{"codex", ProviderCodex, `"codex"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.provider)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("json = %s, want %s", data, tt.wantJSON)
			}

			var got Provider
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if got != tt.provider {
				t.Errorf("roundtrip: got %q, want %q", got, tt.provider)
			}
		})
	}
}

func TestConfigJSON_InvalidJSON(t *testing.T) {
	var cfg Config
	err := json.Unmarshal([]byte(`{invalid`), &cfg)
	if err == nil {
		t.Error("should return error for invalid JSON")
	}
}

func TestConfigJSON_PartialOverride(t *testing.T) {
	// 일부 필드만 있는 JSON도 파싱 가능해야 함
	data := []byte(`{"provider": "codex", "enabled": false}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if cfg.Provider != ProviderCodex {
		t.Errorf("Provider = %q, want codex", cfg.Provider)
	}
	if cfg.Enabled != false {
		t.Error("Enabled should be false")
	}
	// 미지정 필드는 zero value
	if cfg.MaxFileNameLen != 0 {
		t.Errorf("MaxFileNameLen = %d, want 0 (zero value)", cfg.MaxFileNameLen)
	}
}
