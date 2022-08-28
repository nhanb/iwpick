package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Network struct {
	SSID        string
	Security    string
	Strength    int
	StrengthStr string
	IsCurrent   bool
}

func GetNetworks() (networks []Network) {
	scanCmd := exec.Command("iwctl", "station", "wlan0", "scan")
	if err := scanCmd.Run(); err != nil {
		log.Fatal(err)
	}

	listCmd := exec.Command("iwctl", "station", "wlan0", "get-networks")
	var out bytes.Buffer
	listCmd.Stdout = &out
	if err := listCmd.Run(); err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(strings.NewReader(out.String()))
	for i := 0; i < 4; i++ {
		s.Scan() // ignore first 4 lines, which are headers
	}
	for s.Scan() {
		nw := Network{}

		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		// Wifi strength is at end of line, indicated by number of white
		// asterisks, right-padded with grey asterisks to always reach 4 chars.
		lastSpaceIndex := strings.LastIndex(line, " ")
		strengthStr := line[lastSpaceIndex+1:]
		// \x1b (ANSI escape) starts the grey color switch, which means
		// whatever precedes it is the actual strength.
		nw.Strength = strings.Index(strengthStr, "\x1b")
		if nw.Strength == -1 {
			// no grey asterisks means all 4 are white asterisks
			// => full strength
			nw.Strength = 4
		}

		// We don't need formatting info for the rest of the data
		line = StripANSI(line[:lastSpaceIndex])
		parts := strings.Fields(line)

		if parts[0] == ">" {
			nw.IsCurrent = true
			parts = parts[1:]
		} else {
			nw.IsCurrent = false
		}

		size := len(parts)
		nw.Security = parts[size-1]
		nw.SSID = strings.Join(parts[:size-1], " ")

		networks = append(networks, nw)
	}
	return networks
}

func drawListItems(networks []Network, list *tview.List, query string) {
	list.Clear()
	query = strings.ToLower(query)

	for _, nw := range networks {
		if query != "" && !strings.Contains(strings.ToLower(nw.SSID), query) {
			continue
		}
		currentStr := ""
		if nw.IsCurrent {
			currentStr = " (*)"
		}
		itemName := fmt.Sprintf(
			"[%d]%s %s (%s)", nw.Strength, currentStr, nw.SSID, nw.Security,
		)
		list.AddItem(itemName, "", 0, nil)
	}
}

func main() {
	networks := GetNetworks()
	fmt.Println(networks)

	app := tview.NewApplication().EnableMouse(false)

	input := tview.NewInputField().SetLabel("Filter: ")
	list := tview.NewList().ShowSecondaryText(false)
	drawListItems(networks, list, "")

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(input, 2, 1, true)
	flex.AddItem(list, 0, 1, false)

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			i := list.GetCurrentItem()
			if i > 0 {
				list.SetCurrentItem(i - 1)
			}
		case tcell.KeyDown:
			list.SetCurrentItem(list.GetCurrentItem() + 1)
		case tcell.KeyEsc:
			app.Stop()
		}
		return event
	})

	input.SetChangedFunc(func(query string) {
		drawListItems(networks, list, query)
	})

	if err := app.SetRoot(flex, true).SetFocus(input).Run(); err != nil {
		panic(err)
	}
}
