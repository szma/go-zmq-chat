package main

import (
	t "github.com/gizak/termui"
	"log"
)

const (
	lw = 20

	ih = 5
)

var listItems = []string{
	"Line 1",
	"Line 2",
	"Line 3",
	"Line 4",
	"Line 5",
}

func main() {

	err := t.Init()
	if err != nil {
		log.Fatalln("Cannot initialize termui")
	}

	defer t.Close()

	th := t.TermHeight()

	lb := t.NewList()
	lb.Height = th
	lb.BorderLabel = "List"
	lb.BorderLabelFg = t.ColorGreen
	lb.BorderFg = t.ColorGreen
	lb.ItemFgColor = t.ColorWhite
	lb.Items = listItems

	ib := t.NewPar("")
	ib.Height = ih
	ib.BorderLabel = "Input"
	ib.BorderLabelFg = t.ColorYellow
	ib.BorderFg = t.ColorYellow
	ib.TextFgColor = t.ColorWhite

	ob := t.NewPar("\nPress Ctrl-C to quit")
	ob.Height = th - ih
	ob.BorderLabel = "Output"
	ob.BorderLabelFg = t.ColorCyan
	ob.BorderFg = t.ColorCyan
	ob.TextFgColor = t.ColorWhite

	t.Body.AddRows(
		t.NewRow(
			t.NewCol(9, 0, ob, ib),
			t.NewCol(3, 0, lb)))

	t.Body.Align()
	t.Render(t.Body)

	t.Handle("/sys/wnd/resize", func(t.Event) {

		lb.Height = t.TermHeight()
		ob.Height = t.TermHeight() - ih
		t.Body.Width = t.TermWidth()
		t.Body.Align()
		t.Render(t.Body)
	})

	t.Handle("/sys/kbd/C-c", func(t.Event) {
		t.StopLoop()
	})

	t.Handle("/sys/kbd", func(t.Event) {
		ib.Text += "h"
		t.Render(ib)
	})

	t.Loop()
}
