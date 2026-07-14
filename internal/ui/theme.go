// Package ui implements the calculator main window and all UI components.
package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Theme colours — imperial Chinese red & gold palette.
var (
	ColorGold    = walk.RGB(0xFF, 0xD7, 0x00)
	ColorDarkRed = walk.RGB(0x8B, 0x00, 0x00)
	ColorDeepRed = walk.RGB(0x4A, 0x00, 0x00)
	ColorRed     = walk.RGB(0xFF, 0x00, 0x00)
)

// SimSun creates a SimSun font with the given size and boldness.
func SimSun(size int, bold bool) Font {
	return Font{Family: "SimSun", PointSize: size, Bold: bold}
}
