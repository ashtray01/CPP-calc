// Package ui 负责计算器窗口、交互和视觉呈现。
package ui

import (
	"fmt"
	"log"
	"strings"

	"calculator/internal/audio"
	"calculator/internal/chinese"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Calculator 保存窗口状态、计算状态和界面资源。
type Calculator struct {
	*walk.MainWindow

	num1      string
	num2      string
	operator  string
	result    string
	display   string
	history   string
	calcCount int
	chairman  bool
	fresh     bool

	canvas        *walk.CustomWidget
	leader        *walk.Bitmap
	keys          []calculatorKey
	modeRect      walk.Rectangle
	hoverAction   string
	pressedAction string
	player        *audio.Player
	fontTitle     *walk.Font
	fontCardTitle *walk.Font
	fontDisplay   *walk.Font
	fontButton    *walk.Font
	fontSmall     *walk.Font
	fontSmallBold *walk.Font
	fontTiny      *walk.Font
}

// New 创建尚未运行的计算器。
func New() *Calculator {
	return &Calculator{fresh: true, display: "零"}
}

// Run 创建窗口并进入 Windows 消息循环。
func (c *Calculator) Run() {
	c.player = audio.New()
	defer c.dispose()

	c.fontTitle, _ = walk.NewFont("Microsoft YaHei UI", 20, walk.FontBold)
	c.fontCardTitle, _ = walk.NewFont("Microsoft YaHei UI", 14, walk.FontBold)
	c.fontDisplay, _ = walk.NewFont("Microsoft YaHei UI", 28, walk.FontBold)
	c.fontButton, _ = walk.NewFont("Microsoft YaHei UI", 14, walk.FontBold)
	c.fontSmall, _ = walk.NewFont("Microsoft YaHei UI", 10, 0)
	c.fontSmallBold, _ = walk.NewFont("Microsoft YaHei UI", 10, walk.FontBold)
	c.fontTiny, _ = walk.NewFont("Microsoft YaHei UI", 9, 0)

	var err error
	c.leader, err = c.loadLeaderImage()
	if err != nil {
		logImageError(err)
	}

	const windowWidth, windowHeight = 980, 750
	if err := (MainWindow{
		AssignTo:   &c.MainWindow,
		Title:      "中华人民共和国 · 中文计算器",
		MinSize:    Size{Width: windowWidth, Height: windowHeight},
		MaxSize:    Size{Width: windowWidth, Height: windowHeight},
		Size:       Size{Width: windowWidth, Height: windowHeight},
		Background: SolidColorBrush{Color: ColorCanvas},
		Layout:     VBox{MarginsZero: true, Spacing: 0},
		Children:   []Widget{c.buildCanvas()},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	c.centreWindow()
	c.MainWindow.KeyPress().Attach(c.handleKey)
	c.canvas.KeyPress().Attach(c.handleKey)
	c.MainWindow.Run()
}

func (c *Calculator) dispose() {
	if c.player != nil {
		c.player.Cleanup()
	}
	if c.leader != nil {
		c.leader.Dispose()
	}
	fonts := []*walk.Font{c.fontTitle, c.fontCardTitle, c.fontDisplay, c.fontButton, c.fontSmall, c.fontSmallBold, c.fontTiny}
	for _, font := range fonts {
		if font != nil {
			font.Dispose()
		}
	}
}

func (c *Calculator) displayText() string {
	if c.display == "" {
		return "零"
	}
	return c.display
}

func (c *Calculator) historyText() string {
	if c.history != "" {
		return c.history
	}
	if c.operator != "" && c.num1 != "" {
		return c.num1 + " " + operatorLabel(c.operator)
	}
	return "准备计算"
}

func (c *Calculator) countText() string {
	return fmt.Sprintf("已完成 %d 次计算", c.calcCount)
}

func operatorLabel(operator string) string {
	switch operator {
	case "*":
		return "×"
	case "/":
		return "÷"
	case "-":
		return "−"
	case "+":
		return "＋"
	default:
		return operator
	}
}

func (c *Calculator) performAction(action string) {
	switch action {
	case "clear":
		c.Clear()
	case "sign":
		c.ToggleSign()
	case "backspace":
		c.Backspace()
	case "percent":
		c.Percentage()
	case "equals":
		c.Calculate()
	case "chairman":
		c.ToggleChairman()
	case "+", "-", "*", "/":
		c.handleOperator(action)
	default:
		c.handleDigit(action)
	}
}

// handleDigit 处理中文数字和小数点输入。
func (c *Calculator) handleDigit(digit string) {
	if c.fresh {
		c.display = ""
		c.fresh = false
	}
	if digit == "." {
		if strings.Contains(c.display, "点") {
			return
		}
		if c.display == "" || c.display == "负" {
			c.display += "零"
		}
		c.display += "点"
	} else {
		if c.display == "零" {
			c.display = ""
		}
		c.display += digit
	}
	c.refresh()
}

// handleOperator 保存第一个操作数和运算符。
func (c *Calculator) handleOperator(operator string) {
	if c.display == "" {
		return
	}
	if c.operator != "" && !c.fresh {
		c.Calculate()
	}
	c.num1 = c.displayText()
	c.operator = operator
	c.history = ""
	c.fresh = true
	c.refresh()
}

// Calculate 执行当前的二元运算。
func (c *Calculator) Calculate() {
	if c.operator == "" || c.num1 == "" {
		return
	}
	if !c.fresh {
		c.num2 = c.display
	} else if c.num2 == "" {
		return
	}

	a, errA := chinese.ParseFloat(c.num1)
	b, errB := chinese.ParseFloat(c.num2)
	if errA != nil || errB != nil {
		c.showError("输入中包含无效数字")
		return
	}

	var value float64
	switch c.operator {
	case "+":
		value = a + b
	case "-":
		value = a - b
	case "*":
		value = a * b
	case "/":
		if b == 0 {
			c.showError("除数不能为零")
			return
		}
		value = a / b
	default:
		return
	}

	c.result = chinese.FormatFloat(value)
	c.history = c.num1 + " " + operatorLabel(c.operator) + " " + c.num2 + " ＝"
	c.display = c.result
	c.calcCount++
	c.num1 = c.result
	c.fresh = true
	c.refresh()
}

func (c *Calculator) showError(message string) {
	c.display = "错误"
	c.fresh = true
	c.refresh()
	walk.MsgBox(c.MainWindow, "计算错误", message, walk.MsgBoxIconError)
}

// Backspace 删除当前输入的最后一个字符。
func (c *Calculator) Backspace() {
	if c.fresh || c.display == "" {
		return
	}
	runes := []rune(c.display)
	if len(runes) > 0 {
		c.display = string(runes[:len(runes)-1])
	}
	if c.display == "" || c.display == "负" {
		c.display = "零"
		c.fresh = true
	}
	c.refresh()
}

// Clear 重置当前运算，但保留累计次数和音乐模式。
func (c *Calculator) Clear() {
	c.num1 = ""
	c.num2 = ""
	c.operator = ""
	c.result = ""
	c.display = "零"
	c.history = ""
	c.fresh = true
	c.refresh()
}

// ToggleSign 切换当前数字的正负号。
func (c *Calculator) ToggleSign() {
	if c.display == "" || c.display == "零" || c.display == "错误" {
		return
	}
	if strings.HasPrefix(c.display, "负") {
		c.display = strings.TrimPrefix(c.display, "负")
	} else {
		c.display = "负" + c.display
	}
	c.fresh = false
	c.refresh()
}

// Percentage 将当前数字转换为百分比。
func (c *Calculator) Percentage() {
	value, err := chinese.ParseFloat(c.display)
	if err != nil {
		return
	}
	c.display = chinese.FormatFloat(value / 100)
	c.fresh = true
	c.refresh()
}

// ToggleChairman 控制主席模式和原有背景音乐。
func (c *Calculator) ToggleChairman() {
	c.chairman = !c.chairman
	if c.player != nil {
		if c.chairman {
			c.player.Play()
		} else {
			c.player.Stop()
		}
	}
	c.refresh()
}

func (c *Calculator) refresh() {
	if c.canvas != nil {
		c.canvas.Invalidate()
	}
}

// handleKey 提供数字键、运算键和常用快捷键支持。
func (c *Calculator) handleKey(key walk.Key) {
	value := int(key)
	if key == walk.Key8 && walk.ModifiersDown()&walk.ModShift != 0 {
		c.handleOperator("*")
		return
	}
	if value >= '0' && value <= '9' {
		c.handleDigit(string(chinese.ChineseDigits[value-'0']))
		return
	}
	if key >= walk.KeyNumpad0 && key <= walk.KeyNumpad9 {
		c.handleDigit(string(chinese.ChineseDigits[int(key-walk.KeyNumpad0)]))
		return
	}
	switch {
	case key == walk.KeyReturn:
		c.Calculate()
	case key == walk.KeyEscape:
		c.Clear()
	case key == walk.KeyBack:
		c.Backspace()
	case key == walk.KeyAdd || value == 0xBB:
		c.handleOperator("+")
	case key == walk.KeySubtract || value == 0xBD:
		c.handleOperator("-")
	case key == walk.KeyMultiply:
		c.handleOperator("*")
	case key == walk.KeyDivide || value == 0xBF:
		c.handleOperator("/")
	case key == walk.KeyDecimal || value == 0xBE:
		c.handleDigit(".")
	}
}
