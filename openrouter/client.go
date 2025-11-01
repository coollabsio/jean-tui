package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	model   string
	baseURL string
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type CommitMessage struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewClient creates a new OpenRouter API client
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "anthropic/claude-3.5-haiku"
	}
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://openrouter.ai/api/v1",
	}
}

// GenerateCommitMessage generates a conventional commit message based on git diff
// If customPrompt is empty, uses the default prompt
func (c *Client) GenerateCommitMessage(diff, customPrompt string) (subject, body string, err error) {
	if c.apiKey == "" {
		return "", "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Limit diff to reasonable size to avoid token limits
	if len(diff) > 5000 {
		diff = diff[:5000]
	}

	// Use custom prompt if provided, otherwise use default
	prompt := customPrompt
	if prompt == "" {
		prompt = DefaultCommitPrompt
	}
	// Replace {diff} placeholder with actual diff
	prompt = strings.ReplaceAll(prompt, "{diff}", diff)

	response, err := c.callAPI(prompt)
	if err != nil {
		return "", "", err
	}

	// Parse JSON response
	var msg CommitMessage
	if err := json.Unmarshal([]byte(response), &msg); err != nil {
		return "", "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate subject
	msg.Subject = strings.TrimSpace(msg.Subject)
	if msg.Subject == "" {
		return "", "", fmt.Errorf("AI generated empty commit subject")
	}
	if len(msg.Subject) > 72 {
		msg.Subject = msg.Subject[:72]
	}

	// Clean up body
	msg.Body = strings.TrimSpace(msg.Body)
	if len(msg.Body) > 500 {
		msg.Body = msg.Body[:500]
	}

	return msg.Subject, msg.Body, nil
}

// GenerateBranchName generates a semantic branch name based on git diff
// If customPrompt is empty, uses the default prompt
func (c *Client) GenerateBranchName(diff, customPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Limit diff to reasonable size
	if len(diff) > 3000 {
		diff = diff[:3000]
	}

	// Use custom prompt if provided, otherwise use default
	prompt := customPrompt
	if prompt == "" {
		prompt = DefaultBranchNamePrompt
	}
	// Replace {diff} placeholder with actual diff
	prompt = strings.ReplaceAll(prompt, "{diff}", diff)

	name, err := c.callAPI(prompt)
	if err != nil {
		return "", err
	}

	// Clean up response
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove non-alphanumeric except hyphens
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	name = string(result)

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Limit to 40 chars
	if len(name) > 40 {
		name = name[:40]
	}

	if name == "" {
		return "", fmt.Errorf("AI generated invalid branch name")
	}

	return name, nil
}

// GeneratePRContent generates a PR title and description from a git diff
// If customPrompt is empty, uses the default prompt
func (c *Client) GeneratePRContent(diff, customPrompt string) (title, description string, err error) {
	if c.apiKey == "" {
		return "", "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Limit diff to reasonable size
	if len(diff) > 5000 {
		diff = diff[:5000]
	}

	// Use custom prompt if provided, otherwise use default
	prompt := customPrompt
	if prompt == "" {
		prompt = DefaultPRPrompt
	}
	// Replace {diff} placeholder with actual diff
	prompt = strings.ReplaceAll(prompt, "{diff}", diff)

	response, err := c.callAPI(prompt)
	if err != nil {
		return "", "", err
	}

	// Parse response in format "TITLE: ... \n DESCRIPTION: ..."
	lines := strings.Split(response, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TITLE:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
			// Continue looking for description
			for j := i + 1; j < len(lines); j++ {
				descLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(descLine, "DESCRIPTION:") {
					description = strings.TrimSpace(strings.TrimPrefix(descLine, "DESCRIPTION:"))
					break
				}
			}
			break
		}
	}

	// Validate we got both fields
	if title == "" {
		return "", "", fmt.Errorf("failed to parse PR title from AI response")
	}
	if description == "" {
		description = "" // Description is optional
	}

	// Limit title to 72 characters
	if len(title) > 72 {
		title = title[:72]
	}

	return title, description, nil
}

// callAPI makes a request to the OpenRouter API
func (c *Client) callAPI(prompt string) (string, error) {
	req := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Low temperature for deterministic output
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	// Check for HTTP error status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	// Clean up response: remove markdown code block formatting if present
	content := chatResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Remove markdown code block delimiters (```json ... ``` or ``` ... ```)
	if strings.HasPrefix(content, "```") {
		// Remove opening ``` with optional language specifier
		if idx := strings.Index(content, "\n"); idx != -1 {
			content = content[idx+1:]
		} else {
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
		}
		// Remove closing ```
		if strings.HasSuffix(content, "```") {
			content = strings.TrimSuffix(content, "```")
		}
		content = strings.TrimSpace(content)
	}

	return content, nil
}
