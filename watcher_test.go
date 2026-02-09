package main

import "testing"

func TestIsScreenshot(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// 영어 패턴 매칭
		{"english png", "Screenshot 2025-01-15 at 12.30.45.png", true},
		{"english jpg", "Screenshot 2025-01-15 at 12.30.45.jpg", true},
		{"english jpeg", "Screenshot 2025-01-15 at 12.30.45.jpeg", true},
		{"english with seconds", "Screenshot 2025-12-31 at 23.59.59.png", true},

		// 한국어 패턴 매칭
		{"korean png", "스크린샷 2025-01-15 오후 12.30.45.png", true},
		{"korean jpg", "스크린샷 2025-01-15 오전 9.00.00.jpg", true},
		{"korean without ampm", "스크린샷 2025-06-30 12.30.45.png", true},

		// 비매칭 - 다른 파일
		{"regular png", "photo.png", false},
		{"document", "report.pdf", false},
		{"partial match", "Screenshot.png", false},
		{"no date", "Screenshot something.png", false},

		// 비매칭 - 유사하지만 다른 패턴
		{"lowercase screenshot", "screenshot 2025-01-15 at 12.30.45.png", false},
		{"missing space", "Screenshot2025-01-15 at 12.30.45.png", false},
		{"wrong date format", "Screenshot 25-01-15 at 12.30.45.png", false},
		{"already renamed", "2025-01-15_슬랙-대화.png", false},

		// 비매칭 - 지원하지 않는 확장자
		{"gif extension", "Screenshot 2025-01-15 at 12.30.45.gif", false},
		{"webp extension", "Screenshot 2025-01-15 at 12.30.45.webp", false},
		{"no extension", "Screenshot 2025-01-15 at 12.30.45", false},

		// 경계값
		{"year boundary low", "Screenshot 0000-01-01 at 00.00.00.png", true},
		{"year boundary high", "Screenshot 9999-12-31 at 23.59.59.png", true},
		{"empty string", "", false},

		// 이미 리네이밍된 파일은 무시해야 함
		{"renamed korean", "2025-01-15_카카오톡-대화.png", false},
		{"renamed english", "2025-01-15_github-pr-review.png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isScreenshot(tt.filename)
			if got != tt.want {
				t.Errorf("isScreenshot(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
