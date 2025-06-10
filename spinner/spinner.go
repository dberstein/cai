package spinner

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
)

func New(width int) *Spinner {
	return &Spinner{
		chars: []string{"+", "*", "|", "/", "-", "\\"},
		width: width,
	}
}

type Spinner struct {
	chars   []string
	width   int
	running bool
	startAt time.Time
}

func (s *Spinner) IsRunning() bool {
	return s.running
}

func (s *Spinner) StartedAt() time.Time {
	return s.startAt
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
		c := color.New(a...)
		defer c.Printf("\x1b[?25h") // show cursor
		for s.running {
			for i := 0; i < len(s.chars); i++ {
				payload := fmt.Sprintf("%-5s", time.Since(s.startAt).Truncate(10*time.Millisecond))
				c.Printf("\r╰─(%s)%s╯", payload, strings.Repeat("─", s.width-len(payload)-5))
				c.Printf("\x1b[?25l") // hide cursor
				time.Sleep(333 * time.Millisecond)
			}
		}

		c.Println()
	}()

	return f()
}

func (s *Spinner) Start() error {
	s.running = true
	s.startAt = time.Now()
	return nil
}

func (s *Spinner) Stop() error {
	s.running = false
	return nil
}
