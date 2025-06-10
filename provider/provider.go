package provider

import (
	"context"
	"strings"

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

func (p *Provider) Generate(ctx context.Context, prompt string) (string, error) {
	r, err := p.client.Models.GenerateContent(ctx, *p.model, genai.Text(prompt), &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](1.0),
		// TopP:        0.2,
		// TopK:        1,
		// CandidateCount: 1,
		// ResponseMIMEType: "application/json",
	})
	if err != nil {
		return "", err
	}

	res := []string{}
	for i, c := range r.Candidates {
		lines := 0
		for _, p := range c.Content.Parts {
			for _, l := range strings.Split(p.Text, "\n") {
				res = append(res, l)
				lines += 1
			}
		}
		res = append(res, color.GreenString("│ End of response candidate: %d, Lines: %d", i+1, lines))
	}

	return strings.Join(res, "\n"), nil
}
