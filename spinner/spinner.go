package spinner

import (
	"fmt"
	"io"
	"time"

	"github.com/dberstein/cai/tui"
	"github.com/fatih/color"
)

func New(width int) *Spinner {
	return &Spinner{
		width: width,
		sleep: 150 * time.Millisecond,
	}
}

// Spinner is a simple spinner.
type Spinner struct {
	width     int
	running   bool
	startedAt time.Time
	sleep     time.Duration
}

func (s *Spinner) IsRunning() bool {
	return s.running
}

func (s *Spinner) StartedAt() time.Time {
	return s.startedAt
}

func (s *Spinner) Reset() {
	s.running = false
}

func (s *Spinner) Wrap(w io.Writer, f func() error, a ...color.Attribute) error {
	err := s.Start()
	if err != nil {
		return err
	}
	defer s.Stop()

	go func() {
		clr := color.New(a...)
		clr.Printf("\x1b[?25l")       // hide cursor
		defer clr.Printf("\x1b[?25h") // show cursor

		leftColor := color.New(color.FgHiGreen, color.ReverseVideo)
		rightColor := color.New(color.FgHiGreen)

		tui.Spinner(s.IsRunning, s.StartedAt(), s.sleep, s.width, leftColor, rightColor)
	}()

	return f()
}

func (s *Spinner) Start() error {
	s.startedAt = time.Now()
	s.running = true
	return nil
}

func (s *Spinner) Stop() error {
	if s.IsRunning() {
		s.running = false
		fmt.Printf("\n")
	}

	return nil
}
