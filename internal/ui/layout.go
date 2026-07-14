package ui

import (
	"bytes"
	"image"
	_ "image/png"
	"log"

	"calculator/assets"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"
)

// ── Panel builders ──────────────────────────────────────────────────────────

func (c *Calculator) buildTopPanel() Composite {
	return Composite{
		AssignTo:   &c.topPanel,
		Background: SolidColorBrush{Color: ColorDarkRed},
		Layout:     VBox{Alignment: AlignHCenterVNear, Margins: Margins{0, 4, 0, 1}},
		Children: []Widget{
			Label{Text: "★ ★ ★ ★ ★", Font: SimSun(10, false), TextColor: ColorGold, Background: SolidColorBrush{Color: ColorDarkRed}},
			Label{Text: "中华人民共和国超级计算器", Font: SimSun(12, true), TextColor: ColorGold, Background: SolidColorBrush{Color: ColorDarkRed}},
			Label{Text: "v1.0 | 代号: 红龙 🐉", Font: SimSun(9, false), TextColor: ColorGold, Background: SolidColorBrush{Color: ColorDarkRed}},
		},
	}
}

func (c *Calculator) buildDisplayPane() Composite {
	return Composite{
		AssignTo:   &c.displayPane,
		Background: SolidColorBrush{Color: ColorDeepRed},
		Layout:     VBox{Margins: Margins{15, 3, 15, 3}},
		Children: []Widget{
			Label{
				AssignTo:      &c.displayLabel,
				Font:          SimSun(18, true),
				TextColor:     ColorGold,
				Background:    SolidColorBrush{Color: ColorDeepRed},
				MinSize:       Size{0, 36},
				TextAlignment: AlignFar,
			},
		},
	}
}

func (c *Calculator) buildButtonGrid() Composite {
	return Composite{
		AssignTo: &c.btnPane,
		Layout:   VBox{Margins: Margins{15, 0, 15, 0}, Spacing: 3},
		Children: []Widget{
			c.makeRow("七", "八", "九", "÷"),
			c.makeRow("四", "五", "六", "×"),
			c.makeRow("一", "二", "三", "−"),
			c.makeRow("零", ".", "＝", "＋"),
			c.makeLastRow(),
		},
	}
}

func (c *Calculator) buildStatsPane() Composite {
	return Composite{
		AssignTo:   &c.statsPane,
		Background: SolidColorBrush{Color: ColorDarkRed},
		Layout:     HBox{Margins: Margins{15, 2, 15, 4}},
		Children: []Widget{
			Label{
				AssignTo:   &c.statsLabel,
				Font:       SimSun(9, false),
				TextColor:  ColorGold,
				Background: SolidColorBrush{Color: ColorDarkRed},
				Text:       "总计算次数: 0",
			},
		},
	}
}

// ── Button helpers ──────────────────────────────────────────────────────────

func (c *Calculator) makeButton(text string, sz Size, onClick func()) CustomWidget {
	return CustomWidget{
		MinSize:       sz,
		MaxSize:       sz,
		StretchFactor: 0,
		PaintMode:     PaintNoErase,
		Paint: func(canvas *walk.Canvas, bounds walk.Rectangle) error {
			return canvas.DrawText(text, c.btnFont, ColorGold, bounds,
				walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
		},
		OnMouseDown: func(_, _ int, _ walk.MouseButton) { onClick() },
	}
}

func (c *Calculator) makeRow(labels ...string) Composite {
	var children []Widget
	for _, lbl := range labels {
		label := lbl
		children = append(children, c.makeButton(label, Size{c.btnWidth, 28},
			func() { c.onButtonClick(label) }))
	}
	return Composite{
		Layout:   HBox{Spacing: 0, MarginsZero: true},
		Children: children,
	}
}

func (c *Calculator) makeLastRow() Composite {
	b := c.btnWidth
	return Composite{
		Layout: HBox{Spacing: 0, MarginsZero: true},
		Children: []Widget{
			c.makeButton("清除", Size{b*2 + 3, 28}, c.Clear),
			c.makeButton("←", Size{b, 28}, c.Backspace),
			c.makeButton("主席", Size{b, 28}, c.ToggleChairman),
		},
	}
}

func (c *Calculator) onButtonClick(label string) {
	switch label {
	case "÷":
		c.handleOperator("/")
	case "×":
		c.handleOperator("*")
	case "−":
		c.handleOperator("-")
	case "＋":
		c.handleOperator("+")
	case "＝":
		c.Calculate()
	case "清除":
		c.Clear()
	case "←":
		c.Backspace()
	case "主席":
		c.ToggleChairman()
	default:
		c.handleDigit(label)
	}
}

// ── Background ──────────────────────────────────────────────────────────────

func (c *Calculator) applyBackground(w walk.Form, col walk.Color) {
	brush, err := walk.NewSolidColorBrush(col)
	if err != nil {
		return
	}
	w.SetBackground(brush)
}

func (c *Calculator) setBackgroundAll(col walk.Color) {
	brush, err := walk.NewSolidColorBrush(col)
	if err != nil {
		return
	}
	panels := []*walk.Composite{c.topPanel, c.displayPane, c.statsPane, c.btnPane}
	for _, p := range panels {
		if p == nil {
			continue
		}
		p.SetBackground(brush)
		for i := 0; i < p.Children().Len(); i++ {
			if l, ok := p.Children().At(i).(*walk.Label); ok {
				l.SetBackground(brush)
			}
		}
	}
	c.portraitLabel.SetBackground(brush)
	c.displayLabel.SetBackground(brush)
}

// ── Background image ────────────────────────────────────────────────────────

func (c *Calculator) loadLeaderImage() (*walk.Bitmap, error) {
	data, err := assets.FS.ReadFile("leader.png")
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return walk.NewBitmapFromImage(img)
}

// SetBackgroundImage loads leader.png and sets it as the window background.
func (c *Calculator) SetBackgroundImage(winW, winH int) {
	bgBitmap, err := c.loadLeaderImage()
	if err != nil {
		log.Printf("Ошибка загрузки изображения: %v", err)
		return
	}
	dpi := c.MainWindow.DPI()
	physW := winW * dpi / 96
	physH := winH * dpi / 96
	resized, err := walk.NewBitmap(walk.Size{Width: physW, Height: physH})
	if err != nil {
		return
	}

	canvas, err := walk.NewCanvasFromImage(resized)
	if err != nil {
		return
	}
	canvas.DrawImageStretchedPixels(bgBitmap, walk.Rectangle{Width: physW, Height: physH})
	canvas.Dispose()

	if brush, err := walk.NewBitmapBrush(resized); err == nil {
		c.MainWindow.SetBackground(brush)
	}
}

// ── Window centering ────────────────────────────────────────────────────────

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
)

const (
	smCXScreen  = 0
	smCYScreen  = 1
	hwnDTop     = 0
	swpNoSize   = 0x0001
	swpNoZOrder = 0x0004
)

func (c *Calculator) centreWindow() {
	sz := c.MainWindow.Size()
	screenW, _, _ := procGetSystemMetrics.Call(uintptr(smCXScreen))
	screenH, _, _ := procGetSystemMetrics.Call(uintptr(smCYScreen))
	x := (int(screenW) - sz.Width) / 2
	y := (int(screenH) - sz.Height) / 2
	procSetWindowPos.Call(
		uintptr(c.MainWindow.Handle()),
		hwnDTop,
		uintptr(x), uintptr(y),
		0, 0,
		swpNoSize|swpNoZOrder,
	)
}
