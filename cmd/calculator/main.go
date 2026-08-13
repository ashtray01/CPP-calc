// Command calculator 是中文桌面计算器的程序入口。
package main

import (
	"calculator/internal/ui"

	"golang.org/x/sys/windows"
)

func main() {
	shcore := windows.NewLazySystemDLL("shcore.dll")
	setDPIAwareness := shcore.NewProc("SetProcessDpiAwareness")
	setDPIAwareness.Call(1) // 启用按显示器缩放，避免高分屏界面模糊。

	ui.New().Run()
}
