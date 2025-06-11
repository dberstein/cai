package spinner

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func New(width int) *Spinner {
	return &Spinner{
		width: width,
		sleep: 123 * time.Millisecond,
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

		for s.IsRunning() {
			stopwatch := time.Since(s.StartedAt()).Truncate(100 * time.Millisecond)
			w := s.width - 12 - len(stopwatch.String())
			for i := 0; i < w; i++ {
				if !s.IsRunning() {
					break
				}
				stopwatch = time.Since(s.StartedAt()).Truncate(100 * time.Millisecond)
				clr.Fprintf(os.Stdout, "\r╰─%s─┤%s├─%s─╯", strings.Repeat("─", i), stopwatch.String(), strings.Repeat("─", w-i-1))
				time.Sleep(s.sleep)
			}
			for i := w; i > 0; i-- {
				if !s.IsRunning() {
					break
				}
				stopwatch = time.Since(s.StartedAt()).Truncate(100 * time.Millisecond)
				clr.Fprintf(os.Stdout, "\r╰─%s─┤%s├─%s─╯", strings.Repeat("─", i), stopwatch.String(), strings.Repeat("─", w-i+1))
				time.Sleep(s.sleep)
			}
		}
	}()

	return f()
}

func (s *Spinner) Start() error {
	s.startedAt = time.Now()
	s.running = true
	return nil
}

func (s *Spinner) Stop() error {
	s.running = false
	return nil
}
