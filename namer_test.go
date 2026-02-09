package main

import (
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		// 기본 동작
		{"normal english", "github-pr-review", 80, "github-pr-review"},
		{"normal korean", "슬랙-대화-정리", 80, "슬랙-대화-정리"},
		{"mixed", "vscode-에러-로그", 80, "vscode-에러-로그"},

		// 공백 처리
		{"spaces to hyphens", "hello world test", 80, "hello-world-test"},
		{"leading trailing spaces", "  hello  ", 80, "hello"},
		{"multiple spaces", "a  b   c", 80, "a-b-c"},

		// 특수문자 제거
		{"special chars", "hello@world#test!", 80, "helloworldtest"},
		{"dots in name", "file.name.here", 80, "filenamehere"},
		{"parentheses", "screenshot (1)", 80, "screenshot-1"},
		{"brackets and quotes", `file[1]"test"`, 80, "file1test"},
		{"unicode symbols", "file★test◆name", 80, "filetestname"},

		// 확장자 제거
		{"remove .png", "my-screenshot.png", 80, "my-screenshot"},
		{"remove .jpg", "photo.jpg", 80, "photo"},
		{"remove .jpeg", "image.jpeg", 80, "image"},
		{"keep .gif", "animation.gif", 80, "animationgif"},
		{"double extension", "file.png.png", 80, "filepng"},

		// 줄바꿈 처리
		{"newline takes first line", "first-line\nsecond-line", 80, "first-line"},
		{"carriage return", "first\rsecond", 80, "first"},
		{"crlf", "first\r\nsecond", 80, "first"},

		// 빈/극단적 입력
		{"empty string", "", 80, "screenshot"},
		{"only spaces", "   ", 80, "screenshot"},
		{"only special chars", "!@#$%^&*()", 80, "screenshot"},
		{"only hyphens", "---", 80, "screenshot"},
		{"single char", "a", 80, "a"},

		// 연속 하이픈 정리
		{"double hyphens", "a--b", 80, "a-b"},
		{"triple hyphens", "a---b", 80, "a-b"},
		{"mixed special creating hyphens", "a @ b ! c", 80, "a-b-c"},

		// 한글 보존
		{"korean full", "카카오톡-그룹채팅-일정공유", 80, "카카오톡-그룹채팅-일정공유"},
		{"korean with special", "카카오톡 (대화)", 80, "카카오톡-대화"},
		{"japanese", "テスト-ファイル", 80, "テスト-ファイル"},

		// maxLen 경계값
		{"exact maxLen", "abcde", 5, "abcde"},
		{"one over maxLen", "abcdef", 5, "abcde"},
		{"maxLen 1", "hello", 1, "h"},
		{"korean maxLen rune", "가나다라마바", 5, "가나다라마"},
		{"truncate at hyphen", "abc-d", 4, "abc"},

		// maxLen에서 하이픈으로 끝나는 경우
		{"truncate trailing hyphen", "abcd-efgh", 5, "abcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename_NeverEmpty(t *testing.T) {
	// 어떤 입력이든 빈 문자열을 반환하면 안 됨
	inputs := []string{"", "   ", "!!!", "...", "\n\n", "---", ".png", "@#$"}
	for _, input := range inputs {
		got := SanitizeFilename(input, 80)
		if got == "" {
			t.Errorf("SanitizeFilename(%q, 80) returned empty string", input)
		}
	}
}

func TestSanitizeFilename_MaxLenRespected(t *testing.T) {
	// 긴 한글 문자열도 maxLen(rune 기준)을 초과하면 안 됨
	long := strings.Repeat("가", 200)
	got := SanitizeFilename(long, 80)
	runes := []rune(got)
	if len(runes) > 80 {
		t.Errorf("SanitizeFilename with maxLen=80 returned %d runes", len(runes))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{"short string unchanged", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated with ellipsis", "hello world", 5, "hello..."},
		{"empty string", "", 5, ""},
		{"zero max", "hello", 0, "..."},
		{"korean truncate", "가나다라마바사", 3, "가나다..."},
		{"single rune limit", "abcdef", 1, "a..."},

		// 경계값
		{"max equals length", "abc", 3, "abc"},
		{"max one less", "abc", 2, "ab..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxRunes)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxRunes, got, tt.want)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	t.Run("with OCR text", func(t *testing.T) {
		result := buildPrompt("/path/to/image.png", OCRResult{Text: "Hello World", HasText: true})

		if !strings.Contains(result, "/path/to/image.png") {
			t.Error("prompt should contain image path")
		}
		if !strings.Contains(result, "참고용 OCR 텍스트") {
			t.Error("prompt should contain OCR reference label")
		}
		if !strings.Contains(result, "Hello World") {
			t.Error("prompt should contain OCR text")
		}
	})

	t.Run("without OCR text", func(t *testing.T) {
		result := buildPrompt("/path/to/image.png", OCRResult{HasText: false})

		if !strings.Contains(result, "/path/to/image.png") {
			t.Error("prompt should contain image path")
		}
		if strings.Contains(result, "참고용 OCR 텍스트") {
			t.Error("prompt should NOT contain OCR section when no text")
		}
	})

	t.Run("long OCR text truncated", func(t *testing.T) {
		longText := strings.Repeat("가", 1000)
		result := buildPrompt("/img.png", OCRResult{Text: longText, HasText: true})

		// 500자 + "..." 이므로 원본 1000자보다 짧아야 함
		if strings.Contains(result, longText) {
			t.Error("prompt should truncate long OCR text")
		}
		if !strings.Contains(result, "...") {
			t.Error("truncated text should end with ellipsis")
		}
	})
}

func TestBuildCodexPrompt(t *testing.T) {
	t.Run("includes system prompt", func(t *testing.T) {
		result := buildCodexPrompt(OCRResult{HasText: false})

		if !strings.Contains(result, "스크린샷 파일명 생성기") {
			t.Error("codex prompt should include system prompt content")
		}
		if !strings.Contains(result, "---") {
			t.Error("codex prompt should have separator between system and user")
		}
	})

	t.Run("with OCR text", func(t *testing.T) {
		result := buildCodexPrompt(OCRResult{Text: "some text", HasText: true})

		if !strings.Contains(result, "참고용 OCR 텍스트") {
			t.Error("codex prompt should contain OCR text when available")
		}
	})
}
