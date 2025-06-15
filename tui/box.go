package tui

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Helpful characters:
// ═ ║ ╒ ╓ ╔ ╕ ╖ ╗ ╘ ╙ ╚ ╛ ╜ ╝ ╞ ╟ ╠ ╡ ╢ ╣ ╤ ╥ ╦ ╧ ╨ ╩ ╪ ╫ ╬
// ┌ ┍ ┎ ┏ ┐ ┑ ┒ ┓ └ ┕ ┖ ┗ ┘ ┙ ┚ ┛
// ╭ ╮ ╯ ╰ ╱ ╲ ╳
// ┄ ┅ ┆ ┇ ┈ ┉ ┊ ┋╌ ╍ ╎ ╏
// ├ ┝ ┞ ┟ ┠ ┡ ┢ ┣ ┤ ┥ ┦ ┧ ┨ ┩ ┪ ┫ ┬ ┭ ┮ ┯ ┰ ┱ ┲ ┳ ┴ ┵ ┶ ┷ ┸ ┹ ┺ ┻ ┼ ┽ ┾ ┿ ╀ ╁ ╂ ╃ ╄ ╅ ╆ ╇ ╈ ╉ ╊ ╋
// ─ ━ ┃ ┄ ┅ ┆ ┇ ┈ ┉ ┊ ┋ ┌ ┍ ┎ ┏ ┐ ┑ ┒ ┓ └ ┕ ┖ ┗ ┘ ┙ ┚ ┛ ├ ┝ ┞ ┟ ┠ ┡ ┢ ┣ ┤ ┥ ┦ ┧ ┨ ┩ ┪ ┫ ┬ ┭ ┮ ┯ ┰ ┱ ┲ ┳ ┴ ┵ ┶ ┷ ┸ ┹ ┺ ┻ ┼ ┽ ┾ ┿

// Box is a terminal box.
type Box struct {
	Width   int
	Height  int
	Content strings.Builder
	Buffer  io.Writer
	Colors  []color.Attribute
}

const (
	DefaultWidth  = 120
	DefaultHeight = math.MaxInt
)

func NewBox(f *os.File, buf io.Writer, a ...color.Attribute) *Box {
	width, height, err := term.GetSize(int(f.Fd()))
	if err != nil {
		// If we can't get the terminal size, use defaults.
		log.Printf("could not get terminal size: %v, using defaults", err)
		width, height = DefaultWidth, DefaultHeight
	}
	return &Box{
		Width:   width,
		Height:  height,
		Buffer:  buf,
		Colors:  a,
		Content: strings.Builder{},
	}
}

func (b *Box) getClock() string {
	return time.Now().Format(time.RFC1123Z)
}

func (b *Box) wordWrap(text string, lineWidth int) []string {
	res := []string{}
	for len(text) > lineWidth {
		res = append(res, text[:lineWidth])
		text = text[lineWidth:]
	}
	if len(res) == 0 && len(text) > 0 {
		return []string{text}
	}
	return res
}

func (b *Box) String() string {
	buf := bytes.NewBufferString("")

	// Header
	clock := b.getClock()
	length := b.Width - 2 - len(clock) // Number of horizontal characters required to fill the line
	buf.WriteString(fmt.Sprintf("\r╭%s%s╮\n", strings.Repeat("─", length), clock))

	// Body
	for _, rawLine := range strings.Split(b.Content.String(), "\n") {
		for _, line := range b.wordWrap(rawLine, b.Width-2) {
			_, err := buf.WriteString(fmt.Sprintf("│%s%s│\n", line, strings.Repeat(" ", b.Width-2-len(line))))
			if err != nil {
				panic(err)
			}
		}
	}

	// Footer
	length = b.Width - 2 // Number of horizontal characters required to fill the line
	_, err := buf.WriteString(fmt.Sprintf("\r╰%s╯\n", strings.Repeat("─", length)))
	if err != nil {
		return buf.String()
	}

	return buf.String()
}

func (b *Box) Print() error {
	content := b.String()
	if len(content) == 0 {
		return nil
	}
	bs := []byte(color.New(b.Colors...).Sprintf("%s", content))
	n, err := os.Stdout.Write(bs)
	if err != nil {
		return err
	}
	if n != len(bs) {
		return fmt.Errorf("failed to write all bytes")
	}
	return nil
}

func (b *Box) Close() error {
	clock := b.getClock()
	// Number of horizontal characters required to fill the line
	length := b.Width - len(clock) - 2 // consider the two corners
	// Write the footer to the content buffer
	_, err := b.Content.WriteString(fmt.Sprintf("\r╰%s%s╯\n", strings.Repeat("─", length), clock))
	if err != nil {
		return err
	}
	return b.Print()
}

func getDelay(since time.Time) string {
	return time.Since(since).Truncate(100*time.Millisecond).String() + " "
}

// Spinner is a spinner within a box footer border.
func Spinner(isRunning func() bool, start time.Time, sleep time.Duration, width int, leftColor, rightColor *color.Color) {
	var delay string
	for isRunning() {
		// Move left to right
		for i := 0; isRunning() && i < width-len(delay)-8; i++ {
			delay = getDelay(start)
			leftColor.Fprintf(color.Output, "\r╰─%s─╯ %s%s─╯",
				strings.Repeat("─", i), leftColor.Sprint(delay), rightColor.Sprint(strings.Repeat("─", width-len(delay)-8-i)))
			time.Sleep(sleep)
		}

		// Move right to left
		for i := width - len(delay) - 8; isRunning() && i > 0; i-- {
			delay = getDelay(start)
			leftColor.Fprintf(color.Output, "\r╰─%s%s ├─%s─╯",
				strings.Repeat("─", i), leftColor.Sprint(delay), rightColor.Sprint(strings.Repeat("─", width-len(delay)-8-i)))
			time.Sleep(sleep)
		}
	}
}
