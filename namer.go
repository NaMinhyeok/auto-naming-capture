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

const systemPrompt = `너는 스크린샷 파일명 생성기야. 스크린샷의 내용을 분석하고 짧고 설명적인 파일명을 생성해.

규칙:
- 파일명만 출력 (확장자 제외, 설명이나 부가 텍스트 없이 파일명 한 줄만)
- 영어 또는 한글 사용
- 공백 대신 하이픈(-) 사용
- 간결하게 (2-5단어)
- 특수문자 사용 금지

예시 출력:
슬랙-대화-정리
github-pr-review
터미널-빌드-에러`

func buildPrompt(imagePath string, ocrResult OCRResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("이미지 경로: %s\n", imagePath))

	if ocrResult.HasText {
		sb.WriteString(fmt.Sprintf("OCR 추출 텍스트:\n%s", ocrResult.Text))
	} else {
		sb.WriteString("OCR 텍스트 없음. 이미지를 직접 분석해서 파일명을 생성해줘.")
	}

	return sb.String()
}

func buildCodexPrompt(ocrResult OCRResult) string {
	// Codex는 system prompt 플래그가 없으므로 prompt에 역할 지시 포함
	var sb strings.Builder
	sb.WriteString(systemPrompt)
	sb.WriteString("\n\n---\n\n")

	if ocrResult.HasText {
		sb.WriteString(fmt.Sprintf("OCR 추출 텍스트:\n%s", ocrResult.Text))
	} else {
		sb.WriteString("OCR 텍스트 없음. 첨부된 이미지를 분석해서 파일명을 생성해줘.")
	}

	return sb.String()
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
