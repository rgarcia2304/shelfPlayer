package tape

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"os"
	"strings"
)

type Track struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	File   string `json:"file"` // filename only, relative to tape dir
}

type Tape struct {
	Name   string  `json:"name"`
	Artist string  `json:"artist"`
	Color  string  `json:"color"` // lipgloss color for the tape label
	Tracks []Track `json:"tracks"`
	Dir    string  `json:"-"` // absolute path to tape folder, not stored in json
}

func (t *Tape) TrackPath(track Track) string {
	return filepath.Join(t.Dir, track.File)
}

func RootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(home, ".shelfplayer", "tapes")
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	return root, nil
}
// LoadAll scans the tapes directory and returns all valid tapes
func LoadAll() ([]*Tape, error) {
	root, err := RootDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var tapes []*Tape
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		tapeDir := filepath.Join(root, entry.Name())
		tape, err := loadTape(tapeDir)
		if err != nil {
			// skip malformed tapes silently
			continue
		}
		tapes = append(tapes, tape)
	}

	// sort alphabetically by name
	sort.Slice(tapes, func(i, j int) bool {
		return tapes[i].Name < tapes[j].Name
	})

	return tapes, nil
}

func loadTape(dir string) (*Tape, error) {
	metaPath := filepath.Join(dir, "tape.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		// no tape.json — try to build one from mp3s in the folder
		return buildFromDir(dir)
	}

	var t Tape
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	t.Dir = dir
	return &t, nil
}

// buildFromDir creates a Tape from a folder of MP3s with no tape.json
// Useful for dropping files in manually without metadata
func buildFromDir(dir string) (*Tape, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tracks []Track
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".mp3") {
			continue
		}
		// strip extension and leading track numbers for display
		title := strings.TrimSuffix(name, filepath.Ext(name))
		title = strings.TrimLeft(title, "0123456789. -_")
		if title == "" {
			title = name
		}
		tracks = append(tracks, Track{
			Title: title,
			File:  name,
		})
	}

	if len(tracks) == 0 {
		return nil, os.ErrNotExist
	}

	// use folder name as tape name
	folderName := filepath.Base(dir)
	tapeName := strings.ReplaceAll(folderName, "-", " ")
	tapeName = strings.ReplaceAll(tapeName, "_", " ")

	return &Tape{
		Name:   tapeName,
		Tracks: tracks,
		Dir:    dir,
		Color:  "205", // default pink
	}, nil
}

// Save writes tape.json to the tape directory
func (t *Tape) Save() error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(t.Dir, "tape.json"), data, 0644)
}
