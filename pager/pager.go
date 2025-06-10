package pager

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

func New() *Pager {
	return &Pager{}
}

type Pager struct {
	buf bytes.Buffer
}

// runWithPager executes a pager command and pipes the given content to it.
func runWithPager(content string) error {
	// Find the pager to use.

	// The convention is to use the PAGER env var if set, otherwise fall back.
	pager := os.Getenv("PAGER")

	pagerArgs := []string{}
	if pager == "" {
		// Look for 'less' first, then 'more'.
		if path, err := exec.LookPath("less"); err == nil {
			pager = path
			// The -R flag allows `less` to interpret ANSI color escape codes.
			// The -F flag causes `less` to exit if the entire content fits on one screen.
			// The -X flag disables sending termcap init/deinit strings to the terminal.
			pagerArgs = []string{"-R", "-F", "-X"}
		} else if path, err := exec.LookPath("more"); err == nil {
			pager = path
		} else {
			// No pager found, so we'll just print it with line numbers.
			pager = "cat"
			pagerArgs = []string{"-n"}
		}
	}

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
	_, err = io.Copy(pagerIn, bytes.NewBufferString(content))
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
func (p *Pager) WriteFunc() func(fmt.Stringer) {
	_, terminalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// If we can't get the terminal size, fall back to printing directly.
		log.Printf("could not get terminal size: %v, printing directly", err)
		return nil
	}

	return func(s fmt.Stringer) {
		// 2. Check if we are in an interactive terminal.
		// If not, just print the content and exit. We don't want to pipe to `less`
		// if the user is redirecting output to a file.
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return
		}

		// 4. Count the lines in our content.
		lineCount := strings.Count(s.String(), "\n")

		// 5. Decide if we need a pager.
		if lineCount > terminalHeight {
			// Content is taller than the screen, so use a pager.
			if err := runWithPager(s.String()); err != nil {
				// If the pager fails, fall back to printing directly.
				log.Printf("pager failed: %v, printing directly", err)
				fmt.Print(s.String())
			}
		} else {
			// Content fits on the screen, just print it.
			fmt.Print(s.String())
		}
	}
}
