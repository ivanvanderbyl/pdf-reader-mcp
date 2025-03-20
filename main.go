package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/urfave/cli/v2"
)

const (
	apiURL = "https://api.perplexity.ai/chat/completions"
)

// Message represents a message in the chat completion request
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to the Perplexity API
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatCompletionResponse represents the response from the Perplexity API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Citations []string `json:"citations,omitempty"`
}

// performChatCompletion sends a request to the Perplexity API and returns the response
func performChatCompletion(apiKey string, model string, messages []Message) (string, error) {
	request := ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response ChatCompletionResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned in response")
	}

	// Get the message content from the response
	messageContent := response.Choices[0].Message.Content

	// Append citations to the message content if they exist
	if len(response.Citations) > 0 {
		messageContent += "\n\nCitations:\n"
		for i, citation := range response.Citations {
			messageContent += fmt.Sprintf("[%d] %s\n", i+1, citation)
		}
	}

	return messageContent, nil
}

func main() {
	app := &cli.App{
		Name:  "perplexity-mcp",
		Usage: "A Model Context Protocol server for Perplexity API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   "sonar-pro",
				Usage:   "The model to use for chat completions",
				EnvVars: []string{"PERPLEXITY_MODEL"},
			},
			&cli.StringFlag{
				Name:    "reasoning-model",
				Aliases: []string{"r"},
				Value:   "sonar-reasoning-pro",
				Usage:   "The model to use for reasoning tasks",
				EnvVars: []string{"PERPLEXITY_REASONING_MODEL"},
			},
			&cli.StringFlag{
				Name:     "api-key",
				Aliases:  []string{"k"},
				Usage:    "The API key to use for Perplexity API requests",
				EnvVars:  []string{"PERPLEXITY_API_KEY"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			// Get API key from environment variable
			apiKey := c.String("api-key")

			// Get model from command line argument or flag
			model := c.String("model")
			reasoningModel := c.String("reasoning-model")

			// Create a new MCP server using the mcp-go library
			s := server.NewMCPServer(
				"perplexity-mcp",
				"0.2.0",
			)

			// Add the Perplexity ask tool
			perplexityTool := mcp.NewTool("perplexity_ask",
				mcp.WithDescription("Engages in a conversation using the Perplexity to search and reason about the query. "+
					"Accepts an array of messages (each with a role and content) "+
					"and returns a chat completion response from the Perplexity model."),
				mcp.WithArray("messages",
					mcp.Required(),
					mcp.Description("Array of conversation messages"),
					mcp.Items(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"role": map[string]any{
								"type":        "string",
								"description": "Role of the message (e.g., system, user, assistant)",
							},
							"content": map[string]any{
								"type":        "string",
								"description": "The content of the message",
							},
						},
						"required": []string{"role", "content"},
					}),
				),
			)

			// Define the handler for the perplexity_ask tool
			s.AddTool(perplexityTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				// Extract messages from request
				messagesRaw, ok := request.Params.Arguments["messages"].([]any)
				if !ok {
					return mcp.NewToolResultError("'messages' must be an array"), nil
				}

				// Convert messages to the correct format
				var messages []Message
				for _, msgRaw := range messagesRaw {
					msgMap, ok := msgRaw.(map[string]any)
					if !ok {
						return mcp.NewToolResultError("Invalid message format"), nil
					}

					role, ok := msgMap["role"].(string)
					if !ok {
						return mcp.NewToolResultError("Message must have a 'role' field of type string"), nil
					}

					content, ok := msgMap["content"].(string)
					if !ok {
						return mcp.NewToolResultError("Message must have a 'content' field of type string"), nil
					}

					messages = append(messages, Message{Role: role, Content: content})
				}

				// Call the Perplexity API
				result, err := performChatCompletion(apiKey, model, messages)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error calling Perplexity API: %v", err)), nil
				}

				// Return the result
				return mcp.NewToolResultText(result), nil
			})

			// Add the Perplexity reasoning tool
			reasoningTool := mcp.NewTool("perplexity_reason",
				mcp.WithDescription("Uses the Perplexity reasoning model to perform complex reasoning tasks. "+
					"Accepts a query string and returns a comprehensive reasoned response."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("The query or problem to reason about"),
				),
			)

			// Define the handler for the perplexity_reason tool
			s.AddTool(reasoningTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				// Extract query from request
				query, ok := request.Params.Arguments["query"].(string)
				if !ok {
					return mcp.NewToolResultError("'query' must be a string"), nil
				}

				// Create a message for reasoning
				messages := []Message{
					{Role: "system", Content: "You are a reasoning assistant focused on solving complex problems through step-by-step reasoning."},
					{Role: "user", Content: query},
				}

				// Call the Perplexity API with the reasoning model
				result, err := performChatCompletion(apiKey, reasoningModel, messages)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error calling Perplexity API: %v", err)), nil
				}

				// Return the result
				return mcp.NewToolResultText(result), nil
			})


			// Start the server
			if err := server.ServeStdio(s); err != nil {
				return cli.Exit(fmt.Sprintf("Server error: %v", err), 1)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
