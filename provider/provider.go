package provider

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dberstein/cai/box"
	"github.com/fatih/color"
	"google.golang.org/genai"
)

// Provider represents Gemini LLM provider.
type Provider struct {
	model  *string
	client *genai.Client
}

// Provides a new instance of the provider for the given model
func New(model *string) (*Provider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{Backend: genai.BackendGeminiAPI})
	if err != nil {
		return nil, err
	}

	return &Provider{client: client, model: model}, nil
}

func (p *Provider) Close() error {
	return nil
}

func (p *Provider) ListModels(ctx context.Context) ([]string, error) {
	res, err := p.client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	var models []string
	for _, m := range res.Items {
		models = append(models, m.Name)
	}
	return models, nil
}

func (p *Provider) Generate(ctx context.Context, prompt string, w io.Writer) error {
	start := time.Now()
	r, err := p.client.Models.GenerateContent(ctx, *p.model, genai.Text(prompt), &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](0.2),
		// TopP:        0.2,
		// TopK:        1,
		CandidateCount: 1,
		// ResponseMIMEType: "application/json",
	})
	if err != nil {
		return err
	}

	for i, c := range r.Candidates {
		lines := []string{}
		for _, p := range c.Content.Parts {
			inner := strings.Split(p.Text, "\n")
			lines = append(lines, inner...)

			_, err := w.Write([]byte(p.Text))
			if err != nil {
				return err
			}

			// Summary box for the candidate
			bx := box.New(os.Stdout, w, color.FgHiMagenta, color.Bold)
			_, err = bx.Content.WriteString(strings.Join([]string{
				"Candidate #" + strconv.Itoa(i+1) + " / " + strconv.Itoa(len(inner)) + " line(s)",
				"Delay: " + time.Since(start).Truncate(100*time.Millisecond).String(),
			}, " ┆ "))
			if err != nil {
				return err
			}

			// err = bx.Close()
			err = bx.Print()
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}
