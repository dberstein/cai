package box

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type Box struct {
	Width   int
	Height  int
	Content strings.Builder
	Output  io.Writer
	Buffer  io.Writer
	Colors  []color.Attribute
}

func New(f *os.File, buf io.Writer, a ...color.Attribute) *Box {
	width, height, err := term.GetSize(int(f.Fd()))
	if err != nil {
		// If we can't get the terminal size, fall back to printing directly.
		log.Printf("could not get terminal size: %v, printing directly", err)
		return nil
	}
	b := &Box{
		Width:   width,
		Height:  height,
		Buffer:  buf,
		Colors:  a,
		Content: strings.Builder{},
	}

	return b
}

// ═ ║ ╒ ╓ ╔ ╕ ╖ ╗ ╘ ╙ ╚ ╛ ╜ ╝ ╞ ╟ ╠ ╡ ╢ ╣ ╤ ╥ ╦ ╧ ╨ ╩ ╪ ╫ ╬
// ┌ ┍ ┎ ┏ ┐ ┑ ┒ ┓ └ ┕ ┖ ┗ ┘ ┙ ┚ ┛
// ╭ ╮ ╯ ╰ ╱ ╲ ╳
// ┄ ┅ ┆ ┇ ┈ ┉ ┊ ┋╌ ╍ ╎ ╏
// ├ ┝ ┞ ┟ ┠ ┡ ┢ ┣ ┤ ┥ ┦ ┧ ┨ ┩ ┪ ┫ ┬ ┭ ┮ ┯ ┰ ┱ ┲ ┳ ┴ ┵ ┶ ┷ ┸ ┹ ┺ ┻ ┼ ┽ ┾ ┿ ╀ ╁ ╂ ╃ ╄ ╅ ╆ ╇ ╈ ╉ ╊ ╋
// ─ ━ ┃ ┄ ┅ ┆ ┇ ┈ ┉ ┊ ┋ ┌ ┍ ┎ ┏ ┐ ┑ ┒ ┓ └ ┕ ┖ ┗ ┘ ┙ ┚ ┛ ├ ┝ ┞ ┟ ┠ ┡ ┢ ┣ ┤ ┥ ┦ ┧ ┨ ┩ ┪ ┫ ┬ ┭ ┮ ┯ ┰ ┱ ┲ ┳ ┴ ┵ ┶ ┷ ┸ ┹ ┺ ┻ ┼ ┽ ┾ ┿

func (b *Box) String() string {
	w := bytes.NewBufferString("")
	w.WriteString("╭" + strings.Repeat("─", b.Width-2) + "╮\n") // todo: top continuation "┘  └"
	for _, line := range strings.Split(b.Content.String(), "\n") {
		ll := len(line)
		width := b.Width - ll - 1
		if ll > 0 && width > -10 {
			w.WriteString("│" + line + strings.Repeat(" ", width) + "│\n")
		} else if ll > 0 {
			w.WriteString(line + "\n")
		}
	}

	clock := time.Now().Format(time.RFC1123Z)
	length := b.Width - 2 - len(clock) // Number of horizontal characters required to fill the line
	_, err := w.WriteString("\r╰" + strings.Repeat("─", length) + clock + "╯")
	if err != nil {
		return ""
	}

	return w.String()
}

func (b *Box) Close() error {
	clock := time.Now().Format(time.RFC1123Z)
	length := b.Width - 2 - len(clock) // Number of horizontal characters required to fill the line
	_, err := b.Content.WriteString("\r╰" + strings.Repeat("─", length) + clock + "╯\n")
	if err != nil {
		return err
	}
	return b.Print()
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
