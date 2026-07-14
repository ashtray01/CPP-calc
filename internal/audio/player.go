// Package audio provides playback of a single MP3 file via Windows MCI.
//
// It extracts the embedded music to a temporary file and plays it
// in a loop using the winmm.dll MCI interface.
package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"calculator/assets"

	"golang.org/x/sys/windows"
)

var (
	winmm         = windows.NewLazySystemDLL("winmm.dll")
	mciSendString = winmm.NewProc("mciSendStringW")
)

// Player manages audio playback via MCI.
// Must be created via New.
type Player struct {
	path    string
	alias   string
	mu      sync.Mutex
	started bool
}

// New creates a new Player, extracting the embedded music to a temp file.
// Returns nil if the embedded music cannot be read or extracted.
func New() *Player {
	data, err := assets.FS.ReadFile("music.mp3")
	if err != nil {
		return nil
	}
	tmpFile := filepath.Join(os.TempDir(), "cp_calc_music.mp3")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return nil
	}
	return &Player{
		path:  tmpFile,
		alias: fmt.Sprintf("cpcalc_%d", os.Getpid()),
	}
}

// Cleanup removes the temporary music file.
// Should be called on program exit via defer.
func (p *Player) Cleanup() {
	if p == nil || p.path == "" {
		return
	}
	p.Stop()
	os.Remove(p.path)
}

func mciSend(cmd string) error {
	cmdPtr, err := syscall.UTF16PtrFromString(cmd)
	if err != nil {
		return fmt.Errorf("UTF16 conversion failed: %w", err)
	}
	ret, _, _ := mciSendString.Call(uintptr(unsafe.Pointer(cmdPtr)), 0, 0, 0)
	if ret != 0 {
		return fmt.Errorf("MCI error (ret=%d): %s", ret, cmd)
	}
	return nil
}

// Play starts looping playback of the extracted music.
// Safe for repeated calls — stops any existing playback first.
func (p *Player) Play() {
	if p == nil || p.path == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopLocked()

	shortPath := strings.ReplaceAll(p.path, "/", "\\")
	if err := mciSend(fmt.Sprintf(`open "%s" type mpegvideo alias %s`, shortPath, p.alias)); err != nil {
		return
	}
	if err := mciSend(fmt.Sprintf("play %s repeat", p.alias)); err != nil {
		mciSend(fmt.Sprintf("close %s", p.alias))
		return
	}
	p.started = true
}

// Stop halts playback and closes the MCI device.
func (p *Player) Stop() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *Player) stopLocked() {
	if !p.started {
		return
	}
	mciSend(fmt.Sprintf("stop %s", p.alias))
	mciSend(fmt.Sprintf("close %s", p.alias))
	p.started = false
}
