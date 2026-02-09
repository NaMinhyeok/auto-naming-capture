package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		// 정상 케이스
		{"english screenshot", "Screenshot 2025-01-15 at 12.30.45.png", "2025-01-15"},
		{"korean screenshot", "스크린샷 2025-01-15 오후 12.30.45.png", "2025-01-15"},
		{"date only", "2024-12-31.png", "2024-12-31"},

		// 여러 날짜가 있으면 첫 번째
		{"multiple dates", "2024-01-01_backup_2024-06-15.png", "2024-01-01"},

		// 경계값 날짜
		{"year boundary", "Screenshot 2025-12-31 at 23.59.59.png", "2025-12-31"},
		{"new year", "Screenshot 2025-01-01 at 00.00.01.png", "2025-01-01"},
		{"leap year", "Screenshot 2024-02-29 at 10.00.00.png", "2024-02-29"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDate(tt.filename)
			if got != tt.want {
				t.Errorf("extractDate(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}

	t.Run("no date returns today", func(t *testing.T) {
		got := extractDate("random-file.png")
		// 형식이 YYYY-MM-DD인지만 확인
		if len(got) != 10 || got[4] != '-' || got[7] != '-' {
			t.Errorf("extractDate with no date should return today's date format, got %q", got)
		}
	})

	t.Run("partial date not matched", func(t *testing.T) {
		got := extractDate("file-2025-1-5.png") // 월/일이 한 자리
		// 패턴이 \d{4}-\d{2}-\d{2}이므로 매칭 안 됨 → 오늘 날짜
		if len(got) != 10 || got[4] != '-' || got[7] != '-' {
			t.Errorf("partial date should fallback to today, got %q", got)
		}
	})
}

func TestResolveConflict(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.png")
		got := resolveConflict(path)
		if got != path {
			t.Errorf("no conflict: got %q, want %q", got, path)
		}
	})

	t.Run("single conflict appends -2", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.png")
		os.WriteFile(path, []byte{}, 0644)

		got := resolveConflict(path)
		want := filepath.Join(dir, "test-2.png")
		if got != want {
			t.Errorf("single conflict: got %q, want %q", got, want)
		}
	})

	t.Run("multiple conflicts increment", func(t *testing.T) {
		dir := t.TempDir()
		base := filepath.Join(dir, "test.png")
		os.WriteFile(base, []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "test-2.png"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "test-3.png"), []byte{}, 0644)

		got := resolveConflict(base)
		want := filepath.Join(dir, "test-4.png")
		if got != want {
			t.Errorf("multiple conflicts: got %q, want %q", got, want)
		}
	})

	t.Run("preserves extension", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "image.jpeg")
		os.WriteFile(path, []byte{}, 0644)

		got := resolveConflict(path)
		if !strings.HasSuffix(got, ".jpeg") {
			t.Errorf("should preserve .jpeg extension, got %q", got)
		}
	})

	t.Run("no extension file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "noext")
		os.WriteFile(path, []byte{}, 0644)

		got := resolveConflict(path)
		want := filepath.Join(dir, "noext-2")
		if got != want {
			t.Errorf("no extension: got %q, want %q", got, want)
		}
	})

	t.Run("gap in numbering fills correctly", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "test.png"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "test-2.png"), []byte{}, 0644)
		// test-3.png 없음
		os.WriteFile(filepath.Join(dir, "test-4.png"), []byte{}, 0644)

		got := resolveConflict(filepath.Join(dir, "test.png"))
		want := filepath.Join(dir, "test-3.png")
		if got != want {
			t.Errorf("gap fill: got %q, want %q", got, want)
		}
	})

	t.Run("korean filename conflict", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "2025-01-15_슬랙-대화.png")
		os.WriteFile(path, []byte{}, 0644)

		got := resolveConflict(path)
		want := filepath.Join(dir, "2025-01-15_슬랙-대화-2.png")
		if got != want {
			t.Errorf("korean conflict: got %q, want %q", got, want)
		}
	})
}
