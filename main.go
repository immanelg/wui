package main

import (
	"log"

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
func (w *TextWidget) Resize(rect Rect) {
	w.rect = rect
}

func (w *TextWidget) GetRect() Rect {
	return w.rect
}

type ListWidget struct {
	rect   Rect
	lines  []string
	offset int
}

func (w *ListWidget) Render() {
	x, y, x1, y1 := w.rect.Values()

	lineCount := len(w.lines)
	for j := y; j <= y1 && j-y < lineCount; j++ {
		lineLen := len(w.lines[j-y])
		for i := x; i <= x1 && (i-x) < lineLen; i++ {
			screen.SetContent(i, j, rune(w.lines[j-y][i-x]), nil, tcell.StyleDefault)
		}
	}
}

func (w *ListWidget) Resize(rect Rect) {
	w.rect = rect
}

func (w *ListWidget) GetRect() Rect {
	return w.rect
}

type BorderedWidget struct {
	rect  Rect
	inner Widget
}

func (self *BorderedWidget) Render() {
	self.inner.Render()

	x, y, x1, y1 := self.rect.Values()

	style := tcell.StyleDefault
	for i := x; i < x1; i++ {
		screen.SetContent(i, y, tcell.RuneHLine, nil, style)
		screen.SetContent(i, y1, tcell.RuneHLine, nil, style)
	}
	for j := y; j < y1; j++ {
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
	innerRect := Rect{x: rect.x + 1, y: rect.y + 1, x1: rect.x1 - 2, y1: rect.y1 - 2}
	w.inner.Resize(innerRect)
}

func (w *BorderedWidget) GetRect() Rect {
	return w.rect
}

type Compositor struct {
	rect        Rect
	widgets         []Widget
	focusedWidgetId int
}

func (c *Compositor) Render() {
	for _, w := range c.widgets {
		w.Render()
	}
}

func (c *Compositor) HandleKey(key *tcell.EventKey) {
}

func (c *Compositor) Resize(rect Rect) {
}

type SplitWidget struct {
    rect Rect
    left, right Widget
}

func (w *SplitWidget) Render() {
    w.left.Render()
    w.right.Render()
}

func (w *SplitWidget) Resize(rect Rect) {
    w.rect = rect

    xCenter := (rect.x + rect.x1) / 2

    leftRect := Rect{x: rect.x, y: rect.y, x1: xCenter, y1: rect.y1}
    w.left.Resize(leftRect)

    rightRect := Rect{x: min(xCenter+1, rect.x1), y: rect.y, x1: rect.x1, y1: rect.y1}
    w.right.Resize(rightRect)
}

func (w *SplitWidget) GetRect() Rect {
    return w.rect
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

	termW, termH := screen.Size()
	c := Compositor{rect: Rect{x1: termW-1, y1: termH-1}}

	textWidget := TextWidget{text: "abcdefghiklmnopqrstuvwxyzw"}
	textWidgetBordered := BorderedWidget{inner: &textWidget}
	textWidgetBordered.Resize(Rect{x: 0, y: 0, x1: 5, y1: 4})

	listWidget := ListWidget{lines: []string{"111111111", "222222222'", "333333333333333", "4444"}}
	listWidgetBordered := BorderedWidget{inner: &listWidget}
	listWidgetBordered.Resize(Rect{x: 6, y: 0, x1: 15, y1: 4})

	textWidget2 := TextWidget{text: "!@#$_+)+_)+_+_((*()&(*&(*(*()*()_)#%$%$$%^$^%$$##@#######%$_%^&*()_+{}:'>?()*()#&(!&(*&!!$&*<?"}
	textwidget2Bordered := BorderedWidget{inner: &textWidget2}
	textwidget2Bordered.Resize(Rect{x: 0, y: 5, x1: termW-1, y1: 10})

    left := TextWidget{text: "LEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFTLEFT"}
    right := TextWidget{text: "RIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGHTRIGH"}
    splittingWidget := SplitWidget{left: &left, right: &right}
    splittingWidget.Resize(Rect{x: 0, y: 11, x1: 30, y1: 15})

	c.widgets = []Widget{
		&listWidgetBordered,
		&textWidgetBordered,
		&textwidget2Bordered,
        &splittingWidget,
	}

	for {
		screen.Fill(' ', tcell.StyleDefault)

		c.Render()

		screen.Show()

		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			// termW, termH := ev.Size()
			//          c.rect = (Rect{x1: termW-1, y1: termY-1})
			c.Render()
			screen.Sync()
		case *tcell.EventKey:
			c.HandleKey(ev)
			if ev.Rune() == 'q' || ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}

}

func main() {
	run()
}
