package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// 영어/한국어 macOS 스크린샷 파일명 패턴
var screenshotPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^Screenshot \d{4}-\d{2}-\d{2}.*\.(png|jpg|jpeg)$`),
	regexp.MustCompile(`^스크린샷 \d{4}-\d{2}-\d{2}.*\.(png|jpg|jpeg)$`),
}

type Watcher struct {
	cfg        *Config
	cfgLock    *sync.Mutex
	fsWatcher  *fsnotify.Watcher
	onRenamed  func(RenameResult)
	processing map[string]bool
}

func NewWatcher(cfg *Config, cfgLock *sync.Mutex, onRenamed func(RenameResult)) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		cfg:        cfg,
		cfgLock:    cfgLock,
		fsWatcher:  fsw,
		onRenamed:  onRenamed,
		processing: make(map[string]bool),
	}, nil
}

func (w *Watcher) Start() error {
	if err := w.fsWatcher.Add(w.cfg.ScreenshotDir); err != nil {
		return fmt.Errorf("failed to watch %s: %w", w.cfg.ScreenshotDir, err)
	}
	fmt.Printf("[Watcher] 감시 시작: %s\n", w.cfg.ScreenshotDir)

	go w.loop()
	return nil
}

func (w *Watcher) Stop() {
	w.fsWatcher.Close()
}

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				w.handleCreate(event.Name)
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("[Watcher] error: %v\n", err)
		}
	}
}

func (w *Watcher) handleCreate(path string) {
	filename := filepath.Base(path)

	if !isScreenshot(filename) {
		return
	}

	// 이미 처리 중인 파일은 무시
	if w.processing[path] {
		return
	}
	w.processing[path] = true

	fmt.Printf("[Watcher] 스크린샷 감지: %s\n", filename)

	// 500ms 대기 후 비동기 처리 (파일 쓰기 완료 대기)
	go func() {
		time.Sleep(500 * time.Millisecond)
		defer func() { delete(w.processing, path) }()

		// 현재 설정의 스냅샷을 lock 하에 복사
		w.cfgLock.Lock()
		snapshot := *w.cfg
		w.cfgLock.Unlock()

		if !snapshot.Enabled {
			fmt.Println("[Watcher] 비활성 상태 - 건너뜀")
			return
		}

		result := ProcessScreenshot(snapshot, path)
		if w.onRenamed != nil {
			w.onRenamed(result)
		}
	}()
}

func isScreenshot(filename string) bool {
	for _, p := range screenshotPatterns {
		if p.MatchString(filename) {
			return true
		}
	}
	return false
}
