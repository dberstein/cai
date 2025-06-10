package box

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type Box struct {
	Width   int
	Height  int
	Content strings.Builder
	Output  io.Writer
	Colors  []color.Attribute
}

func New(f *os.File, a ...color.Attribute) *Box {
	width, height, err := term.GetSize(int(f.Fd()))
	if err != nil {
		// If we can't get the terminal size, fall back to printing directly.
		log.Printf("could not get terminal size: %v, printing directly", err)
		return nil
	}
	b := &Box{
		Width:   width,
		Height:  height,
		Output:  f,
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

func (b Box) String() string {
	sb := strings.Builder{}
	sb.WriteString("╭" + strings.Repeat("─", b.Width-2) + "╮\n")
	for _, line := range strings.Split(b.Content.String(), "\n") {
		if line == "" {
			continue
		}
		sb.WriteString("│" + line + strings.Repeat(" ", b.Width-len(line)-2) + "│\n")
	}
	// sb.Write([]byte("├" + strings.Repeat("─", b.Width-2) + "┤\n"))
	return sb.String()
}

func (b Box) Print(c ...color.Attribute) error {
	bs := []byte(color.New(c...).Sprintf("%s", b.String()))
	n, err := b.Output.Write(bs)
	if err != nil {
		return err
	}
	if n != len(bs) {
		return fmt.Errorf("failed to write all bytes")
	}
	return nil
}
