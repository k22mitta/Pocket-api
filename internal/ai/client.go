package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatMessage struct {
	Role    string
	Content string
}

type Client struct {
	inner *genai.Client
	model string
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("creating gemini client: %w", err)
	}
	return &Client{inner: c, model: "gemini-2.5-flash-lite"}, nil
}

func (c *Client) Close() {
	c.inner.Close()
}

func (c *Client) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) (string, error) {
	model := c.inner.GenerativeModel(c.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	var genHistory []*genai.Content
	for _, msg := range history {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		genHistory = append(genHistory, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	session := model.StartChat()
	session.History = genHistory

	var resp *genai.GenerateContentResponse
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = session.SendMessage(ctx, genai.Text(userMessage))
		if err == nil {
			break
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "429") || strings.Contains(msg, "quota") || strings.Contains(msg, "rate") ||
			strings.Contains(msg, "503") || strings.Contains(msg, "high demand") || strings.Contains(msg, "overloaded") ||
			strings.Contains(msg, "unavailable") || strings.Contains(msg, "500") {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return "", fmt.Errorf("sending message: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response part type")
	}

	return string(text), nil
}
