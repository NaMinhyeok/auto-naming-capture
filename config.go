package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	ScreenshotDir  string `json:"screenshot_dir"`
	OCRHelperPath  string `json:"ocr_helper_path"`
	ClaudePath     string `json:"claude_path"`
	MaxFileNameLen int    `json:"max_filename_length"`
	Enabled        bool   `json:"enabled"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "auto-naming-capture")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func detectScreenshotDir() string {
	out, err := exec.Command("defaults", "read", "com.apple.screencapture", "location").Output()
	if err == nil {
		dir := strings.TrimSpace(string(out))
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Desktop")
}

func detectExecutablePath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return name
	}
	return path
}

func DefaultConfig() Config {
	execPath, _ := os.Executable()
	ocrHelper := filepath.Join(filepath.Dir(execPath), "ocr-helper")
	if _, err := os.Stat(ocrHelper); err != nil {
		ocrHelper = filepath.Join(".", "ocr-helper", "ocr-helper")
	}

	return Config{
		ScreenshotDir:  detectScreenshotDir(),
		OCRHelperPath:  ocrHelper,
		ClaudePath:     detectExecutablePath("claude"),
		MaxFileNameLen: 80,
		Enabled:        true,
	}
}

func LoadConfig() Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.ScreenshotDir != "" {
		cfg.ScreenshotDir = fileCfg.ScreenshotDir
	}
	if fileCfg.OCRHelperPath != "" {
		cfg.OCRHelperPath = fileCfg.OCRHelperPath
	}
	if fileCfg.ClaudePath != "" {
		cfg.ClaudePath = fileCfg.ClaudePath
	}
	if fileCfg.MaxFileNameLen > 0 {
		cfg.MaxFileNameLen = fileCfg.MaxFileNameLen
	}
	cfg.Enabled = fileCfg.Enabled

	return cfg
}

func SaveConfig(cfg Config) error {
	if err := os.MkdirAll(configDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}
