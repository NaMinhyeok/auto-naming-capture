package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
)

//go:embed assets/icon.png
var iconData []byte

var (
	cfg     Config
	cfgLock sync.Mutex
	watcher *Watcher
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	cfg = LoadConfig()

	systray.SetIcon(iconData)
	systray.SetTooltip("Auto Naming Capture")

	mEnabled := systray.AddMenuItem("✓ Enabled", "Toggle auto-renaming")
	systray.AddSeparator()
	mProvider := systray.AddMenuItem("Provider", "AI Provider")
	mClaude := mProvider.AddSubMenuItem("  Claude", "Use Claude CLI")
	mCodex := mProvider.AddSubMenuItem("  Codex", "Use OpenAI Codex CLI")
	systray.AddSeparator()
	mLast := systray.AddMenuItem("Last: (none)", "Last renamed file")
	mLast.Disable()
	systray.AddSeparator()
	mOpenFolder := systray.AddMenuItem("Open Screenshot Folder", "Open in Finder")
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About Auto Naming Capture")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// 상태에 따라 메뉴 표시 업데이트
	updateEnabledMenu(mEnabled, cfg.Enabled)
	updateProviderMenu(mClaude, mCodex, cfg.Provider)

	// Watcher 시작
	var err error
	watcher, err = NewWatcher(cfg, func(result RenameResult) {
		if result.Success {
			mLast.SetTitle(fmt.Sprintf("Last: %s", filepath.Base(result.NewPath)))
		} else if result.Error != nil {
			mLast.SetTitle(fmt.Sprintf("Last: error - %s", result.Error))
		}
	})
	if err != nil {
		fmt.Printf("Failed to create watcher: %v\n", err)
		return
	}

	if err := watcher.Start(); err != nil {
		fmt.Printf("Failed to start watcher: %v\n", err)
	}

	go func() {
		for {
			select {
			case <-mEnabled.ClickedCh:
				cfgLock.Lock()
				cfg.Enabled = !cfg.Enabled
				updateEnabledMenu(mEnabled, cfg.Enabled)
				SaveConfig(cfg)
				cfgLock.Unlock()

			case <-mClaude.ClickedCh:
				cfgLock.Lock()
				cfg.Provider = ProviderClaude
				updateProviderMenu(mClaude, mCodex, cfg.Provider)
				SaveConfig(cfg)
				cfgLock.Unlock()

			case <-mCodex.ClickedCh:
				cfgLock.Lock()
				cfg.Provider = ProviderCodex
				updateProviderMenu(mClaude, mCodex, cfg.Provider)
				SaveConfig(cfg)
				cfgLock.Unlock()

			case <-mOpenFolder.ClickedCh:
				openFolder(cfg.ScreenshotDir)

			case <-mAbout.ClickedCh:
				showAbout()

			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	if watcher != nil {
		watcher.Stop()
	}
	fmt.Println("Auto Naming Capture 종료")
}

func updateEnabledMenu(m *systray.MenuItem, enabled bool) {
	if enabled {
		m.SetTitle("✓ Enabled")
	} else {
		m.SetTitle("  Disabled")
	}
}

func updateProviderMenu(mClaude, mCodex *systray.MenuItem, provider Provider) {
	if provider == ProviderCodex {
		mClaude.SetTitle("  Claude")
		mCodex.SetTitle("✓ Codex")
	} else {
		mClaude.SetTitle("✓ Claude")
		mCodex.SetTitle("  Codex")
	}
}

func openFolder(path string) {
	if runtime.GOOS == "darwin" {
		exec.Command("open", path).Start()
	}
}

func showAbout() {
	if runtime.GOOS == "darwin" {
		exec.Command("osascript", "-e",
			`display dialog "Auto Naming Capture\n\nScreenshot auto-renaming using OCR + Claude AI\n\nVersion 1.0.0" with title "About" buttons {"OK"} default button "OK"`).Start()
	}
}
