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

// Calculator is the main application window and logic controller.
type Calculator struct {
	*walk.MainWindow

	num1      string
	num2      string
	operator  string
	result    string
	display   string
	calcCount int
	chairman  bool
	fresh     bool // next digit starts a new number

	displayLabel  *walk.Label
	statsLabel    *walk.Label
	portraitLabel *walk.Label

	topPanel    *walk.Composite
	displayPane *walk.Composite
	statsPane   *walk.Composite
	btnPane     *walk.Composite

	player   *audio.Player
	btnFont  *walk.Font
	btnWidth int
}

// New creates and initialises the Calculator window.
func New() *Calculator {
	return &Calculator{}
}

// Run builds the window and starts the application event loop.
func (c *Calculator) Run() {
	c.player = audio.New()
	c.fresh = true

	const winW, winH = 400, 550
	c.btnWidth = 35
	c.btnFont, _ = walk.NewFont("SimSun", 10, walk.FontBold)

	portraitText := `     ╔══════════════╗
     ║   毛泽东     ║
     ║  (1893-1976) ║
     ║              ║
     ║   🚩 东方红  ║
     ╚══════════════╝`

	if err := (MainWindow{
		AssignTo: &c.MainWindow,
		Title:    "中华人民共和国超级计算器 v1.0",
		MinSize:  Size{winW, winH},
		MaxSize:  Size{winW, winH},
		Size:     Size{winW, winH},
		Layout:   VBox{MarginsZero: true, Spacing: 0},
		Children: []Widget{
			c.buildTopPanel(),
			c.buildDisplayPane(),
			c.buildButtonGrid(),
			c.buildStatsPane(),
			Label{
				AssignTo:   &c.portraitLabel,
				Text:       portraitText,
				Font:       SimSun(10, false),
				TextColor:  ColorGold,
				Background: SolidColorBrush{Color: ColorDarkRed},
				Visible:    false,
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	c.applyBackground(c.MainWindow, ColorDarkRed)
	c.SetBackgroundImage(winW, winH)
	c.centreWindow()
	c.showWelcome()

	c.MainWindow.KeyPress().Attach(func(key walk.Key) {
		c.handleKey(key)
	})
	c.MainWindow.Run()
}

func (c *Calculator) showWelcome() {
	c.displayLabel.SetText("欢迎使用超级计算器")
}

// handleDigit processes digit input (Chinese or Arabic converted).
func (c *Calculator) handleDigit(digit string) {
	if c.fresh {
		c.display = ""
		c.fresh = false
	}
	if digit == "." {
		if strings.Contains(c.display, "点") {
			return
		}
		if c.display == "" {
			c.display = "零"
		}
		c.display += "点"
	} else {
		c.display += digit
	}
	c.displayLabel.SetText(c.display)
}

// handleOperator stores the first operand and operator.
func (c *Calculator) handleOperator(op string) {
	if c.display == "" && c.num1 == "" {
		return
	}
	if c.display != "" {
		c.num1 = c.display
	}
	c.operator = op
	c.fresh = true
	if c.num1 != "" {
		c.displayLabel.SetText(c.num1 + " " + op)
	}
}

// Calculate performs the pending operation and shows the result.
func (c *Calculator) Calculate() {
	if c.operator == "" || c.num1 == "" {
		return
	}
	if c.display != "" {
		c.num2 = c.display
	} else if c.num2 == "" {
		return
	}

	a, err1 := chinese.Parse(c.num1)
	b, err2 := chinese.Parse(c.num2)
	if err1 != nil || err2 != nil {
		c.displayLabel.SetText("错误")
		walk.MsgBox(c.MainWindow, "错误", "无效的数字", walk.MsgBoxIconError)
		return
	}

	var result float64
	switch c.operator {
	case "+":
		result = float64(a + b)
	case "-":
		result = float64(a - b)
	case "*":
		result = float64(a * b)
	case "/":
		if b == 0 {
			c.displayLabel.SetText("错误")
			walk.MsgBox(c.MainWindow, "错误", "除以零错误！", walk.MsgBoxIconError)
			return
		}
		result = float64(a) / float64(b)
	default:
		c.displayLabel.SetText("错误")
		return
	}

	if chinese.IsWhole(result) {
		c.result = chinese.Format(int(result))
	} else {
		c.result = chinese.FormatFloat(result)
	}

	c.display = c.result
	c.displayLabel.SetText(c.display)

	c.calcCount++
	c.statsLabel.SetText(fmt.Sprintf("总计算次数: %d", c.calcCount))

	c.num1 = c.result
	c.num2 = ""
	c.fresh = true
}

// Backspace removes the last character from the display.
func (c *Calculator) Backspace() {
	if c.fresh || c.display == "" {
		return
	}
	runes := []rune(c.display)
	if len(runes) > 0 {
		c.display = string(runes[:len(runes)-1])
	}
	c.displayLabel.SetText(c.display)
}

// Clear resets all state.
func (c *Calculator) Clear() {
	c.num1 = ""
	c.num2 = ""
	c.operator = ""
	c.result = ""
	c.display = ""
	c.fresh = true
	c.displayLabel.SetText("")
}

// ToggleChairman switches between normal and chairman modes.
func (c *Calculator) ToggleChairman() {
	c.chairman = !c.chairman
	if c.chairman {
		c.setBackgroundAll(ColorRed)
		c.portraitLabel.SetVisible(true)
		c.displayLabel.SetText("毛主席万岁！")
		if c.player != nil {
			c.player.Play()
		}
	} else {
		c.setBackgroundAll(ColorDarkRed)
		c.portraitLabel.SetVisible(false)
		c.displayLabel.SetText("欢迎使用")
		if c.player != nil {
			c.player.Stop()
		}
	}
}

// handleKey processes keyboard input.
func (c *Calculator) handleKey(key walk.Key) {
	k := int(key)

	if k >= '0' && k <= '9' {
		c.handleDigit(string(chinese.ChineseDigits[k-'0']))
		return
	}
	for i, r := range chinese.ChineseDigits {
		if int(r) == k {
			c.handleDigit(string(chinese.ChineseDigits[i]))
			return
		}
	}

	switch {
	case key == walk.KeyReturn:
		c.Calculate()
	case key == walk.KeyEscape:
		c.Clear()
	case key == walk.KeyBack:
		c.Backspace()
	case k == '+':
		c.handleOperator("+")
	case k == '-':
		c.handleOperator("-")
	case k == '*':
		c.handleOperator("*")
	case k == '/':
		c.handleOperator("/")
	case k == '.':
		c.handleDigit(".")
	}
}
