package pager

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"golang.org/x/term"
)

func New() *Pager {
	return &Pager{}
}

type Pager struct {
	buf bytes.Buffer
}

// runWithPager executes a pager command and pipes the given content to it.
func runWithPager(content fmt.Stringer) error {
	// // The convention is to use the PAGER env var if set, otherwise fall back.
	// pager := os.Getenv("PAGER")
	var pager string

	pagerArgs := []string{}

	// Look for 'less'.
	if path, err := exec.LookPath("less"); err == nil {
		pager = path
		pagerArgs = []string{
			"-R", // The -R flag allows `less` to interpret ANSI color escape codes.
			"-F", // The -F flag causes `less` to exit if the entire content fits on one screen.
			"-X", // The -X flag disables sending termcap init/deinit strings to the terminal.
			"-N", // The -N flag causes `less` to print line numbers.
		}
	} else {
		// No pager found, so we'll just print it with line numbers.
		pager = "cat"
		pagerArgs = append(pagerArgs, "-n")
	}
	// }

	cmd := exec.Command(pager, pagerArgs...)

	// Create a pipe to send our content to the pager.
	pagerIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	// The pager should write to the original TTY.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the pager.
	if err := cmd.Start(); err != nil {
		return err
	}

	// Write our content to the pager's Stdin.
	// We use a buffer to ensure the write is non-blocking.
	// A direct io.WriteString might block if the content is large.
	_, err = io.Copy(pagerIn, bytes.NewBufferString(content.String()))
	if err != nil {
		// If we can't write to the pager, we should probably kill it.
		_ = cmd.Process.Kill()
		return fmt.Errorf("failed to write to pager stdin: %w", err)
	}

	// Close the pipe to signal EOF to the pager.
	if err := pagerIn.Close(); err != nil {
		return err
	}

	// Wait for the pager to exit.
	// The user will interact with the pager until they quit (e.g., by pressing 'q').
	return cmd.Wait()
}

func (p *Pager) GetSize() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

func (p *Pager) WriteFunc(s fmt.Stringer) {
	// Check if we are in an interactive terminal.
	// If not, just print the content and exit. We don't want to pipe to `less`
	// if the user is redirecting output to a file.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}

	// Always run through pager.
	if err := runWithPager(s); err != nil {
		log.Printf("pager failed: %v", err)
	}

	// Always print content after the pager.
	fmt.Print("\033[H\033[2J")
	fmt.Printf(s.String())
}
