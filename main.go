package main

import (
	"log"

	"github.com/gdamore/tcell"
)

var screen tcell.Screen

type Rect struct{ x, y, w, h int }

type Widget interface {
	Render()
	Resize(rect Rect)
}

type TextWidget struct {
	rect Rect
	text string
}

func (w *TextWidget) Render() {
	runeOrSpace := func(idx int) rune {
		if idx < len(w.text) {
			return rune(w.text[idx])
		} else {
			return ' '
		}
	}
	for j := 0; j < w.rect.h; j++ {
		for i := 0; i < w.rect.w; i++ {
			screen.SetContent(i+w.rect.x, j+w.rect.y, runeOrSpace(j*w.rect.w+i), nil, tcell.StyleDefault)
		}
	}
}
func (w *TextWidget) Resize(rect Rect) {
	w.rect = rect
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

type Compositor struct {
	termRect Rect
	widgets  []Widget
}

func run() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err = screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	cleanup := func() {
		err := recover()
		screen.Fini()
		if err != nil {
			panic(err)
		}
	}
	defer cleanup()

	screen.EnableMouse()
	screen.Clear()

	termW, termH := screen.Size()
	c := Compositor{termRect: Rect{w: termW, h: termH}}

	textWidget := TextWidget{text: "123456789abcefghikl"}
	textWidget.Resize(Rect{x: 0, y: 0, w: 5, h: 20})

	listWidget := ListWidget{lines: []string{"UVXYлфорывылофрвло", "ABCDфлорвлофыр'", "лофырвлдыфрввфалр"}}
	listWidget.Resize(Rect{x: 0, y: 5, w: 4, h: 2})

	c.widgets = []Widget{
		&textWidget,
		&listWidget,
	}

	for {
		for _, w := range c.widgets {
			w.Render()
		}

		screen.Show()

		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			// c.termRect =
            // do layout (resize widgets)
			screen.Sync()
		case *tcell.EventKey:
			if ev.Rune() == 'q' || ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}

}

func main() {
	run()
}
