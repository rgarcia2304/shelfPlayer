package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rgarcia2304/shelfPlayer/audio"
	"github.com/rgarcia2304/shelfPlayer/tape"
	"github.com/rgarcia2304/shelfPlayer/ui"
	"os/exec"
)

func main() {
	player := audio.NewPlayer()
	defer player.Close()

	if _, err := exec.LookPath("yt-dlp"); err != nil{
		fmt.Fprintf(os.Stderr, "\n shelfplayer requires yt-dlp to create tapes\n")
		fmt.Fprintf(os.Stderr, " install it with: brew install yt-dlp\n\n")
		os.Exit(1)
	}

	tapes, err := tape.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load tapes: %v\n", err)
		os.Exit(1)
	}

	m := ui.NewModel(player, tapes)
	p := tea.NewProgram(m, tea.WithAltScreen())
	ui.SetProgram(p)

	player.OnFinish(func() {
		p.Send(ui.NextTrackMsg{})
	})

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
