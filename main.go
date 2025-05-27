package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/urfave/cli/v2"
	"google.golang.org/genai"
	_ "google.golang.org/genai"
)

type GeminiConfig struct {
	APIKey string
	Model  string
}

// readPDFUsingGemini handles the pdf_reader tool request
func readPDFUsingGemini(config GeminiConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, ok := request.Params.Arguments["file_path"].(string)
		if !ok {
			return mcp.NewToolResultError("'file_path' must be a string"), nil
		}

		result, err := readPDF(config.APIKey, config.Model, filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error calling Perplexity API: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	}
}

func readPDF(apiKey, model, filePath string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	if stat.IsDir() {
		return "", fmt.Errorf("file is a directory")
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("file is a symlink")
	}

	if stat.Mode()&os.ModeNamedPipe != 0 {
		return "", fmt.Errorf("file is a named pipe")
	}

	if stat.Mode()&os.ModeDevice != 0 {
		return "", fmt.Errorf("file is a device")
	}

	if stat.Mode()&os.ModeSocket != 0 {
		return "", fmt.Errorf("file is a socket")
	}

	if stat.Size() == 0 {
		return "", fmt.Errorf("file is empty")
	}

	parts := genai.Text(`Convert the following PDF to markdown. Do not return any references to the PDF file, do not wrap in code blocks, or return any commentary. Return the markdown only.`)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	parts = append(parts, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{genai.NewPartFromBytes(fileData, "application/pdf")},
	})

	modelConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: "text/plain",
	}

	resp, generateErr := client.Models.GenerateContent(ctx, model, parts, modelConfig)
	if generateErr != nil {
		return "", generateErr
	}

	return resp.Text(), nil
}

// registerPDFReaderTool creates and registers the pdf_reader tool
func registerPDFReaderTool(s *server.MCPServer, config GeminiConfig) {
	pdfReaderTool := mcp.NewTool("pdf_reader",
		mcp.WithDescription("Reads a PDF file and returns the text content."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The path to the PDF file to read"),
		),
	)

	s.AddTool(pdfReaderTool, readPDFUsingGemini(config))
}

func main() {
	app := &cli.App{
		Name:  "pdf-reader",
		Usage: "A Model Context Protocol server for reading local PDFs using Gemini",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   "gemini-2.0-flash",
				Usage:   "The model to use for reading PDFs",
				EnvVars: []string{"GEMINI_MODEL"},
			},
			&cli.StringFlag{
				Name:     "api-key",
				Aliases:  []string{"k"},
				Usage:    "The API key to use for Gemini API requests",
				EnvVars:  []string{"GEMINI_API_KEY"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			// Create configuration from CLI arguments
			config := GeminiConfig{
				APIKey: c.String("api-key"),
				Model:  c.String("model"),
			}

			buildInfo, ok := debug.ReadBuildInfo()
			version := "v0.0.1"
			if ok {
				version = buildInfo.Main.Version
			}

			// Create a new MCP server
			s := server.NewMCPServer(
				"pdf-reader",
				version,
			)

			// Register tools
			registerPDFReaderTool(s, config)

			// Start the server
			if err := server.ServeStdio(s); err != nil {
				return cli.Exit(fmt.Sprintf("Server error: %v", err), 1)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
