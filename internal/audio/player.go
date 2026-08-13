// Package audio 通过 Windows MCI 循环播放嵌入的 MP3 音乐。
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

// Player 管理临时音乐文件和 MCI 播放状态。
type Player struct {
	path    string
	alias   string
	mu      sync.Mutex
	started bool
}

// New 将嵌入音乐写入临时目录，并创建播放器。
func New() *Player {
	data, err := assets.FS.ReadFile("music.mp3")
	if err != nil {
		return nil
	}
	temporaryFile := filepath.Join(os.TempDir(), fmt.Sprintf("cp_calc_music_%d.mp3", os.Getpid()))
	if err := os.WriteFile(temporaryFile, data, 0600); err != nil {
		return nil
	}
	return &Player{path: temporaryFile, alias: fmt.Sprintf("cpcalc_%d", os.Getpid())}
}

// Cleanup 停止播放并删除临时文件。
func (p *Player) Cleanup() {
	if p == nil || p.path == "" {
		return
	}
	p.Stop()
	_ = os.Remove(p.path)
}

func mciSend(command string) error {
	commandPointer, err := syscall.UTF16PtrFromString(command)
	if err != nil {
		return fmt.Errorf("UTF-16 转换失败：%w", err)
	}
	result, _, _ := mciSendString.Call(uintptr(unsafe.Pointer(commandPointer)), 0, 0, 0)
	if result != 0 {
		return fmt.Errorf("MCI 命令失败（代码 %d）：%s", result, command)
	}
	return nil
}

// Play 从头开始循环播放音乐。
func (p *Player) Play() {
	if p == nil || p.path == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()

	path := strings.ReplaceAll(p.path, "/", "\\")
	if err := mciSend(fmt.Sprintf(`open "%s" type mpegvideo alias %s`, path, p.alias)); err != nil {
		return
	}
	if err := mciSend(fmt.Sprintf("play %s repeat", p.alias)); err != nil {
		_ = mciSend(fmt.Sprintf("close %s", p.alias))
		return
	}
	p.started = true
}

// Stop 停止播放并关闭 MCI 设备。
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
	_ = mciSend(fmt.Sprintf("stop %s", p.alias))
	_ = mciSend(fmt.Sprintf("close %s", p.alias))
	p.started = false
}
