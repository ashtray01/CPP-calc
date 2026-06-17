package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"
)

//go:embed leader.png
var 主席图像数据 []byte

var (
	金色 = walk.RGB(0xFF, 0xD7, 0x00)
	暗红 = walk.RGB(0x8B, 0x00, 0x00)
	深红 = walk.RGB(0x4A, 0x00, 0x00)
	红色 = walk.RGB(0xFF, 0x00, 0x00)
)

var 中数字_到_阿拉伯 = map[rune]int{
	'零': 0, '一': 1, '二': 2, '三': 3, '四': 4,
	'五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
}

var 阿拉伯_到_中数字 = map[int]rune{
	0: '零', 1: '一', 2: '二', 3: '三', 4: '四',
	5: '五', 6: '六', 7: '七', 8: '八', 9: '九',
}

var 中文数字列表 = []rune{'零', '一', '二', '三', '四', '五', '六', '七', '八', '九'}

func 中文转阿拉伯(中文文本 string) (int, error) {
	结果 := 0
	for _, 字符 := range 中文文本 {
		if v, ok := 中数字_到_阿拉伯[字符]; ok {
			结果 = 结果*10 + v
		} else {
			return 0, fmt.Errorf("无效的中文数字字符: %c", 字符)
		}
	}
	return 结果, nil
}

func 阿拉伯转中文(数字 int) string {
	if 数字 == 0 {
		return "零"
	}
	符号 := ""
	if 数字 < 0 {
		符号 = "负"
		数字 = -数字
	}
	var 结果 strings.Builder
	for _, 字符 := range strconv.Itoa(数字) {
		结果.WriteRune(阿拉伯_到_中数字[int(字符-'0')])
	}
	return 符号 + 结果.String()
}

func 阿拉伯浮点转中文(数字 float64) string {
	if 数字 == 0 {
		return "零"
	}
	if 数字 == math.Trunc(数字) {
		return 阿拉伯转中文(int(数字))
	}
	符号 := ""
	if 数字 < 0 {
		符号 = "负"
		数字 = -数字
	}
	整数部分 := int(math.Floor(数字))
	小数部分 := 数字 - math.Floor(数字)
	结果 := 阿拉伯转中文(整数部分) + "点"
	小数文本 := fmt.Sprintf("%.10f", 小数部分)
	小数文本 = strings.TrimRight(小数文本, "0")
	小数文本 = strings.TrimLeft(小数文本, "0.")
	for _, 字符 := range 小数文本 {
		结果 += string(阿拉伯_到_中数字[int(字符-'0')])
	}
	return 符号 + 结果
}

var (
	winmm         = windows.NewLazySystemDLL("winmm.dll")
	mciSendString = winmm.NewProc("mciSendStringW")
)

type 音频 struct {
	路径 string
}

func new音频() *音频 {
	return &音频{路径: 找到MP3()}
}

func 找到MP3() string {
	seen := map[string]bool{}
	dirs := []string{}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		dirs = append(dirs, exeDir)
		dirs = append(dirs, filepath.Dir(exeDir)) // parent dir
	}
	for _, dir := range dirs {
		abs, _ := filepath.Abs(dir)
		if seen[abs] {
			continue
		}
		seen[abs] = true
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".mp3") {
				return filepath.Join(abs, e.Name())
			}
		}
	}
	return ""
}

func mci执行(命令 string) {
	cmdPtr, _ := syscall.UTF16PtrFromString(命令)
	mciSendString.Call(uintptr(unsafe.Pointer(cmdPtr)), 0, 0, 0)
}

func (a *音频) 播放() {
	if a.路径 == "" {
		return
	}
	a.停止()
	短路径 := strings.ReplaceAll(a.路径, "/", "\\")
	mci执行(fmt.Sprintf(`open "%s" type mpegvideo alias 音乐`, 短路径))
	mci执行("play 音乐 repeat")
}

func (a *音频) 停止() {
	mci执行("stop 音乐")
	mci执行("close 音乐")
}

type 计算器 struct {
	*walk.MainWindow

	数字一    string
	数字二    string
	运算符    string
	结果     string
	显示文本   string
	总计算次数  int
	主席模式激活 bool
	输入开始   bool

	显示Label   *walk.Label
	统计Label   *walk.Label
	主席画像Label *walk.Label

	顶部面板 *walk.Composite
	显示面板 *walk.Composite
	统计面板 *walk.Composite
	按钮面板 *walk.Composite

	音频     *音频
	按钮字体 *walk.Font

	btnWidth int
}

func simsun(size int, bold bool) Font {
	return Font{Family: "SimSun", PointSize: size, Bold: bold}
}

func (c *计算器) 初始化() {
	c.音频 = new音频()
	c.输入开始 = true

	winW, winH := 400, 550
	c.btnWidth = 35
	c.按钮字体, _ = walk.NewFont("SimSun", 10, walk.FontBold)
	bgBitmap, _ := c.loadLeaderImage()

	肖像文本 := `     ╔══════════════╗
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
			Composite{
				AssignTo:   &c.顶部面板,
				Background: SolidColorBrush{Color: 暗红},
				Layout:     VBox{Alignment: AlignHCenterVNear, Margins: Margins{0, 4, 0, 1}},
				Children: []Widget{
					Label{Text: "★ ★ ★ ★ ★", Font: simsun(10, false), TextColor: 金色, Background: SolidColorBrush{Color: 暗红}},
					Label{Text: "中华人民共和国超级计算器", Font: simsun(12, true), TextColor: 金色, Background: SolidColorBrush{Color: 暗红}},
					Label{Text: "v1.0 | 代号: 红龙 🐉", Font: simsun(9, false), TextColor: 金色, Background: SolidColorBrush{Color: 暗红}},
				},
			},
			Composite{
				AssignTo:   &c.显示面板,
				Background: SolidColorBrush{Color: 深红},
				Layout:     VBox{Margins: Margins{15, 3, 15, 3}},
				Children: []Widget{
					Label{
						AssignTo:      &c.显示Label,
						Font:          simsun(18, true),
						TextColor:     金色,
						Background:    SolidColorBrush{Color: 深红},
						MinSize:       Size{0, 36},
						TextAlignment: AlignFar,
					},
				},
			},
			Composite{
				AssignTo: &c.按钮面板,
				Layout:   VBox{Margins: Margins{15, 0, 15, 0}, Spacing: 3},
				Children: []Widget{
					c.创建按钮行("七", "八", "九", "÷"),
					c.创建按钮行("四", "五", "六", "×"),
					c.创建按钮行("一", "二", "三", "−"),
					c.创建按钮行("零", ".", "＝", "＋"),
					c.创建最后行(),
				},
			},
			Composite{
				AssignTo:   &c.统计面板,
				Background: SolidColorBrush{Color: 暗红},
				Layout:     HBox{Margins: Margins{15, 2, 15, 4}},
				Children: []Widget{
					Label{
						AssignTo:   &c.统计Label,
						Font:       simsun(9, false),
						TextColor:  金色,
						Background: SolidColorBrush{Color: 暗红},
						Text:       "总计算次数: 0",
					},
				},
			},
			Label{
				AssignTo:   &c.主席画像Label,
				Text:       肖像文本,
				Font:       simsun(10, false),
				TextColor:  金色,
				Background: SolidColorBrush{Color: 暗红},
				Visible:    false,
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	// Use pre-loaded background image, resized to match window at system DPI
	if bgBitmap != nil {
		dpi := c.MainWindow.DPI()
		// Physical pixel size for the client area background
		physW := winW * dpi / 96
		physH := winH * dpi / 96
		if resized, err := walk.NewBitmap(walk.Size{Width: physW, Height: physH}); err == nil {
			if canvas, err := walk.NewCanvasFromImage(resized); err == nil {
				canvas.DrawImageStretchedPixels(bgBitmap, walk.Rectangle{Width: physW, Height: physH})
				canvas.Dispose()
				if brush, err := walk.NewBitmapBrush(resized); err == nil {
					c.MainWindow.SetBackground(brush)
				}
			}
		}
	}

	c.центрироватьОкно()

	c.播放欢迎()
	c.MainWindow.KeyPress().Attach(func(key walk.Key) {
		c.键盘处理(key)
	})
}

func (c *计算器) loadLeaderImage() (*walk.Bitmap, error) {
	img, _, err := image.Decode(bytes.NewReader(主席图像数据))
	if err != nil {
		return nil, err
	}
	return walk.NewBitmapFromImage(img)
}

func (c *计算器) 创建按钮(文本 string, 大小 Size, 点击 func()) CustomWidget {
	return CustomWidget{
		MinSize:       大小,
		MaxSize:       大小,
		StretchFactor: 0,
		PaintMode:     PaintNoErase,
		Paint: func(canvas *walk.Canvas, bounds walk.Rectangle) error {
			return canvas.DrawText(文本, c.按钮字体, 金色, bounds, walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
		},
		OnMouseDown: func(x, y int, button walk.MouseButton) {
			点击()
		},
	}
}

func (c *计算器) 创建按钮行(文本列表 ...string) Composite {
	var children []Widget
	for _, t := range 文本列表 {
		bt := t
		children = append(children, c.创建按钮(bt, Size{c.btnWidth, 28}, func() { c.处理按钮点击(bt) }))
	}
	return Composite{
		Layout:   HBox{Spacing: 0, MarginsZero: true},
		Children: children,
	}
}

func (c *计算器) 创建最后行() Composite {
	b := c.btnWidth
	return Composite{
		Layout: HBox{Spacing: 0, MarginsZero: true},
		Children: []Widget{
			c.创建按钮("清除", Size{b*2 + 3, 28}, c.执行清除),
			c.创建按钮("←", Size{b, 28}, c.删除),
			c.创建按钮("主席", Size{b, 28}, c.切换主席模式),
		},
	}
}

func (c *计算器) 处理按钮点击(文本 string) {
	switch 文本 {
	case "÷":
		c.操作输入("/")
	case "×":
		c.操作输入("*")
	case "−":
		c.操作输入("-")
	case "＋":
		c.操作输入("+")
	case "＝":
		c.执行计算()
	case "清除":
		c.执行清除()
	case "←":
		c.删除()
	case "主席":
		c.切换主席模式()
	default:
		c.数字输入(文本)
	}
}

func (c *计算器) 数字输入(数字 string) {
	if c.输入开始 {
		c.显示文本 = ""
		c.输入开始 = false
	}
	if 数字 == "." || 数字 == "点" {
		if strings.Contains(c.显示文本, "点") {
			return
		}
		if c.显示文本 == "" {
			c.显示文本 = "零"
		}
		c.显示文本 += "点"
	} else {
		c.显示文本 += 数字
	}
	c.显示Label.SetText(c.显示文本)
}

func (c *计算器) 操作输入(操作 string) {
	if c.显示文本 == "" && c.数字一 == "" {
		return
	}
	if c.显示文本 != "" {
		c.数字一 = c.显示文本
	}
	c.运算符 = 操作
	c.输入开始 = true
	if c.数字一 != "" {
		c.显示Label.SetText(c.数字一 + " " + 操作)
	}
}

func (c *计算器) 执行计算() {
	if c.运算符 == "" || c.数字一 == "" {
		return
	}
	if c.显示文本 != "" {
		c.数字二 = c.显示文本
	} else if c.数字二 == "" {
		return
	}

	数一, err1 := 中文转阿拉伯(c.数字一)
	数二, err2 := 中文转阿拉伯(c.数字二)
	if err1 != nil || err2 != nil {
		c.显示Label.SetText("错误")
		walk.MsgBox(c.MainWindow, "错误", "无效的数字", walk.MsgBoxIconError)
		return
	}

	var 运算结果 float64
	switch c.运算符 {
	case "+":
		运算结果 = float64(数一 + 数二)
	case "-":
		运算结果 = float64(数一 - 数二)
	case "*":
		运算结果 = float64(数一 * 数二)
	case "/":
		if 数二 == 0 {
			c.显示Label.SetText("错误")
			walk.MsgBox(c.MainWindow, "错误", "除以零错误！", walk.MsgBoxIconError)
			return
		}
		运算结果 = float64(数一) / float64(数二)
	default:
		c.显示Label.SetText("错误")
		return
	}

	if 运算结果 == float64(int(运算结果)) {
		c.结果 = 阿拉伯转中文(int(运算结果))
	} else {
		c.结果 = 阿拉伯浮点转中文(运算结果)
	}
	c.显示文本 = c.结果
	c.显示Label.SetText(c.显示文本)

	c.总计算次数++
	c.统计Label.SetText(fmt.Sprintf("总计算次数: %d", c.总计算次数))

	c.数字一 = c.结果
	c.数字二 = ""
	c.输入开始 = true
}

func (c *计算器) 删除() {
	if c.输入开始 || c.显示文本 == "" {
		return
	}
	runes := []rune(c.显示文本)
	if len(runes) > 0 {
		c.显示文本 = string(runes[:len(runes)-1])
	}
	c.显示Label.SetText(c.显示文本)
}

func (c *计算器) 执行清除() {
	c.数字一 = ""
	c.数字二 = ""
	c.运算符 = ""
	c.结果 = ""
	c.显示文本 = ""
	c.输入开始 = true
	c.显示Label.SetText("")
}

func (c *计算器) 切换主席模式() {
	c.主席模式激活 = !c.主席模式激活
	if c.主席模式激活 {
		c.设置全部背景(红色)
		c.主席画像Label.SetVisible(true)
		c.显示Label.SetText("毛主席万岁！")
		c.音频.播放()
	} else {
		c.设置全部背景(暗红)
		c.主席画像Label.SetVisible(false)
		c.显示Label.SetText("欢迎使用")
		c.音频.停止()
	}
}

func (c *计算器) 设置全部背景(col walk.Color) {
	brush, err := walk.NewSolidColorBrush(col)
	if err != nil {
		return
	}
	c.顶部面板.SetBackground(brush)
	c.显示面板.SetBackground(brush)
	c.统计面板.SetBackground(brush)
	c.按钮面板.SetBackground(brush)
	c.主席画像Label.SetBackground(brush)
	// Update labels inside top panel
	for i := 0; i < c.顶部面板.Children().Len(); i++ {
		w := c.顶部面板.Children().At(i)
		if l, ok := w.(*walk.Label); ok {
			l.SetBackground(brush)
		}
	}
	// Update labels inside stat panel
	for i := 0; i < c.统计面板.Children().Len(); i++ {
		w := c.统计面板.Children().At(i)
		if l, ok := w.(*walk.Label); ok {
			l.SetBackground(brush)
		}
	}
	c.显示Label.SetBackground(brush)
}

func (c *计算器) 播放欢迎() {
	c.显示Label.SetText("欢迎使用超级计算器")
}

func (c *计算器) 键盘处理(key walk.Key) {
	k := int(key)

	if k >= '0' && k <= '9' {
		c.数字输入(string(中文数字列表[k-'0']))
		return
	}

	for i, r := range 中文数字列表 {
		if int(r) == k {
			c.数字输入(string(中文数字列表[i]))
			return
		}
	}

	switch {
	case key == walk.KeyReturn:
		c.执行计算()
	case key == walk.KeyEscape:
		c.执行清除()
	case key == walk.KeyBack:
		c.删除()
	case k == '+':
		c.操作输入("+")
	case k == '-':
		c.操作输入("-")
	case k == '*':
		c.操作输入("*")
	case k == '/':
		c.操作输入("/")
	case k == '.':
		c.数字输入(".")
	}
}

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
)

const (
	SM_CXSCREEN  = 0
	SM_CYSCREEN  = 1
	HWND_TOP     = 0
	SWP_NOSIZE   = 0x0001
	SWP_NOZORDER = 0x0004
)

func (c *计算器) центрироватьОкно() {
	sz := c.MainWindow.Size()
	screenW, _, _ := procGetSystemMetrics.Call(uintptr(SM_CXSCREEN))
	screenH, _, _ := procGetSystemMetrics.Call(uintptr(SM_CYSCREEN))
	x := (int(screenW) - sz.Width) / 2
	y := (int(screenH) - sz.Height) / 2
	procSetWindowPos.Call(
		uintptr(c.MainWindow.Handle()),
		HWND_TOP,
		uintptr(x), uintptr(y),
		0, 0,
		SWP_NOSIZE|SWP_NOZORDER,
	)
}

func main() {
	// Отключаем DPI-масштабирование
	user32 := windows.NewLazySystemDLL("user32.dll")
	user32.NewProc("SetProcessDPIAware").Call()

	calc := new(计算器)
	calc.初始化()
	calc.Run()
}
