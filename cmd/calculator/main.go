// Command calculator — 中华人民共和国超级计算器
//
// A Chinese-language desktop calculator with Chairman Mao Zedong theme mode.
// Built with lxn/walk native Windows UI toolkit.
package main

import (
	"os"
	"path/filepath"

	"calculator/internal/ui"

	"golang.org/x/sys/windows"
)

func main() {
	shcore := windows.NewLazySystemDLL("shcore.dll")
	procSetDPI := shcore.NewProc("SetProcessDpiAwareness")
	procSetDPI.Call(1) // PROCESS_PER_MONITOR_DPI_AWARE

	calc := ui.New()
	calc.Run()

	// Cleanup temp music file on normal exit
	tmpFile := filepath.Join(os.TempDir(), "cp_calc_music.mp3")
	os.Remove(tmpFile)
}
