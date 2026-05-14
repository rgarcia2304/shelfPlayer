package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rgarcia2304/shelfPlayer/audio"
	"github.com/rgarcia2304/shelfPlayer/ui"
	"github.com/rgarcia2304/shelfPlayer/tape"
)

func main() {
	player := audio.NewPlayer()
	defer player.Close()

	// pass player into the model
	m := ui.NewModel(player)
	p := tea.NewProgram(m, tea.WithAltScreen())
	ui.SetProgram(p)

	tapes, err := tape.LoadAll()
	if err != nil {
    		fmt.Fprintf(os.Stderr, "failed to load tapes: %v\n", err)
    		os.Exit(1)
	}

	for _, t := range tapes {
    		fmt.Printf("found tape: %s (%d tracks)\n", t.Name, len(t.Tracks))
    
		for _, tr := range t.Tracks {
        		fmt.Printf("  - %s\n", tr.Title)
    		}
	}
	
	if err := player.Load("test.mp3"); err != nil {
    		fmt.Fprintf(os.Stderr, "failed to load audio: %v\n", err)
    		os.Exit(1)
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
