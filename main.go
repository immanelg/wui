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

type Rect struct{ x, y, w, h int }
func (r Rect) Values() (int, int, int, int) { return r.x, r.y, r.w, r.h }

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
    for i := 0; i < w.rect.w; i++ {
        for j := 0; j < w.rect.h; j++ {
            // text is wrapped
            var r rune = ' '
            if idx := j*w.rect.w+i; idx < len(w.text) { 
                r = rune(w.text[idx])
            }
			screen.SetContent(i+w.rect.x, j+w.rect.y, r, nil, tcell.StyleDefault)
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
    lineCount := len(w.lines)
	for j := 0; j < w.rect.h && j < lineCount; j++ {
		lineLen := len(w.lines[j])
		for i := 0; i < w.rect.w && i < lineLen; i++ {
			screen.SetContent(i+w.rect.x, j+w.rect.y, rune(w.lines[j][i]), nil, tcell.StyleDefault)
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
	rect   Rect
    inner  Widget
}

func (self *BorderedWidget) Render() {
    self.inner.Render()

    x, y, w, h := self.rect.Values()

    style := tcell.StyleDefault
	for i := 0; i < w; i++ {
		screen.SetContent(i+x, y, tcell.RuneHLine, nil, style)
		screen.SetContent(i+x, y+h-1, tcell.RuneHLine, nil, style)
	}
	for j := 0; j < h; j++ {
		screen.SetContent(x,     j+y, tcell.RuneVLine, nil, style)
		screen.SetContent(x+w-1, j+y, tcell.RuneVLine, nil, style)
	}

    screen.SetContent(x,     y, tcell.RuneULCorner, nil, style)
    screen.SetContent(x+w-1, y, tcell.RuneURCorner, nil, style)
    screen.SetContent(x,     y+h-1, tcell.RuneLLCorner, nil, style)
    screen.SetContent(x+w-1, y+h-1, tcell.RuneLRCorner, nil, style)
}

func (w *BorderedWidget) Resize(rect Rect) {
    w.rect = rect
    innerRect := Rect{x: rect.x+1, y: rect.y+1, w: rect.w-2, h: rect.h-2}
	w.inner.Resize(innerRect)
}

func (w *BorderedWidget) GetRect() Rect {
    return w.rect
}

type Compositor struct {
	termRect Rect
	widgets  []Widget
    focusedWidgetId int
}

func (c *Compositor) Render() {
    for _, w := range c.widgets {
        w.Render()
    }
}

func (c *Compositor) HandleKey(key *tcell.EventKey) {
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
	c := Compositor{termRect: Rect{w: termW, h: termH}}

    textWidget := TextWidget{text: "abcdefghiklmnopqrstuvwxyzw"}
    textWidgetBordered := BorderedWidget{inner: &textWidget}
    textWidgetBordered.Resize(Rect{x: 0, y: 0, w: 5, h: 4})

    listWidget := ListWidget{lines: []string{"111111111", "222222222'", "333333333333333", "4444"}}
    listWidgetBordered := BorderedWidget{inner: &listWidget}
    listWidgetBordered.Resize(Rect{x: 0, y: 4, w: 7, h: 5})

    textWidget2 := TextWidget{text: "!@#$%^&*()_+{}:'>?()*()#&(!&(*&!!$&*<?"}
    textwidget2Bordered := BorderedWidget{inner: &textWidget2}
    textwidget2Bordered.Resize(Rect{x: 0, y: 9, w: 10, h: 10})

	c.widgets = []Widget{
		&listWidgetBordered,
        &textWidgetBordered,
        &textwidget2Bordered,
	}

	for {
		screen.Fill(' ', tcell.StyleDefault)

        c.Render()

		screen.Show()

		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
            // Do layouting...
            // termW, termH := ev.Size()
            // c.termRect = Rect{w: termW, h: termH}
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
