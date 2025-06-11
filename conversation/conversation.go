package conversation

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

	"github.com/dberstein/cai/box"
	"github.com/dberstein/cai/content"
	"github.com/dberstein/cai/pager"
	"github.com/dberstein/cai/provider"
	"github.com/dberstein/cai/request"
	"github.com/dberstein/cai/response"
	"github.com/dberstein/cai/spinner"
)

const PromptStart = " Ctrl+D:EOF> "
const PromptContinuation = ".. "

// New returns a new conversation for the given provider.
func New(provider *provider.Provider) *C {
	historyFile := ""
	homeDir, err := os.UserHomeDir()
	if err == nil {
		historyFile = filepath.Join(homeDir, ".go_cli_multiline_history") // todo: make configurable
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine home directory. History will not be saved: %v\n", err)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            color.CyanString(PromptStart),
		HistoryFile:       historyFile,
		HistorySearchFold: true,
		// UniqueEditLine:    true,
		InterruptPrompt:     "^C",
		EOFPrompt:           "^D",
		ForceUseInteractive: true,
		// VimMode:             true,
		AutoComplete: readline.NewPrefixCompleter(readline.PcItemDynamic(func(string) []string {
			return []string{"load", "save"}
		})),
	})
	if err != nil {
		panic(fmt.Errorf("error initializing readline: %w", err))
	}

	return &C{
		rl:       rl,
		provider: provider,
		// prompt:   prompt.New(),
		// history:  make(map[time.Time]Interaction),
	}
}

type C struct {
	rl       *readline.Instance
	provider *provider.Provider
	// prompt   *prompt.Prompt
	// history  map[time.Time]Interaction
}

// NL is a newline character.
var NL = "\n"

// Close closes the conversation.
func (c *C) Close() {
	c.rl.Close()
	c.provider.Close()
}

// Submit submits a user request to the provider and returns the response.
func (c *C) Submit(bs *bytes.Buffer, request *request.Request) *response.Response {
	s := spinner.New(box.New(os.Stdout, nil).Width)
	err := s.Wrap(c.rl.Stdout(), func() error {
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
func (c *C) Iter(bs *bytes.Buffer, model *string) iter.Seq[*request.Request] {
	return func(yield func(*request.Request) bool) {
		for {
			// Denote new request input.
			c.rl.SetPrompt(color.New(color.FgHiYellow).Sprintf("@%s%s", *model, color.CyanString(PromptStart)))
			r := request.New()

		ContinuationLoop:
			for {
				line := c.rl.Line()
				switch line.Error {
				case nil: // Append line with trailing new line not captured by readline
					r.Content.Append(&line.Line)
					r.Content.Append(&NL)
				case io.EOF: // In this context, io.EOF means end of input for this request.
					fmt.Printf("\x1b[%dA", 1) // move cursor up one line
					break ContinuationLoop
				default: // Unexpected error
					log.Fatalf("Error reading line: %v\n", line.Error)
				}

				// Denote request input continuation.
				c.rl.SetPrompt(PromptContinuation)
			}

			// Denote end of request input.
			buf := bytes.NewBufferString("")
			bx := box.New(os.Stdout, buf, color.FgHiGreen)
			text := r.Content.String()
			bx.Content.WriteString(text)
			err := bx.Close()
			if err != nil {
				log.Fatalf("Error printing box: %v\n", err)
			}
			io.Copy(c.rl.Stdout(), buf)

			// Yield the request.
			if !yield(r) {
				return
			}
		}
	}
}

func (c *C) Run(model *string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	pager := pager.New()
	bs := bytes.NewBuffer(nil)

	// iterate over input requests
	for req := range c.Iter(bs, model) {
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

		// Display response
		rmd, err := renderer.Render(res.Content.String())
		if err != nil {
			return err
		}

		// Invoke paginator
		pager.WriteFunc(content.NewString(&rmd))
	}

	return nil
}
