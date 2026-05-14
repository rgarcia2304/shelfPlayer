package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rgarcia2304/shelfPlayer/audio"
	"github.com/rgarcia2304/shelfPlayer/tape"
	"github.com/rgarcia2304/shelfPlayer/ui"
)

func main() {
	player := audio.NewPlayer()
	defer player.Close()

	tapes, err := tape.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load tapes: %v\n", err)
		os.Exit(1)
	}

	m := ui.NewModel(player, tapes)
	p := tea.NewProgram(m, tea.WithAltScreen())
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
