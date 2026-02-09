# Auto Naming Capture

macOS 스크린샷을 자동으로 의미있는 이름으로 변경해주는 메뉴바 앱.

`Screenshot 2025-01-15 at 12.30.45.png` → `2025-01-15_슬랙-프로젝트-대화.png`

## How it works

```
스크린샷 촬영 → fsnotify 감지 → Apple Vision OCR → AI 파일명 생성 → 자동 리네이밍
```

1. 스크린샷 디렉토리를 실시간 감시
2. 새 스크린샷 감지 시 Apple Vision으로 OCR 텍스트 추출
3. AI (Claude 또는 Codex)가 이미지 + OCR 텍스트를 분석하여 파일명 제안
4. `YYYY-MM-DD_제안된-이름.png` 형태로 자동 리네이밍

## Requirements

- **macOS** (Apple Vision framework 필요)
- **Go 1.21+**
- **Swift** (Xcode Command Line Tools)
- **Claude CLI** 또는 **Codex CLI** (하나 이상 설치)

## Installation

### From source

```bash
git clone https://github.com/NaMinhyeok/auto-naming-capture.git
cd auto-naming-capture
make build
```

### From release

[Releases](https://github.com/NaMinhyeok/auto-naming-capture/releases) 페이지에서 최신 바이너리를 다운로드하세요.

## Usage

```bash
# 빌드 후 실행
make run

# 또는 직접 실행
./auto-naming-capture
```

메뉴바에 카메라 아이콘이 나타나면 동작 중입니다.

### Menu

| 메뉴 | 설명 |
|------|------|
| ✓ Enabled | 자동 리네이밍 켜기/끄기 |
| Provider | AI 프로바이더 선택 (Claude / Codex) |
| Last: ... | 마지막 리네이밍 결과 |
| Open Screenshot Folder | 스크린샷 폴더 열기 |
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

스크린샷 디렉토리는 macOS 설정에서 자동 감지됩니다.

## AI Providers

| Provider | 이미지 전달 방식 | 필요 CLI |
|----------|-----------------|----------|
| Claude | `--allowedTools "Read"` (파일 직접 읽기) | `claude` |
| Codex | `-i` 플래그 (이미지 첨부) | `codex` |

메뉴바에서 실시간으로 전환 가능합니다.

## Development

```bash
make build    # Swift OCR helper + Go 바이너리 빌드
make test     # 테스트 실행
make run      # 빌드 후 실행
make clean    # 빌드 아티팩트 정리
```

## License

[MIT](LICENSE)
