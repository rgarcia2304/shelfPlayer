package audio

import (
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type Player struct {
	mu      sync.Mutex
	ctrl    *beep.Ctrl
	format  beep.Format
	file    *os.File
	playing bool
	ready   bool
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Load(path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// close previous
	if p.file != nil {
		speaker.Clear()
		p.file.Close()
		p.file = nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return err
	}

	// init speaker once — safe to call again with same sample rate
	if !p.ready {
		if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
			f.Close()
			return err
		}
		p.ready = true
	}

	p.ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
	p.format = format
	p.file = f
	p.playing = true

	speaker.Play(p.ctrl)
	return nil
}

func (p *Player) Toggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	p.playing = !p.ctrl.Paused
	speaker.Unlock()
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

func (p *Player) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	speaker.Clear()
	if p.file != nil {
		p.file.Close()
	}
}
