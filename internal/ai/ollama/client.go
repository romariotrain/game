package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"solo-leveling/internal/models"
)

const (
	DefaultBaseURL = "http://localhost:11434"
	DefaultModel   = "qwen2.5:3b-instruct"
)

type Client struct {
	BaseURL        string
	Model          string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
	WarmupTimeout  time.Duration
}

type generateRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"`
	Options generateOptions `json:"options"`
}

type generateOptions struct {
	NumCtx      int     `json:"num_ctx"`
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error"`
}

func NewClient() *Client {
	return &Client{
		BaseURL:        DefaultBaseURL,
		Model:          DefaultModel,
		RequestTimeout: 120 * time.Second,
		WarmupTimeout:  180 * time.Second,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) GenerateSuggestions(ctx context.Context, profileText string, stats models.PlayerStats) ([]models.AISuggestion, string, error) {
	c.ensureDefaults()

	prompt := buildPrompt(profileText, stats)
	reqCtx, cancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer cancel()
	raw, err := c.generateRaw(reqCtx, prompt)
	if err != nil {
		return nil, "", err
	}

	suggestions, parseErr := ParseSuggestionsJSON(raw)
	if parseErr == nil {
		normalized := normalizeSuggestions(suggestions)
		if len(normalized) >= 3 {
			if len(normalized) > 5 {
				normalized = normalized[:5]
			}
			return normalized, raw, nil
		}
	}

	repairPrompt := buildRepairPrompt(raw)
	repairCtx, repairCancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer repairCancel()
	repairedRaw, err := c.generateRaw(repairCtx, repairPrompt)
	if err != nil {
		// If repair fails but we had some valid suggestions from first attempt, use them
		if parseErr == nil {
			normalized := normalizeSuggestions(suggestions)
			if len(normalized) > 0 {
				if len(normalized) > 5 {
					normalized = normalized[:5]
				}
				return normalized, raw, nil
			}
		}
		return nil, raw, fmt.Errorf("parse failed and repair failed: %w", parseErr)
	}
	suggestions, parseErr = ParseSuggestionsJSON(repairedRaw)
	if parseErr != nil {
		return nil, repairedRaw, fmt.Errorf("parse failed after repair: %w", parseErr)
	}

	normalized := normalizeSuggestions(suggestions)
	if len(normalized) > 5 {
		normalized = normalized[:5]
	}
	if len(normalized) == 0 {
		return nil, repairedRaw, fmt.Errorf("no valid suggestions after repair")
	}
	return normalized, repairedRaw, nil
}

func (c *Client) Warmup(ctx context.Context) error {
	c.ensureDefaults()

	warmCtx, cancel := context.WithTimeout(ctx, c.WarmupTimeout)
	defer cancel()
	_, err := c.generateRawWithOptions(warmCtx, "Скажи OK", generateOptions{
		NumCtx:      512,
		NumPredict:  16,
		Temperature: 0.0,
	})
	return err
}

func (c *Client) ensureDefaults() {
	if strings.TrimSpace(c.BaseURL) == "" {
		c.BaseURL = DefaultBaseURL
	}
	if strings.TrimSpace(c.Model) == "" {
		c.Model = DefaultModel
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 120 * time.Second
	}
	if c.WarmupTimeout <= 0 {
		c.WarmupTimeout = 180 * time.Second
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 120 * time.Second}
		return
	}
	if c.HTTPClient.Timeout <= 0 {
		c.HTTPClient.Timeout = 120 * time.Second
	}
}

func (c *Client) generateRaw(ctx context.Context, prompt string) (string, error) {
	return c.generateRawWithOptions(ctx, prompt, defaultGenerateOptions())
}

func defaultGenerateOptions() generateOptions {
	return generateOptions{
		NumCtx:      2048,
		NumPredict:  256,
		Temperature: 0.3,
	}
}

func (c *Client) generateRawWithOptions(ctx context.Context, prompt string, opts generateOptions) (string, error) {
	reqBody := generateRequest{
		Model:   c.Model,
		Prompt:  prompt,
		Stream:  true,
		Options: opts,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	apiURL := strings.TrimRight(c.BaseURL, "/") + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Use a client without global timeout — context handles cancellation.
	// http.Client.Timeout counts total time including body read, which kills
	// streaming responses that trickle in token-by-token on CPU.
	streamClient := &http.Client{Timeout: 0}
	resp, err := streamClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var parsed generateResponse
		if json.Unmarshal(body, &parsed) == nil && strings.TrimSpace(parsed.Error) != "" {
			return "", fmt.Errorf("ollama http %d: %s", resp.StatusCode, parsed.Error)
		}
		return "", fmt.Errorf("ollama http %d", resp.StatusCode)
	}

	// Read streaming NDJSON: each line is {"response":"token","done":false/true}
	var result strings.Builder
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk generateResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			// Context cancelled
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "", fmt.Errorf("decode ollama stream: %w", err)
		}
		if strings.TrimSpace(chunk.Error) != "" {
			return "", fmt.Errorf("ollama error: %s", chunk.Error)
		}
		result.WriteString(chunk.Response)
		if chunk.Done {
			break
		}
	}

	text := result.String()
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty ollama response")
	}
	return text, nil
}

func buildPrompt(profileText string, stats models.PlayerStats) string {
	profileText = truncateProfile(profileText, 2500)
	return fmt.Sprintf(
		"Ты — генератор идей задач для TODO-RPG. Верни ТОЛЬКО JSON массив из 5 элементов. Без текста вокруг. Не оценивай ранги. Формат:\n{\"title\":\"...\",\"desc\":\"...\",\"minutes\":5-60,\"effort\":1-5,\"friction\":1-3,\"stat\":\"STR|AGI|INT|STA\",\"tags\":[\"work|health|home|learning|social\"]}\nПрофиль: <<<%s>>>\nСтаты: STR=%d AGI=%d INT=%d STA=%d\nОграничения: ~45 мин в день, макс 60 мин на задачу.",
		profileText,
		stats.STR, stats.AGI, stats.INT, stats.STA,
	)
}

func buildRepairPrompt(raw string) string {
	return fmt.Sprintf(
		"Исправь JSON. Верни ТОЛЬКО корректный JSON массив из 5 элементов в том же формате. Без пояснений.\nИсходный текст:\n%s",
		raw,
	)
}

func normalizeSuggestions(in []models.AISuggestion) []models.AISuggestion {
	out := make([]models.AISuggestion, 0, len(in))
	for _, s := range in {
		s.Title = strings.TrimSpace(s.Title)
		s.Desc = strings.TrimSpace(strings.ReplaceAll(s.Desc, "\n", " "))
		s.Stat = strings.ToUpper(strings.TrimSpace(s.Stat))
		if s.Minutes < 5 {
			s.Minutes = 5
		}
		if s.Minutes > 60 {
			s.Minutes = 60
		}
		if s.Effort < 1 {
			s.Effort = 1
		}
		if s.Effort > 5 {
			s.Effort = 5
		}
		if s.Friction < 1 {
			s.Friction = 1
		}
		if s.Friction > 3 {
			s.Friction = 3
		}
		if s.Title == "" || s.Desc == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func truncateProfile(text string, maxChars int) string {
	if maxChars <= 0 {
		return strings.TrimSpace(text)
	}
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) <= maxChars {
		return trimmed
	}
	return string(runes[:maxChars])
}
