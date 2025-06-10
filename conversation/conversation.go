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
	"strings"

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

const PromptStart = "> "
const PromptContinuation = " .. "

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
		Prompt:            PromptStart,
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
	if request.Content == nil {
		return response.New(nil, fmt.Errorf("request content is nil"))
	}
	s := spinner.New(box.New(os.Stderr).Width)
	body := ""
	err := s.Wrap(c.rl.Stdout(), func() error {
		// TODO: err := prompt.Augment(request.Content)
		b, err := c.provider.Generate(context.Background(), request.Content.String())
		if err != nil {
			return err
		}
		body = b
		return nil
	}, color.FgHiGreen)

	return response.New(content.NewString(&body), err)
}

// Iter returns an iterator of request.Request(s).
func (c *C) Iter(bs *bytes.Buffer, model *string) iter.Seq[*request.Request] {
	return func(yield func(*request.Request) bool) {
		for {
			// Denote new request input.
			c.rl.SetPrompt(fmt.Sprintf("@%s %s", *model, PromptStart))
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
			s := r.Content.String()
			if len(s) > 0 {
				c.Separator(model, s) // Denote end of request input.
			}

			// Denote new request input.
			c.rl.SetPrompt(fmt.Sprintf("%s", PromptStart))

			// if len(strings.TrimSpace(r.Content.String())) > 0 {
			// 	// break
			// }

			if !yield(r) {
				return
			}
		}
	}
}

func (c *C) Run(model *string) error {
	pager := pager.New()
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	bs := bytes.NewBuffer(nil)
	printer := pager.WriteFunc()

	// iterate over input requests
	for req := range c.Iter(bs, model) {
		// 1. handle macros amd commands in input
		// TODO: implement

		// 2. submit to provider
		res := c.Submit(bs, req)
		err := res.Error()
		if err != nil {
			return err
		}

		// 3. process response
		err = res.Process()
		if err != nil {
			return err
		}

		// 4. display response
		for _, ln := range strings.Split(res.Content.String(), "\n") {
			rmd, err := renderer.Render(ln)
			if err != nil {
				return err
			}
			_ = rmd
			// ln = rmd
			cc := content.NewString(&ln)
			cc.Append(&NL)
			printer(cc)
		}
	}

	return nil
}

// Separator prints a separator line.
func (c *C) Separator(model *string, s string) {
	bx := box.New(os.Stderr)
	bx.Output = c.rl.Stdout()
	bx.Content.Write([]byte(s))
	// bx.Print()
	// bx.Content.Write([]byte("xyz\n123\n"))
	err := bx.Print(color.FgHiGreen)
	if err != nil {
		log.Fatalf("Error printing box: %v\n", err)
	}
	// os.Exit(0)

	// color.New(color.FgHiWhite, color.Italic, color.ReverseVideo).
	// 	Fprintf(os.Stdout,
	// 		"__________%s: model: %s|__________\n",
	// 		time.Now().Format(time.RFC3339),
	// 		*model,
	// 	)
}
