package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"regexp"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rgarcia2304/shelfPlayer/tape"
)

type searchResultMsg struct {
	results []searchResult
	err     error
}

type downloadDoneMsg struct {
	track tape.Track
	err   error
}

type searchResult struct {
	title  string
	artist string
	query  string
}

type creatorStep int

const (
	stepName creatorStep = iota
	stepSearch
	stepResults
	stepDownloading
)

type Creator struct {
	step         creatorStep
	tapeName     string
	nameInput    string
	searchInput  string
	results      []searchResult
	resultCursor int
	tracks       []tape.Track
	downloading  string
	err          string
}

func NewCreator() Creator {
	return Creator{step: stepName}
}

func (m Model) updateCreator(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	c := &m.creator

	switch c.step {

	case stepName:
		switch msg.String() {
		case "esc":
			m.screen = screenLibrary
		case "enter":
			if strings.TrimSpace(c.nameInput) != "" {
				c.tapeName = strings.TrimSpace(c.nameInput)
				c.step = stepSearch
			}
		case "backspace":
			if len(c.nameInput) > 0 {
				c.nameInput = c.nameInput[:len(c.nameInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				c.nameInput += msg.String()
			}
		}

	case stepSearch:
		switch msg.String() {
		case "esc":
			c.step = stepName
		case "ctrl+s":
			if len(c.tracks) > 0 {
				if err := saveTape(c.tapeName, c.tracks); err != nil {
					c.err = err.Error()
				} else {
					newTapes, _ := tape.LoadAll()
					m.tapes = newTapes
					m.screen = screenLibrary
					m.creator = NewCreator()
				}
			}
		case "enter":
			if strings.TrimSpace(c.searchInput) != "" {
				return m, doSearch(c.searchInput)
			}
		case "backspace":
			if len(c.searchInput) > 0 {
				c.searchInput = c.searchInput[:len(c.searchInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				c.searchInput += msg.String()
			}
		}

	case stepResults:
		switch msg.String() {
		case "esc":
			c.step = stepSearch
			c.results = nil
		case "up", "k":
			if c.resultCursor > 0 {
				c.resultCursor--
			}
		case "down", "j":
			if c.resultCursor < len(c.results)-1 {
				c.resultCursor++
			}
		case "enter":
			if len(c.results) > 0 {
				r := c.results[c.resultCursor]
				c.downloading = r.title
				c.step = stepDownloading
				trackNum := len(c.tracks) + 1
				return m, doDownload(c.tapeName, r, trackNum)
			}
		case "ctrl+s":
			if len(c.tracks) > 0 {
				if err := saveTape(c.tapeName, c.tracks); err != nil {
					c.err = err.Error()
				} else {
					newTapes, _ := tape.LoadAll()
					m.tapes = newTapes
					m.screen = screenLibrary
					m.creator = NewCreator()
				}
			}
		}

	case stepDownloading:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		args := []string{
			"ytsearch5:" + query,
			"--print", "%(title)s|||%(uploader)s|||%(webpage_url)s",
			"--no-playlist",
			"--no-download",
			"--quiet",
		}
		out, err := exec.Command("yt-dlp", args...).Output()
		if err != nil {
			return searchResultMsg{err: err}
		}

		var results []searchResult
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|||")
			if len(parts) >= 3 {
				results = append(results, searchResult{
					title:  cleanTitle(parts[0]),
					artist: parts[1],
					query:  parts[2],
				})
			}
		}
		return searchResultMsg{results: results}
	}
}

func doDownload(tapeName string, r searchResult, trackNum int) tea.Cmd {
	return func() tea.Msg {
		root, err := tape.RootDir()
		if err != nil {
			return downloadDoneMsg{err: err}
		}
		tapeDir := filepath.Join(root, sanitize(tapeName))
		if err := os.MkdirAll(tapeDir, 0755); err != nil {
			return downloadDoneMsg{err: err}
		}

		filename := fmt.Sprintf("%02d-%s.mp3", trackNum, sanitize(r.title))
		outPath := filepath.Join(tapeDir, filename)

		args := []string{
			"-x",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"-o", outPath,
			"--quiet",
			r.query,
		}

		if err := exec.Command("yt-dlp", args...).Run(); err != nil {
			return downloadDoneMsg{err: err}
		}

		return downloadDoneMsg{
			track: tape.Track{
				Title:  cleanTitle(r.title),
				Artist: r.artist,
				File:   filename,
			},
		}
	}
}
func saveTape(name string, tracks []tape.Track) error {
	root, err := tape.RootDir()
	if err != nil {
		return err
	}
	tapeDir := filepath.Join(root, sanitize(name))
	if err := os.MkdirAll(tapeDir, 0755); err != nil {
		return err
	}
	t := &tape.Tape{
		Name:   name,
		Tracks: tracks,
		Dir:    tapeDir,
		Color:  "136",
	}
	return t.Save()
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var out strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			out.WriteRune(ch)
		}
	}
	return out.String()
}

func (m Model) viewCreator() string {
	c := m.creator

	title  := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	dim    := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	input  := lipgloss.NewStyle().Foreground(lipgloss.Color("77"))
	amber  := lipgloss.NewStyle().Foreground(lipgloss.Color("136"))
	sel    := lipgloss.NewStyle().Foreground(lipgloss.Color("77")).Bold(true)
	hint   := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	errSty := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	num    := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	var b strings.Builder
	w := func(s string) { b.WriteString(s + "\n") }

	w(amber.Render("  + new tape"))
	w("")

	switch c.step {

	case stepName:
		w(dim.Render("  tape name"))
		w(fmt.Sprintf("  %s%s", input.Render(c.nameInput), dim.Render("_")))
		w("")
		w(hint.Render("  enter . confirm    esc . back"))

	case stepSearch:
		w(title.Render("  " + c.tapeName))
		w("")
		if len(c.tracks) > 0 {
			w(amber.Render("  added:"))
			for i, t := range c.tracks {
				w(fmt.Sprintf("  %s  %s",
					num.Render(fmt.Sprintf("%02d", i+1)),
					dim.Render(t.Title),
				))
			}
			w("")
		}
		w(dim.Render("  search for a track"))
		w(fmt.Sprintf("  > %s%s", input.Render(c.searchInput), dim.Render("_")))
		w("")
		w(hint.Render("  enter . search    ctrl+s . save tape    esc . back"))

	case stepResults:
		w(title.Render("  " + c.tapeName))
		w("")
		w(dim.Render("  select a track"))
		w("")
		if len(c.results) == 0 {
			w(dim.Render("  no results found"))
		}
		for i, r := range c.results {
			if i == c.resultCursor {
				w(fmt.Sprintf("  %s %s  %s",
					amber.Render("▌"),
					sel.Render(r.title),
					dim.Render(r.artist),
				))
			} else {
				w(fmt.Sprintf("    %s  %s",
					dim.Render(r.title),
					dim.Render(r.artist),
				))
			}
		}
		w("")
		w(hint.Render("  enter . add    ctrl+s . save tape    esc . search again"))

	case stepDownloading:
		w(title.Render("  " + c.tapeName))
		w("")
		w(dim.Render("  downloading"))
		w(fmt.Sprintf("  %s", amber.Render(c.downloading)))
		w("")
		w(dim.Render("  please wait..."))
	}

	if c.err != "" {
		w("")
		w(errSty.Render("  error: " + c.err))
	}

	return b.String()
}

func cleanTitle(raw string) string {
	// remove anything in parentheses or brackets
	parenRe := regexp.MustCompile(`\s*[\(\[].*?[\)\]]`)
	raw = parenRe.ReplaceAllString(raw, "")

	// remove common separators and artist prefixes

	separators := []string{" — ", " – ", " - "}
	for _, sep := range separators {
		if idx := strings.Index(raw, sep); idx != -1 {
			raw = raw[idx+len(sep):]
		}
	}

	return strings.TrimSpace(raw)
}
