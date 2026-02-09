package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type OCRResult struct {
	Text    string
	HasText bool
}

func RunOCR(cfg Config, imagePath string) OCRResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cfg.OCRHelperPath, imagePath)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("[OCR] error: %v\n", err)
		return OCRResult{}
	}

	text := strings.TrimSpace(string(out))
	// 3글자 미만이면 의미있는 텍스트가 아닌 것으로 판단
	if len([]rune(text)) < 3 {
		return OCRResult{Text: text, HasText: false}
	}

	return OCRResult{Text: text, HasText: true}
}
