package chat

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/chzyer/readline"
	"github.com/fatih/color"

	"github.com/dberstein/cai/content"
	"github.com/dberstein/cai/pager"
	"github.com/dberstein/cai/provider"
	"github.com/dberstein/cai/request"
	"github.com/dberstein/cai/response"
	"github.com/dberstein/cai/spinner"
	"github.com/dberstein/cai/tui"
)

const PromptStart = " Ctrl+D:EOF> "
const PromptContinuation = ".. "
const DefaultHistoryFile = ".cai_history"

// New returns a new conversation for the given provider.
func New(provider provider.Provider) *C {
	var historyFile string
	if homeDir, err := os.UserHomeDir(); err == nil {
		historyFile = filepath.Join(homeDir, DefaultHistoryFile) // todo: make configurable
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine home directory. History will not be saved: %v\n", err)
	}

	rl, err := initReadline(historyFile)
	if err != nil {
		panic(fmt.Errorf("error initializing readline: %w", err))
	}

	return &C{
		rl:       rl,
		provider: provider,
	}
}

func initReadline(historyFile string) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:              color.CyanString(PromptStart),
		HistoryFile:         historyFile,
		HistorySearchFold:   true,
		UniqueEditLine:      false,
		InterruptPrompt:     "^C",
		EOFPrompt:           "^D",
		ForceUseInteractive: false,
		VimMode:             false,
		AutoComplete: readline.NewPrefixCompleter(readline.PcItemDynamic(func(string) []string {
			return []string{} // return []string{"load", "save"}
		})),
	})
}

// C is a chat/conversation with an AI provider.
type C struct {
	rl       *readline.Instance
	provider provider.Provider
}

// Close closes the conversation.
func (c *C) Close() error {
	if err := c.rl.Close(); err != nil {
		return err
	}
	if err := c.provider.Close(); err != nil {
		return err
	}
	return nil
}

// Submit submits a user request to the provider and returns the response.
func (c *C) Submit(bs *bytes.Buffer, request *request.Request) *response.Response {
	width := tui.NewBox(os.Stdout, bs).Width
	err := spinner.New(width).
		// Wrap(os.Stdout, func() error {
		// 	bs.WriteString("hello!") // todo: remove and uncomment following
		Wrap(os.Stdout, func() error {
			// TODO: err := prompt.Augment(request.Content)
			err := c.provider.Generate(context.Background(), request.Content.String(), bs)
			if err != nil {
				return err
			}
			return nil
		}, color.FgHiGreen)

	return response.New(bs, err)
}

// Iter returns an iterator of request.Request(s).
func (c *C) Iter(bs *bytes.Buffer, model string) iter.Seq[*request.Request] {
	return func(yield func(*request.Request) bool) {
		for {
			// Denote new request input.
			c.rl.SetPrompt(color.New(color.FgHiYellow).Sprintf("@%s%s", model, color.CyanString(PromptStart)))
			r := request.New()
		ContinuationLoop:
			for {
				line := c.rl.Line()
				switch line.Error {
				case nil:
					r.Content.Append(&line.Line)
				// In this context, io.EOF means end of input for this request.
				case io.EOF:
					fmt.Printf("\x1b[%dA", 1) // move cursor up one line
					break ContinuationLoop
				// Unexpected error
				default:
					log.Fatalf("Error reading line: %v\n", line.Error)
				}
				// Denote request input continuation.
				c.rl.SetPrompt(PromptContinuation)
			}
			// Denote end of request with request box.
			buf := bytes.NewBufferString("")
			bx := tui.NewBox(os.Stdout, buf, color.FgHiGreen)
			text := r.Content.String()
			bx.Content.WriteString(text)
			bx.Print()

			// Yield the new request.
			if !yield(r) {
				return
			}
		}
	}
}

func (c *C) Run() error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(readline.GetScreenWidth()),
	)
	if err != nil {
		return err
	}

	// Iterate over input requests
	bs := bytes.NewBufferString("")
	for req := range c.Iter(bs, c.provider.CurrentModel()) {
		// Handle macros amd commands in input
		// TODO: implement

		// Submit to provider
		res := c.Submit(bs, req)
		err := res.Error()
		if err != nil {
			return err
		}

		// Process response
		err = res.Process()
		if err != nil {
			return err
		}

		// Render response
		rmd, err := renderer.Render(res.Content.String())
		if err != nil {
			return err
		}

		// Invoke paginator
		if err := pager.New().WriteFunc(content.NewString(&rmd)); err != nil {
			log.Printf("Error writing to pager: %v\n", err)
			// fmt.Errorf("error writing to pager: %w", err)
			// return
		}

		bs.Reset()
	}

	return nil
}
