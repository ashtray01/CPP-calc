package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// 调色板以深红、暖金和米白为主，兼顾庄重感与现代界面的清晰度。
var (
	ColorCanvas       = walk.RGB(0x16, 0x0B, 0x0F)
	ColorSurface      = walk.RGB(0x24, 0x12, 0x18)
	ColorSurfaceLight = walk.RGB(0x31, 0x1A, 0x21)
	ColorKey          = walk.RGB(0x3B, 0x22, 0x29)
	ColorKeyHover     = walk.RGB(0x50, 0x2C, 0x35)
	ColorRed          = walk.RGB(0xC9, 0x26, 0x32)
	ColorRedHover     = walk.RGB(0xE1, 0x38, 0x43)
	ColorGold         = walk.RGB(0xF2, 0xC1, 0x66)
	ColorText         = walk.RGB(0xFA, 0xF5, 0xEC)
	ColorMuted        = walk.RGB(0xB9, 0xA3, 0xA6)
	ColorBorder       = walk.RGB(0x57, 0x35, 0x3C)
)

// YaHei 创建适合中文界面的微软雅黑字体。
func YaHei(size int, bold bool) Font {
	return Font{Family: "Microsoft YaHei UI", PointSize: size, Bold: bold}
}
