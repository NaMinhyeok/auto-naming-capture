# Auto Naming Capture

macOS 스크린샷을 자동으로 의미있는 이름으로 변경해주는 메뉴바 앱.

```
Screenshot 2026-02-09 at 14.30.45.png → 2026-02-09_슬랙-프로젝트-대화.png
```

## How it works

```
┌─────────────┐    ┌──────────┐    ┌───────────────┐    ┌──────────┐    ┌──────────┐
│  스크린샷    │───▶│ fsnotify │───▶│ Apple Vision  │───▶│ AI 분석  │───▶│ 리네이밍 │
│  촬영       │    │ 감지     │    │ OCR           │    │ Claude/  │    │ 완료     │
│             │    │ +500ms   │    │ (en/ko)       │    │ Codex    │    │          │
└─────────────┘    └──────────┘    └───────────────┘    └──────────┘    └──────────┘
```

1. **감지** — 스크린샷 디렉토리를 실시간 감시 (fsnotify)
2. **OCR** — Apple Vision으로 텍스트 추출 (한국어/영어)
3. **AI 분석** — 이미지를 우선 분석하고, OCR 텍스트를 보조로 참고하여 파일명 생성
4. **리네이밍** — `YYYY-MM-DD_제안된-이름.png` 형태로 자동 변경

## Requirements

- **macOS** (Apple Vision framework)
- **Go 1.25+**
- **Swift** (Xcode Command Line Tools)
- **[Claude CLI](https://github.com/anthropics/claude-code)** 또는 **[Codex CLI](https://github.com/openai/codex)** (하나 이상)

## Installation

### From source

```bash
git clone https://github.com/NaMinhyeok/auto-naming-capture.git
cd auto-naming-capture
make build
```

### From release

[Releases](https://github.com/NaMinhyeok/auto-naming-capture/releases) 페이지에서 최신 바이너리를 다운로드하세요.

```bash
# Apple Silicon (M1/M2/M3/M4)
tar -xzf auto-naming-capture-darwin-arm64.tar.gz

# Intel Mac
tar -xzf auto-naming-capture-darwin-amd64.tar.gz

cd auto-naming-capture-darwin-*/
./auto-naming-capture
```

## Usage

```bash
make run    # 빌드 후 실행
```

메뉴바에 카메라 아이콘이 나타나면 동작 중입니다. 스크린샷을 찍으면 5~15초 후 파일명이 자동으로 변경됩니다.

### Menu

| 메뉴 | 설명 |
|------|------|
| ✓ Enabled | 자동 리네이밍 켜기/끄기 |
| Provider → Claude / Codex | AI 프로바이더 실시간 전환 |
| Last: ... | 마지막 리네이밍 결과 |
| Open Screenshot Folder | Finder에서 스크린샷 폴더 열기 |
| Quit | 앱 종료 |

## Configuration

설정 파일: `~/.config/auto-naming-capture/config.json`

```json
{
  "screenshot_dir": "/Users/you/Desktop",
  "provider": "claude",
  "claude_path": "/usr/local/bin/claude",
  "codex_path": "/usr/local/bin/codex",
  "max_filename_length": 80,
  "enabled": true
}
```

| 필드 | 기본값 | 설명 |
|------|--------|------|
| `screenshot_dir` | macOS 설정 자동 감지 | 스크린샷 저장 경로 |
| `provider` | `"claude"` | AI 프로바이더 (`"claude"` 또는 `"codex"`) |
| `max_filename_length` | `80` | 파일명 최대 길이 (rune 기준) |
| `enabled` | `true` | 자동 리네이밍 활성화 |

## AI Providers

| Provider | 이미지 전달 | CLI 명령어 |
|----------|-----------|------------|
| Claude | `--allowedTools "Read"` + `--system-prompt` | `claude -p "..." --output-format text` |
| Codex | `-i` 플래그 (이미지 직접 첨부) | `codex exec -i <image> --full-auto "..."` |

메뉴바에서 실시간 전환 가능합니다. AI는 **이미지 분석을 우선**하고, OCR 텍스트는 보조로 참고합니다.

## Development

```bash
make build    # Swift OCR helper + Go 바이너리 빌드
make test     # 테스트 실행 (83개 케이스)
make run      # 빌드 후 실행
make clean    # 빌드 아티팩트 정리
```

### Project Structure

```
main.go              메뉴바 앱 진입점 (systray)
watcher.go           파일 시스템 감시 (fsnotify)
ocr.go               Swift OCR helper 호출
namer.go             AI CLI 호출 + 파일명 정제
renamer.go           OCR → AI → 리네이밍 오케스트레이션
config.go            설정 로드/저장
ocr-helper/main.swift  Apple Vision OCR CLI
assets/icon.png      메뉴바 아이콘
```

### Release

태그를 푸시하면 GitHub Actions가 자동으로 arm64/amd64 바이너리를 빌드하고 릴리스합니다.

```bash
git tag v0.2.0
git push origin v0.2.0
```

## Contributing

1. Fork & Clone
2. `make build && make test`로 빌드/테스트 확인
3. 기능 브랜치에서 작업 후 PR

## License

[MIT](LICENSE)
