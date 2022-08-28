package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false).
		AddItem("List item 1", "Some explanatory text", 0, nil).
		AddItem("List item 2", "Some explanatory text", 0, nil).
		AddItem("List item 3", "Some explanatory text", 0, nil).
		AddItem("List item 4", "Some explanatory text", 0, nil)

	if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
