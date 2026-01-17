package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// ActivityProvider provides information about engine activity.
type ActivityProvider interface {
	LastOutputAgo() time.Duration
	HasOutput() bool
}

type Spinner struct {
	message  string
	step     string
	start    time.Time
	stop     chan struct{}
	done     chan struct{}
	activity ActivityProvider
	mu       sync.Mutex
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

// SetActivity sets the activity provider for heartbeat display.
func (s *Spinner) SetActivity(activity ActivityProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activity = activity
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
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

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

			// Build activity indicator
			activityStr := ""
			if s.activity != nil && s.activity.HasOutput() {
				lastAgo := s.activity.LastOutputAgo()
				if lastAgo < 2*time.Second {
					activityStr = green.Render(" ● active")
				} else {
					agoSecs := int(lastAgo.Seconds())
					if agoSecs < 60 {
						activityStr = dim.Render(fmt.Sprintf(" ○ %ds ago", agoSecs))
					} else {
						agoMins := agoSecs / 60
						agoSecs = agoSecs % 60
						activityStr = dim.Render(fmt.Sprintf(" ○ %dm%ds ago", agoMins, agoSecs))
					}
				}
			}

			line := fmt.Sprintf("  %c %s │ %s %s%s",
				char,
				cyan.Render(fmt.Sprintf("%-16s", s.step)),
				s.message,
				dim.Render(fmt.Sprintf("[%02d:%02d]", mins, secs)),
				activityStr,
			)
			s.mu.Unlock()
			fmt.Print("\r\033[K" + line)
			idx++
		}
	}
}
