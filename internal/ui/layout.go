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

type keyStyle int

const (
	keyNormal keyStyle = iota
	keyFunction
	keyOperator
	keyPrimary
)

type calculatorKey struct {
	label  string
	action string
	style  keyStyle
	rect   walk.Rectangle
}

// buildCanvas 创建单一自绘画布，避免旧式原生按钮带来的 Windows 95 观感。
func (c *Calculator) buildCanvas() CustomWidget {
	return CustomWidget{
		AssignTo:            &c.canvas,
		PaintMode:           PaintBuffered,
		InvalidatesOnResize: true,
		Paint:               c.paint,
		OnMouseMove:         c.onMouseMove,
		OnMouseDown:         c.onMouseDown,
		OnMouseUp:           c.onMouseUp,
	}
}

// loadLeaderImage 从嵌入资源加载主席照片；资源本身不会被修改。
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

func (c *Calculator) paint(canvas *walk.Canvas, _ walk.Rectangle) error {
	bounds := c.canvas.ClientBounds()
	fillRounded(canvas, ColorCanvas, bounds, 0)

	w := bounds.Width
	h := bounds.Height
	margin := 24
	headerBottom := 78
	leftWidth := 330
	gap := 20
	left := walk.Rectangle{X: margin, Y: 92, Width: leftWidth, Height: h - 116}
	right := walk.Rectangle{X: margin + leftWidth + gap, Y: 92, Width: w - margin*2 - leftWidth - gap, Height: h - 116}

	c.paintHeader(canvas, w, headerBottom)
	c.paintPortraitCard(canvas, left)
	c.paintCalculatorCard(canvas, right)
	return nil
}

func (c *Calculator) paintHeader(canvas *walk.Canvas, width, bottom int) {
	drawText(canvas, "中华人民共和国", c.fontSmallBold, ColorGold,
		walk.Rectangle{X: 24, Y: 20, Width: 250, Height: 22}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)
	drawText(canvas, "中文计算器", c.fontTitle, ColorText,
		walk.Rectangle{X: 24, Y: 39, Width: 300, Height: 34}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)

	badge := walk.Rectangle{X: width - 152, Y: 27, Width: 128, Height: 34}
	fillRounded(canvas, ColorSurfaceLight, badge, 17)
	dotColor := ColorMuted
	status := "音乐模式关闭"
	if c.chairman {
		dotColor = ColorRed
		status = "音乐模式开启"
	}
	fillRounded(canvas, dotColor, walk.Rectangle{X: badge.X + 12, Y: badge.Y + 13, Width: 8, Height: 8}, 4)
	drawText(canvas, status, c.fontTiny, ColorText,
		walk.Rectangle{X: badge.X + 27, Y: badge.Y, Width: 92, Height: badge.Height}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)
	drawLine(canvas, ColorBorder, walk.Point{X: 24, Y: bottom}, walk.Point{X: width - 24, Y: bottom})
}

func (c *Calculator) paintPortraitCard(canvas *walk.Canvas, card walk.Rectangle) {
	fillRounded(canvas, ColorSurface, card, 18)
	inner := walk.Rectangle{X: card.X + 12, Y: card.Y + 12, Width: card.Width - 24, Height: 421}
	if c.leader != nil {
		_ = canvas.DrawImageStretched(c.leader, inner)
	}
	drawRoundedBorder(canvas, ColorBorder, inner, 12)

	textY := inner.Y + inner.Height + 14
	drawText(canvas, "主席模式", c.fontCardTitle, ColorText,
		walk.Rectangle{X: card.X + 18, Y: textY, Width: 150, Height: 26}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)
	drawText(canvas, "保留原始照片与背景音乐", c.fontSmall, ColorMuted,
		walk.Rectangle{X: card.X + 18, Y: textY + 27, Width: card.Width - 36, Height: 22}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)

	mode := walk.Rectangle{X: card.X + 18, Y: card.Y + card.Height - 52, Width: card.Width - 36, Height: 36}
	color := ColorSurfaceLight
	if c.chairman {
		color = ColorRed
	}
	if c.hoverAction == "chairman" {
		if c.chairman {
			color = ColorRedHover
		} else {
			color = ColorKeyHover
		}
	}
	fillRounded(canvas, color, mode, 10)
	label := "开启音乐模式"
	if c.chairman {
		label = "关闭音乐模式"
	}
	drawText(canvas, label, c.fontSmallBold, ColorText, mode, walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
	c.modeRect = mode
}

func (c *Calculator) paintCalculatorCard(canvas *walk.Canvas, card walk.Rectangle) {
	fillRounded(canvas, ColorSurface, card, 18)
	display := walk.Rectangle{X: card.X + 20, Y: card.Y + 20, Width: card.Width - 40, Height: 140}
	fillRounded(canvas, ColorSurfaceLight, display, 14)
	drawText(canvas, "当前输入", c.fontTiny, ColorMuted,
		walk.Rectangle{X: display.X + 18, Y: display.Y + 12, Width: 120, Height: 20}, walk.TextLeft|walk.TextVCenter|walk.TextSingleLine)
	drawText(canvas, c.historyText(), c.fontSmall, ColorMuted,
		walk.Rectangle{X: display.X + 18, Y: display.Y + 34, Width: display.Width - 36, Height: 24}, walk.TextRight|walk.TextVCenter|walk.TextSingleLine)
	drawText(canvas, c.displayText(), c.fontDisplay, ColorText,
		walk.Rectangle{X: display.X + 18, Y: display.Y + 61, Width: display.Width - 36, Height: 58}, walk.TextRight|walk.TextVCenter|walk.TextSingleLine)
	drawText(canvas, c.countText(), c.fontTiny, ColorGold,
		walk.Rectangle{X: display.X + 18, Y: display.Y + 116, Width: display.Width - 36, Height: 18}, walk.TextRight|walk.TextVCenter|walk.TextSingleLine)

	keysTop := display.Y + display.Height + 18
	keysBottom := card.Y + card.Height - 20
	c.keys = c.layoutKeys(walk.Rectangle{X: display.X, Y: keysTop, Width: display.Width, Height: keysBottom - keysTop})
	for _, key := range c.keys {
		c.paintKey(canvas, key)
	}
}

func (c *Calculator) layoutKeys(area walk.Rectangle) []calculatorKey {
	rows := [][]calculatorKey{
		{{"清除", "clear", keyFunction, walk.Rectangle{}}, {"正负", "sign", keyFunction, walk.Rectangle{}}, {"退格", "backspace", keyFunction, walk.Rectangle{}}, {"÷", "/", keyOperator, walk.Rectangle{}}},
		{{"七", "七", keyNormal, walk.Rectangle{}}, {"八", "八", keyNormal, walk.Rectangle{}}, {"九", "九", keyNormal, walk.Rectangle{}}, {"×", "*", keyOperator, walk.Rectangle{}}},
		{{"四", "四", keyNormal, walk.Rectangle{}}, {"五", "五", keyNormal, walk.Rectangle{}}, {"六", "六", keyNormal, walk.Rectangle{}}, {"−", "-", keyOperator, walk.Rectangle{}}},
		{{"一", "一", keyNormal, walk.Rectangle{}}, {"二", "二", keyNormal, walk.Rectangle{}}, {"三", "三", keyNormal, walk.Rectangle{}}, {"＋", "+", keyOperator, walk.Rectangle{}}},
		{{"零", "零", keyNormal, walk.Rectangle{}}, {"点", ".", keyNormal, walk.Rectangle{}}, {"百分比", "percent", keyFunction, walk.Rectangle{}}, {"＝", "equals", keyPrimary, walk.Rectangle{}}},
	}
	gap := 10
	keyWidth := (area.Width - gap*3) / 4
	keyHeight := (area.Height - gap*4) / 5
	keys := make([]calculatorKey, 0, 20)
	for rowIndex, row := range rows {
		for columnIndex, key := range row {
			key.rect = walk.Rectangle{
				X:      area.X + columnIndex*(keyWidth+gap),
				Y:      area.Y + rowIndex*(keyHeight+gap),
				Width:  keyWidth,
				Height: keyHeight,
			}
			keys = append(keys, key)
		}
	}
	return keys
}

func (c *Calculator) paintKey(canvas *walk.Canvas, key calculatorKey) {
	color := ColorKey
	textColor := ColorText
	switch key.style {
	case keyFunction:
		color = ColorSurfaceLight
		textColor = ColorGold
	case keyOperator:
		color = walk.RGB(0x50, 0x20, 0x26)
		textColor = ColorGold
	case keyPrimary:
		color = ColorRed
	}
	if c.hoverAction == key.action {
		if key.style == keyPrimary {
			color = ColorRedHover
		} else {
			color = ColorKeyHover
		}
	}
	if c.pressedAction == key.action {
		color = ColorGold
		textColor = ColorCanvas
	}
	fillRounded(canvas, color, key.rect, 11)
	drawText(canvas, key.label, c.fontButton, textColor, key.rect, walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
}

func (c *Calculator) hitTest(x, y int) string {
	if contains(c.modeRect, x, y) {
		return "chairman"
	}
	for _, key := range c.keys {
		if contains(key.rect, x, y) {
			return key.action
		}
	}
	return ""
}

func contains(rect walk.Rectangle, x, y int) bool {
	return x >= rect.X && x < rect.X+rect.Width && y >= rect.Y && y < rect.Y+rect.Height
}

func (c *Calculator) onMouseMove(x, y int, _ walk.MouseButton) {
	action := c.hitTest(x, y)
	if action != c.hoverAction {
		c.hoverAction = action
		c.canvas.Invalidate()
	}
}

func (c *Calculator) onMouseDown(x, y int, _ walk.MouseButton) {
	c.pressedAction = c.hitTest(x, y)
	c.canvas.Invalidate()
}

func (c *Calculator) onMouseUp(x, y int, _ walk.MouseButton) {
	action := c.hitTest(x, y)
	pressed := c.pressedAction
	c.pressedAction = ""
	if action != "" && action == pressed {
		c.performAction(action)
	}
	c.canvas.Invalidate()
}

func fillRounded(canvas *walk.Canvas, color walk.Color, rect walk.Rectangle, radius int) {
	brush, err := walk.NewSolidColorBrush(color)
	if err != nil {
		return
	}
	defer brush.Dispose()
	if radius > 0 {
		_ = canvas.FillRoundedRectangle(brush, rect, walk.Size{Width: radius * 2, Height: radius * 2})
		return
	}
	_ = canvas.FillRectangle(brush, rect)
}

func drawRoundedBorder(canvas *walk.Canvas, color walk.Color, rect walk.Rectangle, radius int) {
	pen, err := walk.NewCosmeticPen(walk.PenSolid, color)
	if err != nil {
		return
	}
	defer pen.Dispose()
	_ = canvas.DrawRoundedRectangle(pen, rect, walk.Size{Width: radius * 2, Height: radius * 2})
}

func drawLine(canvas *walk.Canvas, color walk.Color, from, to walk.Point) {
	pen, err := walk.NewCosmeticPen(walk.PenSolid, color)
	if err != nil {
		return
	}
	defer pen.Dispose()
	_ = canvas.DrawLine(pen, from, to)
}

func drawText(canvas *walk.Canvas, text string, font *walk.Font, color walk.Color, rect walk.Rectangle, format walk.DrawTextFormat) {
	if font != nil {
		_ = canvas.DrawText(text, font, color, rect, format)
	}
}

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
)

const (
	smCXScreen  = 0
	smCYScreen  = 1
	hwndTop     = 0
	swpNoSize   = 0x0001
	swpNoZOrder = 0x0004
)

// centreWindow 将应用窗口放到主屏幕中央。
func (c *Calculator) centreWindow() {
	size := c.MainWindow.Size()
	screenWidth, _, _ := procGetSystemMetrics.Call(uintptr(smCXScreen))
	screenHeight, _, _ := procGetSystemMetrics.Call(uintptr(smCYScreen))
	x := (int(screenWidth) - size.Width) / 2
	y := (int(screenHeight) - size.Height) / 2
	procSetWindowPos.Call(uintptr(c.MainWindow.Handle()), hwndTop, uintptr(x), uintptr(y), 0, 0, swpNoSize|swpNoZOrder)
}

func logImageError(err error) {
	log.Printf("主席照片加载失败：%v", err)
}
