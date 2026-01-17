package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

type Spinner struct {
	message string
	step    string
	start   time.Time
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		step:    "Thinking",
		start:   time.Now(),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *Spinner) SetStep(step string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.step = step
}

func (s *Spinner) Start() {
	go s.run()
}

func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
	fmt.Print("\r\033[K")
}

func (s *Spinner) run() {
	defer close(s.done)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	idx := 0
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	for {
		select {
		case <-s.stop:
			return
		case <-tick.C:
			s.mu.Lock()
			elapsed := time.Since(s.start)
			mins := int(elapsed.Minutes())
			secs := int(elapsed.Seconds()) % 60
			char := spinnerChars[idx%len(spinnerChars)]
			line := fmt.Sprintf("  %c %s │ %s %s",
				char,
				cyan.Render(fmt.Sprintf("%-16s", s.step)),
				s.message,
				dim.Render(fmt.Sprintf("[%02d:%02d]", mins, secs)),
			)
			s.mu.Unlock()
			fmt.Print("\r\033[K" + line)
			idx++
		}
	}
}
