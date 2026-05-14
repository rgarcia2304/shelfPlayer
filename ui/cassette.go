package ui

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var program *tea.Program

func SetProgram(p *tea.Program) { program = p }

type frameMsg struct{}

func startLoop() {
	go func() {
		t := time.NewTicker(33 * time.Millisecond)
		defer t.Stop()

		for range t.C {
			if program != nil {
				program.Send(frameMsg{})
			}
		}
	}()
}

const (
	reelW = 15
	reelH = 9

	windowW = reelW*2 + 6

	nSpokes = 5
)

type Model struct {
	frame int

	tape float64

	leftSpin  float64
	rightSpin float64

	playing bool
	width   int
	height  int
}

func NewModel() Model {
	return Model{playing: true}
}

func (m Model) Init() tea.Cmd {
	startLoop()
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case frameMsg:
		if m.playing {
			// SIMPLE CAPSTAN MOTION (clean + readable)
			m.tape += 0.03

			m.leftSpin += 0.03
			m.rightSpin += 0.03

			m.frame++
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.playing = !m.playing
		}
	}

	return m, nil
}

func (m Model) View() string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		renderWalkman(m),
	)
}

func dirChar(angle float64) string {
	deg := math.Mod(angle*180/math.Pi+360, 360)

	switch {
	case deg < 22.5 || deg >= 337.5:
		return "|"
	case deg < 67.5:
		return "/"
	case deg < 112.5:
		return "-"
	case deg < 157.5:
		return "\\"
	case deg < 202.5:
		return "|"
	case deg < 247.5:
		return "/"
	case deg < 292.5:
		return "-"
	default:
		return "\\"
	}
}

func makeReel(spin float64) []string {

	cx := float64(reelW-1) / 2
	cy := float64(reelH-1) / 2

	rx := float64(reelW-1) / 2
	ry := float64(reelH-1) / 2

	rows := make([]string, reelH)

	for y := 0; y < reelH; y++ {
		var line strings.Builder

		for x := 0; x < reelW; x++ {

			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry

			dist := math.Hypot(nx, ny)
			angle := math.Atan2(ny, nx)

			var c string

			switch {

			case dist > 1.02:
				c = " "

			case dist > 0.85:
				c = dirChar(angle + math.Pi/2)

			case dist < 0.10:
				c = "O"

			default:
				c = "."

				for i := 0; i < nSpokes; i++ {

					sa := spin + float64(i)*2*math.Pi/float64(nSpokes)

					vx := math.Cos(sa)
					vy := math.Sin(sa)

					perp := math.Abs(nx*vy - ny*vx)
					proj := nx*vx + ny*vy

					if perp < 0.07 && proj > 0 {
						c = dirChar(sa)
						break
					}
				}
			}

			line.WriteString(c)
		}

		rows[y] = line.String()
	}

	return rows
}

func renderWalkman(m Model) string {

	left := makeReel(m.leftSpin)
	right := makeReel(m.rightSpin)

	inner := windowW + 8

	sp := func(n int) string { return strings.Repeat(" ", n) }

	shell := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	frame := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	metal := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hub := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	colorize := func(rows []string) []string {
		out := make([]string, len(rows))

		for i, r := range rows {
			var b strings.Builder

			for _, ch := range r {
				s := string(ch)

				switch s {
				case "O":
					b.WriteString(hub.Render(s))
				case "|", "/", "-", "\\":
					b.WriteString(metal.Render(s))
				case " ":
					b.WriteString(" ")
				default:
					b.WriteString(frame.Render(s))
				}
			}

			out[i] = b.String()
		}

		return out
	}

	left = colorize(left)
	right = colorize(right)

	var b strings.Builder
	w := func(s string) { b.WriteString(s + "\n") }

	// OUTER DEVICE BODY (Walkman frame)
	w(shell.Render("+" + strings.Repeat("=", inner) + "+"))

	title := "Terminal Walkman"
	titlePad := (inner - len(title)) / 2

	w(shell.Render("|") + sp(titlePad) + title + sp(titlePad) + shell.Render("|"))

	// top padding / window frame
	w(shell.Render("|") + sp(inner) + shell.Render("|"))

	// CASSETTE WINDOW
	for i := 0; i < reelH; i++ {

		var row strings.Builder

		row.WriteString(shell.Render("|"))
		row.WriteString(sp(4))
		row.WriteString(left[i])
		row.WriteString(sp(6))
		row.WriteString(right[i])
		row.WriteString(sp(4))
		row.WriteString(shell.Render("|"))

		w(row.String())
	}

	w(shell.Render("|") + sp(inner) + shell.Render("|"))

	// controls (simple, device-like)
	w(shell.Render("|") +
		sp(4) +
		metal.Render("[◄◄] [▶] [■]") +
		sp(inner-12-4) +
		shell.Render("|"))

	w(shell.Render("|") + sp(inner) + shell.Render("|"))

	w(shell.Render("+" + strings.Repeat("=", inner) + "+"))

	w("")
	w("  " + hint.Render("space · play/pause     q · quit"))

	return b.String()
}
