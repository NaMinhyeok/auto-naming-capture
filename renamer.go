package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// 스크린샷 파일명에서 날짜를 추출하는 패턴
var datePattern = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)

type RenameResult struct {
	OriginalPath string
	NewPath      string
	Success      bool
	Error        error
}

func ProcessScreenshot(cfg Config, screenshotPath string) RenameResult {
	result := RenameResult{OriginalPath: screenshotPath}

	// 1. OCR 수행
	fmt.Printf("[Renamer] OCR 시작: %s\n", filepath.Base(screenshotPath))
	ocrResult := RunOCR(cfg, screenshotPath)
	if ocrResult.HasText {
		fmt.Printf("[Renamer] OCR 텍스트 추출됨 (%d자)\n", len([]rune(ocrResult.Text)))
	} else {
		fmt.Println("[Renamer] OCR 텍스트 없음 - 이미지 분석으로 진행")
	}

	// 2. AI CLI로 파일명 생성
	fmt.Printf("[Renamer] %s CLI 호출 중...\n", cfg.Provider)
	suggestedName, err := GenerateName(cfg, screenshotPath, ocrResult)
	if err != nil {
		result.Error = fmt.Errorf("naming failed: %w", err)
		fmt.Printf("[Renamer] 네이밍 실패: %v\n", err)
		return result
	}
	fmt.Printf("[Renamer] 제안된 이름: %s\n", suggestedName)

	// 3. 날짜 추출 + 최종 파일명 조합
	date := extractDate(filepath.Base(screenshotPath))
	ext := filepath.Ext(screenshotPath)
	newName := fmt.Sprintf("%s_%s%s", date, suggestedName, ext)

	// 4. 중복 처리 후 리네이밍
	dir := filepath.Dir(screenshotPath)
	newPath := resolveConflict(filepath.Join(dir, newName))

	if err := os.Rename(screenshotPath, newPath); err != nil {
		result.Error = fmt.Errorf("rename failed: %w", err)
		fmt.Printf("[Renamer] 리네이밍 실패: %v\n", err)
		return result
	}

	result.NewPath = newPath
	result.Success = true
	fmt.Printf("[Renamer] 완료: %s → %s\n", filepath.Base(screenshotPath), filepath.Base(newPath))
	return result
}

func extractDate(filename string) string {
	match := datePattern.FindString(filename)
	if match != "" {
		return match
	}
	return time.Now().Format("2006-01-02")
}

func resolveConflict(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 2; i <= 100; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}

	return fmt.Sprintf("%s-%d%s", base, time.Now().UnixNano(), ext)
}
