package provider

import (
	"context"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dberstein/cai/tui"
	"github.com/fatih/color"
	"google.golang.org/genai"
)

// Gemini defaults
const (
	DefaultModel       string  = "models/gemini-2.5-pro-preview-06-05"
	DefaultMimeType    string  = "text/plain"
	DefaultTopP        float64 = 0.2
	DefaultTopK        int     = 40
	DefaultTemperature float64 = 0.5
	DefaultMaxTokens   int     = 10240
	DefaultCandidates  int     = 1
)

// Gemini flags
var (
	GenerateContentModel          string
	GenerateContentTemperature    float64
	GenerateContentTopP           float64
	GenerateContentTopK           int
	GenerateContentMaxTokens      int
	GenerateContentCandidateCount int
	GenerateContentMimeType       string
)

func init() {
	flag.StringVar(&GenerateContentModel, "model", DefaultModel, "model to use")
	flag.StringVar(&GenerateContentMimeType, "mime", DefaultMimeType, "mimeType to use; ie. text/plain, application/json, etc.")
	flag.Float64Var(&GenerateContentTemperature, "temperature", DefaultTemperature, "temperature to use")
	flag.Float64Var(&GenerateContentTopP, "topp", DefaultTopP, "top_p to use")
	flag.IntVar(&GenerateContentTopK, "topk", DefaultTopK, "top_k to use")
	flag.IntVar(&GenerateContentMaxTokens, "tokens", DefaultMaxTokens, "max_tokens to use")
	flag.IntVar(&GenerateContentCandidateCount, "candidates", DefaultCandidates, "candidate_count to use")
}

// Provides a new instance of the provider for the given model
func NewGemini(ctx context.Context, clientConfig *genai.ClientConfig) (Provider, error) {
	if clientConfig == nil {
		clientConfig = &genai.ClientConfig{Backend: genai.BackendGeminiAPI}
	}
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, err
	}
	return &Gemini{model: GenerateContentModel, client: client}, nil
}

// Gemini provider
type Gemini struct {
	Provider
	model  string
	client *genai.Client
}

// CurrentModel returns the current model
func (p *Gemini) CurrentModel() string {
	return p.model
}

// Close closes the provider
func (p *Gemini) Close() error {
	return nil
}

// ListModels lists the available models
func (p *Gemini) ListModels(ctx context.Context) ([]string, error) {
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

// Generate generates content from the prompt
func (p *Gemini) Generate(ctx context.Context, prompt string, w io.Writer) error {
	topP := float32(GenerateContentTopP)
	topK := float32(GenerateContentTopK)
	temperature := float32(GenerateContentTemperature)
	generateContentConfig := &genai.GenerateContentConfig{
		CandidateCount:   int32(GenerateContentCandidateCount),
		MaxOutputTokens:  int32(GenerateContentMaxTokens),
		ResponseMIMEType: GenerateContentMimeType,
		Temperature:      &temperature,
		TopP:             &topP,
		TopK:             &topK,
		SafetySettings:   nil,
		StopSequences:    nil,
		Tools:            nil,
	}

	start := time.Now()

	// todo: research p.client.Chats
	res, err := p.client.Models.GenerateContent(ctx, p.model, genai.Text(prompt), generateContentConfig)
	if err != nil {
		return err
	}

	// Response can have multiple candidates
	for i, c := range res.Candidates {
		for _, prt := range c.Content.Parts {
			_, err := w.Write([]byte(prt.Text))
			if err != nil {
				return err
			}

			// Summary box for the candidate
			inner := strings.Split(prt.Text, "\n")
			_, err = p.summary(w, i, inner, start)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

// summary writes a summary box for the candidate to the given writer
func (p *Gemini) summary(w io.Writer, i int, inner []string, start time.Time) (*tui.Box, error) {
	bx := tui.NewBox(os.Stdout, w, color.FgHiMagenta, color.Bold)
	_, err := bx.Content.WriteString(strings.Join([]string{
		"Candidate #" + strconv.Itoa(i+1) + " / " + strconv.Itoa(len(inner)) + " line(s)",
		"Delay: " + time.Since(start).Truncate(100*time.Millisecond).String(),
	}, " ┆ "))
	if err != nil {
		return bx, err
	}

	// err = bx.Close()
	err = bx.Print()
	if err != nil {
		return bx, err
	}

	return bx, nil
}
