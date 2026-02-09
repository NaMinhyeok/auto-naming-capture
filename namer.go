package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func GenerateName(cfg Config, imagePath string, ocrResult OCRResult) (string, error) {
	switch cfg.Provider {
	case ProviderCodex:
		return generateWithCodex(cfg, imagePath, ocrResult)
	default:
		return generateWithClaude(cfg, imagePath, ocrResult)
	}
}

func generateWithClaude(cfg Config, imagePath string, ocrResult OCRResult) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := buildPrompt(imagePath, ocrResult)

	cmd := exec.CommandContext(ctx, cfg.ClaudePath,
		"-p", prompt,
		"--system-prompt", systemPrompt,
		"--allowedTools", "Read",
		"--output-format", "text",
	)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude cli error: %w", err)
	}

	name := strings.TrimSpace(string(out))
	return SanitizeFilename(name, cfg.MaxFileNameLen), nil
}

func generateWithCodex(cfg Config, imagePath string, ocrResult OCRResult) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := buildCodexPrompt(ocrResult)

	cmd := exec.CommandContext(ctx, cfg.CodexPath,
		"exec",
		"-i", imagePath,
		"--full-auto",
		prompt,
	)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("codex cli error: %w", err)
	}

	name := strings.TrimSpace(string(out))
	return SanitizeFilename(name, cfg.MaxFileNameLen), nil
}

const systemPrompt = `너는 스크린샷 파일명 생성기야. 파일명만 한 줄로 출력해. 그 외 설명, 인사, 부가 텍스트는 절대 출력하지 마.

## 분석 우선순위

1단계: 이미지를 먼저 분석해. 어떤 앱/화면인지, 핵심 시각 요소가 무엇인지 파악해.
2단계: OCR 텍스트가 있으면 보조 참고만 해. 텍스트에 끌려가지 말고 이미지에서 본 내용을 기반으로 판단해.

## 파일명 작성 원칙

구조: [앱/환경]-[핵심내용] (하이픈으로 연결, 2-5단어)
언어: 한글 또는 영어, 자연스러운 쪽으로 선택
금지: 확장자, 특수문자, 공백`

func buildPrompt(imagePath string, ocrResult OCRResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("이미지 경로: %s\n", imagePath))
	sb.WriteString("이미지를 분석하고 파일명을 생성해줘.")

	if ocrResult.HasText {
		// OCR은 보조 참고용으로만 제공, 500자 제한
		text := truncate(ocrResult.Text, 500)
		sb.WriteString(fmt.Sprintf("\n\n(참고용 OCR 텍스트:\n%s)", text))
	}

	return sb.String()
}

func buildCodexPrompt(ocrResult OCRResult) string {
	var sb strings.Builder
	sb.WriteString(systemPrompt)
	sb.WriteString("\n\n---\n\n")
	sb.WriteString("첨부된 이미지를 분석하고 파일명을 생성해줘.")

	if ocrResult.HasText {
		text := truncate(ocrResult.Text, 500)
		sb.WriteString(fmt.Sprintf("\n\n(참고용 OCR 텍스트:\n%s)", text))
	}

	return sb.String()
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

var unsafeChars = regexp.MustCompile(`[^\p{L}\p{N}\-_]`)

func SanitizeFilename(name string, maxLen int) string {
	name = strings.TrimSpace(name)

	// 줄바꿈이 있으면 첫 줄만 사용
	if idx := strings.IndexAny(name, "\n\r"); idx != -1 {
		name = name[:idx]
	}

	// 확장자가 포함되어 있으면 제거
	name = strings.TrimSuffix(name, ".png")
	name = strings.TrimSuffix(name, ".jpg")
	name = strings.TrimSuffix(name, ".jpeg")

	// 공백 → 하이픈
	name = strings.ReplaceAll(name, " ", "-")

	// 안전하지 않은 문자 제거 (한글, 영문, 숫자, 하이픈, 언더스코어만 허용)
	name = unsafeChars.ReplaceAllString(name, "")

	// 연속 하이픈 정리
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	name = strings.Trim(name, "-")

	// 최대 길이 제한 (rune 기준)
	if utf8.RuneCountInString(name) > maxLen {
		runes := []rune(name)
		name = string(runes[:maxLen])
		name = strings.TrimRight(name, "-")
	}

	if name == "" {
		name = "screenshot"
	}

	return name
}
