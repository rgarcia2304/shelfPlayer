package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rgarcia2304/shelfPlayer/ui"
)

func main() {
	m := ui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
