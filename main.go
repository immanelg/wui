package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell"
)

var screen tcell.Screen

func initScreen() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err = screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	screen.EnableMouse()
	screen.Clear()
}

// Rectangle on screen with absolute coordinates
// x from left to right, y from top to bottom
// inclusive
type Rect struct{ x, y, x1, y1 int }

func (r Rect) Values() (int, int, int, int) { return r.x, r.y, r.x1, r.y1 }

type Widget interface {
	Render()
	Resize(rect Rect)
	GetRect() Rect
}

type TextWidget struct {
	rect Rect
	text string
}

func (w *TextWidget) Render() {
	x, y, x1, y1 := w.rect.Values()
	for i := x; i <= x1; i++ {
		for j := y; j <= y1; j++ {
			// text is wrapped
			var r rune = ' '
			if idx := (j-y)*(x1-x) + (i - x); idx < len(w.text) {
				r = rune(w.text[idx])
			}
			screen.SetContent(i, j, r, nil, tcell.StyleDefault)
		}
	}
}
func (w *TextWidget) Resize(rect Rect) { w.rect = rect }

func (w *TextWidget) GetRect() Rect { return w.rect }

type ListWidget struct {
	rect     Rect
	lines    []string
	offset   int
	selected int
}

func (w *ListWidget) Render() {
	x, y, x1, y1 := w.rect.Values()

	lineCount := len(w.lines)
	for j := y; j <= y1 && j-y < lineCount; j++ {
		lineIdx := w.offset + j - y
		if lineIdx >= len(w.lines) {
			continue
		}
		line := w.lines[lineIdx]
		lineLen := len(line)
		st := tcell.StyleDefault
		if lineIdx == w.selected {
			st = tcell.StyleDefault.Underline(true)
		}
		for i := x; i <= x1 && i-x < lineLen; i++ {
			r := rune(line[i-x])
			screen.SetContent(i, j, r, nil, st)
		}
	}
}

func (w *ListWidget) Down() {
	w.selected = min(len(w.lines)-1, w.selected+1)
	if w.selected > w.offset+w.rect.y1-w.rect.y {
		w.offset++
	}
}

func (w *ListWidget) Up() {
	w.selected = max(0, w.selected-1)
	if w.selected < w.offset {
		w.offset--
	}
}

func (w *ListWidget) Last() {
	w.selected = max(len(w.lines)-1, 0)
	w.offset = max(0, len(w.lines)-(w.rect.y1-w.rect.y))
}

func (w *ListWidget) First() {
	w.selected = 0
	w.offset = 0
}

func (w *ListWidget) Resize(rect Rect) {
	w.rect = rect
}

func (w *ListWidget) GetRect() Rect { return w.rect }

type BorderedWidget struct {
	rect  Rect
	inner Widget
	title string
}

func (self *BorderedWidget) Render() {
	self.inner.Render()

	x, y, x1, y1 := self.rect.Values()

	style := tcell.StyleDefault
	for i := x + 1; i < x1; i++ {
		screen.SetContent(i, y, tcell.RuneHLine, nil, style)
		screen.SetContent(i, y1, tcell.RuneHLine, nil, style)
	}
	for i := x + 1; i < x1 && i-(x+1) < len(self.title); i++ {
		screen.SetContent(i, y, rune(self.title[i-(x+1)]), nil, style)
	}
	for j := y + 1; j < y1; j++ {
		screen.SetContent(x, j, tcell.RuneVLine, nil, style)
		screen.SetContent(x1, j, tcell.RuneVLine, nil, style)
	}

	screen.SetContent(x, y, tcell.RuneULCorner, nil, style)
	screen.SetContent(x1, y, tcell.RuneURCorner, nil, style)
	screen.SetContent(x, y1, tcell.RuneLLCorner, nil, style)
	screen.SetContent(x1, y1, tcell.RuneLRCorner, nil, style)
}

func (w *BorderedWidget) Resize(rect Rect) {
	w.rect = rect
	innerRect := Rect{x: rect.x + 1, y: rect.y + 1, x1: rect.x1 - 1, y1: rect.y1 - 1}
	w.inner.Resize(innerRect)
}

func (w *BorderedWidget) GetRect() Rect {
	return w.rect
}

type SplitWidget struct {
	rect        Rect
	left, right Widget
	ratio       int
	horizontal  bool
}

func (w *SplitWidget) Render() {
	w.left.Render()
	w.right.Render()
}

func (w *SplitWidget) Resize(rect Rect) {
	w.rect = rect

	if w.horizontal {
		yCenter := rect.y + (rect.y1-rect.y)*w.ratio/100

		w.left.Resize(Rect{x: rect.x, y: rect.y, x1: rect.x1, y1: yCenter})
		w.right.Resize(Rect{x: rect.x, y: min(yCenter+1, rect.y1), x1: rect.x1, y1: rect.y1})
	} else {
		xCenter := rect.x + (rect.x1-rect.x)*w.ratio/100

		w.left.Resize(Rect{x: rect.x, y: rect.y, x1: xCenter, y1: rect.y1})
		w.right.Resize(Rect{x: min(xCenter+1, rect.x1), y: rect.y, x1: rect.x1, y1: rect.y1})
	}
}

func (w *SplitWidget) GetRect() Rect {
	return w.rect
}

type Compositor struct {
	rect            Rect
	widgets         []Widget
	focusedWidgetId int
}

func (c *Compositor) Start() {
	textWidget := TextWidget{text: "abcdefghiklmnopqrstuvwxyzw"}
	textWidgetBordered := BorderedWidget{inner: &textWidget}

	listWidget := ListWidget{
		lines:    []string{"00000000", "111111111", "222222222", "333333333333333", "4444", "55555", "666666666", "777777777777", "888888888888", "999999999", "aaaa", "bbbbbbb", "cccccc", "dddddd"},
		selected: 2,
	}
	listWidgetBordered := BorderedWidget{inner: &listWidget}

	logs := make(chan string, 1024)
	go func() {
		c := 0
		for {
			time.Sleep(1500 * time.Millisecond)
			if c%2 == 0 {
				logs <- fmt.Sprintf("INFO %d%d%d", c, c, c)
			} else {
				logs <- fmt.Sprintf("WARN %d%d%d", c, c, c)
			}
			c++
		}
	}()

	textWidget2 := TextWidget{text: "!@#$_+)+_)+_+_((*()&(*&(*(*()*()_)#%$%$$%^$^%$$##@#######%$_%^&*()_+{}:'>?()*()#&(!&(*&!!$&*<?"}
	textwidget2Bordered := BorderedWidget{inner: &textWidget2, title: "title"}

	left := TextWidget{text: "LEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFT"}
	right := TextWidget{text: "RIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHRIGHTRIGHTRIGHTRIGHTRIGH"}
	splittingWidget := SplitWidget{left: &left, right: &right, horizontal: false, ratio: 25}

	c.widgets = []Widget{
		&listWidgetBordered,
		&textWidgetBordered,
		&textwidget2Bordered,
		&splittingWidget,
	}

	resize := func() {
		textWidgetBordered.Resize(Rect{x: 0, y: 0, x1: 40, y1: 10})
		listWidgetBordered.Resize(Rect{x: 41, y: 0, x1: c.rect.x1, y1: 20})
		textwidget2Bordered.Resize(Rect{x: 0, y: 21, x1: c.rect.x1, y1: 25})
		splittingWidget.Resize(Rect{x: 0, y: 26, x1: c.rect.x1, y1: c.rect.y1})
	}

	width, height := screen.Size()
	c.rect = Rect{x1: width - 1, y1: height - 1}
	resize()

	terminalEventsChan := make(chan tcell.Event, 16)
	go func() {
		for {
			terminalEventsChan <- screen.PollEvent()
		}
	}()

	for {
		screen.Fill(' ', tcell.StyleDefault)

		c.Render()

		screen.Show()

		select {
		case ev := <-terminalEventsChan:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				width, height := ev.Size()
				c.rect = Rect{x1: width - 1, y1: height - 1}
				resize()
				screen.Sync()
			case *tcell.EventKey:
				switch {
				case ev.Rune() == 'j':
					listWidget.Down()
				case ev.Rune() == 'k':
					listWidget.Up()
				case ev.Rune() == 'g':
					listWidget.First()
				case ev.Rune() == 'G':
					listWidget.Last()
				case ev.Rune() == 'q' || ev.Key() == tcell.KeyCtrlC:
					return
				}
			}
		case l := <-logs:
			listWidget.lines = append(listWidget.lines, l)
			listWidget.Last()
			c.Render()
		}
	}
}

func (c *Compositor) Render() {
	for _, w := range c.widgets {
		w.Render()
	}
}

func run() {
	initScreen()
	cleanup := func() {
		err := recover()
		screen.Fini()
		if err != nil {
			panic(err)
		}
	}
	defer cleanup()
	c := Compositor{}
	c.Start()
}

func main() {
	run()
}
