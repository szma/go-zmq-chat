package main

import (
	t "github.com/gizak/termui"
	"log"
)

const (
	lw = 20

	ih = 3
)

var listItems = []string{}

func ClientSendText(text string) {

}

func showUI(receiveChan chan string, sendChan chan string) {

	err := t.Init()
	if err != nil {
		log.Fatalln("Cannot initialize termui")
	}

	defer t.Close()

	th := t.TermHeight()

	lb := t.NewList()
	lb.Height = th
	lb.BorderLabel = "Users"
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

	ob := t.NewPar("")
	ob.Height = th - ih
	ob.BorderLabel = "Chat"
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

	t.Handle("/sys/kbd/<enter>", func(event t.Event) {
		sendChan <- ib.Text
		ib.Text = ""
		t.Render(ib)
	})

	t.Handle("/sys/kbd/C-8", func(event t.Event) {
		ib.Text = ""
		t.Render(ib)
	})

	t.Handle("/sys/kbd/<space>", func(event t.Event) {
		ib.Text += " "
		t.Render(ib)
	})

	t.Handle("/sys/kbd/<tab>", func(event t.Event) {
		ib.Text += "\t"
		t.Render(ib)
	})

	t.Handle("/sys/kbd", func(event t.Event) {
		kbd := event.Data.(t.EvtKbd) //event.Path[len(event.Path)-1:]
		ib.Text += kbd.KeyStr
		t.Render(ib)
	})

	t.Loop()
}
