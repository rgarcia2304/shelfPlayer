package pomodoro

import "fmt"
type Phase int

const (
	PhaseWork Phase = iota
	PhaseBreak
	PhaseLongBreak
)

type Pomodoro struct {
	Active       bool
	Phase        Phase
	SecondsLeft  int
	WorkMins     int
	BreakMins    int
	LongBreakMins int
	Sessions     int // completed work sessions
}

func New() Pomodoro {
	p := Pomodoro{
		WorkMins:      25,
		BreakMins:     5,
		LongBreakMins: 15,
	}
	p.SecondsLeft = p.WorkMins * 60
	return p
}

func (p *Pomodoro) Toggle() {
	p.Active = !p.Active
}

// Tick decrements the timer by one second.
// Returns true if the phase just changed — caller should react
// (pause music on break, resume on work).
func (p *Pomodoro) Tick() (phaseChanged bool) {
	if !p.Active {
		return false
	}

	p.SecondsLeft--
	if p.SecondsLeft > 0 {
		return false
	}

	// phase ended — advance
	switch p.Phase {
	case PhaseWork:
		p.Sessions++
		if p.Sessions%4 == 0 {
			p.Phase = PhaseLongBreak
			p.SecondsLeft = p.LongBreakMins * 60
		} else {
			p.Phase = PhaseBreak
			p.SecondsLeft = p.BreakMins * 60
		}
	case PhaseBreak, PhaseLongBreak:
		p.Phase = PhaseWork
		p.SecondsLeft = p.WorkMins * 60
	}

	return true
}

func (p *Pomodoro) FormatTime() string {
	m := p.SecondsLeft / 60
	s := p.SecondsLeft % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (p *Pomodoro) PhaseName() string {
	switch p.Phase {
	case PhaseWork:
		return "FOCUS"
	case PhaseBreak:
		return "BREAK"
	case PhaseLongBreak:
		return "LONG BREAK"
	}
	return ""
}

func (p *Pomodoro) IsBreak() bool {
	return p.Phase == PhaseBreak || p.Phase == PhaseLongBreak
}
