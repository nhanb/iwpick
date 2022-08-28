package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

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

func main() {
	networks := GetNetworks()
	fmt.Println(networks)

	app := tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false)

	for _, nw := range networks {
		list.AddItem(
			fmt.Sprintf("[%d] %s (%s)", nw.Strength, nw.SSID, nw.Security),
			"", 0, nil)
	}

	if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
