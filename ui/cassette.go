package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rgarcia2304/shelfPlayer/audio"
	"github.com/rgarcia2304/shelfPlayer/tape"
)

var program *tea.Program

func SetProgram(p *tea.Program) { program = p }

type frameMsg struct{}
type NextTrackMsg struct{}

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

type screen int

const (
	screenLibrary screen = iota
	screenDetail
	screenPlayer
	screenCreator
)

const (
	reelW    = 15
	reelH    = 9
	nSpokes  = 5
	boxInner = 44
)

type Model struct {
	screen       screen
	width        int
	height       int
	player       *audio.Player
	tapes        []*tape.Tape
	cursor       int
	trackCursor  int
	selectedTape *tape.Tape
	currentTape  *tape.Tape
	currentTrack int
	frame        int
	tape         float64
	leftSpin     float64
	rightSpin    float64
	playing      bool
	creator      Creator
}

func NewModel(player *audio.Player, tapes []*tape.Tape) Model {
	return Model{screen: screenLibrary, player: player, tapes: tapes}
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

	case NextTrackMsg:
		if m.currentTape != nil && m.currentTrack < len(m.currentTape.Tracks)-1 {
			m.currentTrack++
			if err := m.player.Load(m.currentTape.TrackPath(m.currentTape.Tracks[m.currentTrack])); err == nil {
				m.playing = true
			}
		} else {
			m.playing = false
		}

	case frameMsg:
		if m.playing {
			m.tape += 0.03
			m.leftSpin += 0.03
			m.rightSpin += 0.03
			m.frame++
		}

	case searchResultMsg:
		m.creator.results = msg.results
		if msg.err != nil {
			m.creator.err = msg.err.Error()
			m.creator.step = stepSearch
		} else {
			m.creator.step = stepResults
			m.creator.resultCursor = 0
		}

	case downloadDoneMsg:
		if msg.err != nil {
			m.creator.err = msg.err.Error()
			m.creator.step = stepResults
		} else {
			m.creator.tracks = append(m.creator.tracks, msg.track)
			m.creator.step = stepSearch
			m.creator.searchInput = ""
			m.creator.downloading = ""
		}

	case tea.KeyMsg:
		switch m.screen {
		case screenLibrary:
			return m.updateLibrary(msg)
		case screenDetail:
			return m.updateDetail(msg)
		case screenPlayer:
			return m.updatePlayer(msg)
		case screenCreator:
			return m.updateCreator(msg)
		}
	}

	return m, nil
}

func (m Model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tapes)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.tapes) > 0 {
			m.selectedTape = m.tapes[m.cursor]
			m.trackCursor = 0
			m.screen = screenDetail
		}
	case "n":
		m.creator = NewCreator()
		m.screen = screenCreator
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "b":
		m.screen = screenLibrary
	case "enter", " ":
		if m.selectedTape == nil || len(m.selectedTape.Tracks) == 0 {
			break
		}
		m.currentTape = m.selectedTape
		m.currentTrack = 0
		m.playing = false
		m.screen = screenPlayer
		if err := m.player.Load(m.currentTape.TrackPath(m.currentTape.Tracks[0])); err == nil {
			m.playing = true
		}
	case "up", "k":
		if m.trackCursor > 0 {
			m.trackCursor--
		}
	case "down", "j":
		if m.selectedTape != nil && m.trackCursor < len(m.selectedTape.Tracks)-1 {
			m.trackCursor++
		}
	}
	return m, nil
}

func (m Model) updatePlayer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "b":
		m.player.Toggle()
		m.playing = false
		m.screen = screenLibrary
	case " ":
		m.player.Toggle()
		m.playing = !m.playing
	case "n", "right":
		if m.currentTape != nil && m.currentTrack < len(m.currentTape.Tracks)-1 {
			m.currentTrack++
			m.player.Load(m.currentTape.TrackPath(m.currentTape.Tracks[m.currentTrack]))
			m.playing = true
		}
	case "p", "left":
		if m.currentTape != nil && m.currentTrack > 0 {
			m.currentTrack--
			m.player.Load(m.currentTape.TrackPath(m.currentTape.Tracks[m.currentTrack]))
			m.playing = true
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenLibrary:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.viewLibrary())
	case screenDetail:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.viewDetail())
	case screenPlayer:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			renderWalkman(m))
	case screenCreator:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.viewCreator())
	}
	return ""
}

func (m Model) viewLibrary() string {
	title    := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	dim      := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selected := lipgloss.NewStyle().Foreground(lipgloss.Color("77")).Bold(true)
	artist   := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	hint     := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	icon     := lipgloss.NewStyle().Foreground(lipgloss.Color("136"))

	var b strings.Builder
	b.WriteString(title.Render("  shelf player") + "\n")
	b.WriteString(dim.Render("  your tapes") + "\n\n")

	if len(m.tapes) == 0 {
		b.WriteString(dim.Render("  no tapes found") + "\n")
		b.WriteString(dim.Render("  add folders to ~/shelfPlayer/tapes/") + "\n")
	}

	for i, t := range m.tapes {
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("  %s %s  %s\n",
				icon.Render("▌"),
				selected.Render(t.Name),
				artist.Render(t.Artist),
			))
			b.WriteString(fmt.Sprintf("      %s\n", dim.Render(fmt.Sprintf("%d tracks", len(t.Tracks)))))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %s\n",
				dim.Render(t.Name),
				artist.Render(t.Artist),
			))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(hint.Render("  ↑↓ . navigate    enter . inspect    n . new tape    q . quit"))
	return b.String()
}

func (m Model) viewDetail() string {
	if m.selectedTape == nil {
		return ""
	}
	t := m.selectedTape

	title    := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	dim      := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selected := lipgloss.NewStyle().Foreground(lipgloss.Color("77")).Bold(true)
	artist   := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	num      := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	hint     := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	amber    := lipgloss.NewStyle().Foreground(lipgloss.Color("136"))

	var b strings.Builder

	b.WriteString(amber.Render("  ▌ ") + title.Render(t.Name) + "\n")
	if t.Artist != "" {
		b.WriteString(dim.Render("    "+t.Artist) + "\n")
	}
	b.WriteString(dim.Render(fmt.Sprintf("    %d tracks", len(t.Tracks))) + "\n")
	b.WriteString("\n")
	b.WriteString(dim.Render("  ────────────────────────────────") + "\n\n")

	for i, tr := range t.Tracks {
		numStr := fmt.Sprintf("%02d", i+1)
		if i == m.trackCursor {
			b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				num.Render(numStr),
				selected.Render(tr.Title),
				artist.Render(tr.Artist),
			))
		} else {
			b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				num.Render(numStr),
				dim.Render(tr.Title),
				artist.Render(tr.Artist),
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(dim.Render("  ────────────────────────────────") + "\n\n")
	b.WriteString(hint.Render("  enter . play    esc . back    q . quit"))
	return b.String()
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

func colorizeReel(rows []string) []string {
	metal := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hub   := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	fill  := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

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
				b.WriteString(fill.Render(s))
			}
		}
		out[i] = b.String()
	}
	return out
}

func boxRow(content string, shellStyle, contentStyle lipgloss.Style) string {
	vis := lipgloss.Width(content)
	if vis < boxInner {
		content = content + strings.Repeat(" ", boxInner-vis)
	} else if vis > boxInner {
		content = content[:boxInner]
	}
	return shellStyle.Render("|") + contentStyle.Render(content) + shellStyle.Render("|")
}

func centerInBox(text string) string {
	n := len(text)
	if n >= boxInner {
		return text[:boxInner]
	}
	left := (boxInner - n) / 2
	right := boxInner - n - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func renderWalkman(m Model) string {
	shell  := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	metal  := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("77")).Bold(true)
	dim    := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	hint   := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	none   := lipgloss.NewStyle()

	reelRows := colorizeReel(makeReel(m.leftSpin))

	tapeName  := ""
	trackName := "no tape loaded"
	trackNum  := ""
	if m.currentTape != nil {
		tapeName = m.currentTape.Name
		if m.currentTrack < len(m.currentTape.Tracks) {
			tr := m.currentTape.Tracks[m.currentTrack]
			trackName = tr.Title
			if tr.Artist != "" {
				trackName = tr.Title + "  .  " + tr.Artist
			}
			trackNum = fmt.Sprintf("%02d / %02d", m.currentTrack+1, len(m.currentTape.Tracks))
		}
	}

	statusColored := accent.Render(">") + dim.Render(" playing")
	if !m.playing {
		statusColored = dim.Render("= stopped")
	}

	var ctrlColored string
	if m.playing {
		ctrlColored = shell.Render("|<") + "  " +
			shell.Render("[ ]") + "  " +
			accent.Render("[>]") + "  " +
			shell.Render(">|")
	} else {
		ctrlColored = shell.Render("|<") + "  " +
			accent.Render("[ ]") + "  " +
			shell.Render("[>]") + "  " +
			shell.Render(">|")
	}

	const ctrlGap = boxInner - 2 - 9 - 16
	ctrlRowColored := "  " + statusColored + strings.Repeat(" ", ctrlGap) + ctrlColored

	const reelPad = 4
	const reelGap = boxInner - 2*reelW - 2*reelPad

	var b strings.Builder
	ln := func(s string) { b.WriteString(s + "\n") }

	ln(shell.Render("+" + strings.Repeat("=", boxInner) + "+"))
	ln(boxRow(centerInBox(tapeName), shell, dim))
	ln(boxRow("", shell, none))

	for i := 0; i < reelH; i++ {
		ln(shell.Render("|") +
			strings.Repeat(" ", reelPad) +
			reelRows[i] +
			strings.Repeat(" ", reelGap) +
			reelRows[i] +
			strings.Repeat(" ", reelPad) +
			shell.Render("|"))
	}

	ln(boxRow("", shell, none))
	ln(boxRow(centerInBox(trackName), shell, metal))
	ln(boxRow(centerInBox(trackNum), shell, dim))
	ln(boxRow("", shell, none))
	ln(shell.Render("|") + ctrlRowColored + shell.Render("|"))
	ln(boxRow("", shell, none))
	ln(shell.Render("+" + strings.Repeat("=", boxInner) + "+"))

	b.WriteString("\n  " + hint.Render("space . play/pause    n/p . tracks    esc . library    q . quit") + "\n")

	return b.String()
}
