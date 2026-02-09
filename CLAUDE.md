# CLAUDE.md

## Commands

```bash
make build      # Swift OCR helper + Go 바이너리 빌드 (CGO_ENABLED=1 필요)
make test       # go test ./... -v -count=1
make run        # build + 실행
make clean      # 바이너리 정리
```

개별 빌드:
```bash
# OCR helper만
swiftc -O -o ocr-helper/ocr-helper ocr-helper/main.swift -framework Vision -framework CoreGraphics -framework ImageIO

# Go만
CGO_ENABLED=1 go build -o auto-naming-capture .
```

## Architecture

```
main.go        systray 메뉴바 앱 진입점. 전역 *Config 포인터 + sync.Mutex로 상태 관리
watcher.go     fsnotify로 스크린샷 디렉토리 감시. CREATE → 500ms 대기 → goroutine 처리
ocr.go         os/exec로 Swift OCR helper 호출 (10초 타임아웃)
namer.go       Claude/Codex CLI 호출로 파일명 생성 (60초 타임아웃). SanitizeFilename 포함
renamer.go     OCR → AI → 리네이밍 오케스트레이션. 날짜 추출 + 중복 처리
config.go      JSON 설정 로드/저장 (~/.config/auto-naming-capture/config.json)
ocr-helper/    Swift CLI. Apple Vision VNRecognizeTextRequest (en-US, ko-KR)
assets/        go:embed용 메뉴바 아이콘 (16x16 PNG)
```

데이터 흐름: `fsnotify CREATE → isScreenshot() → 500ms sleep → ProcessScreenshot(cfg snapshot) → RunOCR → GenerateName → os.Rename`

## Key Patterns

- **Config는 포인터 공유**: `main.go`의 `*Config`를 Watcher가 참조. 메뉴 토글 시 cfgLock 하에 변경하면 다음 스크린샷부터 반영
- **Config snapshot**: Watcher goroutine에서 처리 시작 시 `cfgLock.Lock()` → `snapshot := *w.cfg` → `Unlock()`. 처리 중 설정 변경 영향 없음
- **Provider 분기**: `namer.go`의 `GenerateName()`이 `cfg.Provider`로 분기. Claude는 `--system-prompt` + `--allowedTools "Read"`, Codex는 `-i` 플래그로 이미지 전달
- **이미지 우선 프롬프트**: system prompt에서 이미지 분석 1단계, OCR 텍스트 2단계(보조). OCR 텍스트는 500자 truncate

## Gotchas

- `getlantern/systray`는 CGO 필수. `CGO_ENABLED=0`이면 빌드 실패
- OCR helper 경로 탐지: `ocr-helper/ocr-helper` (디렉토리/바이너리). `os.Stat`만으로는 디렉토리도 통과하므로 `info.IsDir()` 체크 필요
- `SanitizeFilename`은 빈 결과 시 "screenshot" 반환. 절대 빈 문자열 반환 안 함
- 스크린샷 패턴: 영어 `Screenshot YYYY-MM-DD...` + 한국어 `스크린샷 YYYY-MM-DD...`. 대소문자 구분함
- macOS에서 `defaults read com.apple.screencapture location` 실패 시 ~/Desktop fallback

## Testing

```bash
go test ./... -v -count=1
```

테스트 파일: `namer_test.go`, `renamer_test.go`, `watcher_test.go`, `config_test.go`
- `SanitizeFilename`: 유니코드, 특수문자, maxLen 경계값
- `resolveConflict`: `t.TempDir()`로 실제 파일 시스템 테스트
- `isScreenshot`: 매칭/비매칭 패턴 (이미 리네이밍된 파일 포함)

## Code Style

- 한국어 로그 메시지 (`[Watcher]`, `[Renamer]`, `[OCR]` 접두사)
- table-driven tests
- 에러 시 원본 유지, 앱 크래시 방지 원칙
