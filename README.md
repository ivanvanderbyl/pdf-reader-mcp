# PDF Reader MCP Server

A Model Context Protocol (MCP) server for reading and analyzing PDF documents using Google's Gemini API, written in Go. This server enables AI assistants like Claude (Code and Desktop) and Cursor to seamlessly read, extract, and analyze PDF content directly from their interfaces.

## Description

The PDF Reader MCP Server acts as a bridge between AI assistants and PDF documents, allowing them to:

1. **Read and extract text from PDF files** using the `pdf_read` tool
2. **Analyze PDF content with Gemini's powerful vision capabilities** using the `pdf_analyze` tool
3. **Search within PDF documents** for specific information using the `pdf_search` tool

This integration lets AI assistants like Claude access and understand PDF content without leaving their interface, creating a seamless experience for document analysis and information extraction.

### Key Benefits

- **Direct PDF access**: Read and analyze PDF documents without manual conversion
- **Advanced AI analysis**: Leverage Gemini's vision and language models for deep document understanding
- **Text extraction**: Extract structured and unstructured text from PDFs
- **Content search**: Search for specific information within PDF documents
- **Seamless integration**: Works natively with Claude Code, Claude Desktop, and Cursor
- **Simple installation**: Quick setup with Homebrew, Go, or pre-built binaries

## Installation

### Using Homebrew (macOS and Linux)

```sh
brew tap alcova-ai/tap
brew install pdf-reader-mcp
```

### From Source

Clone the repository and build manually:

```sh
git clone https://github.com/Alcova-AI/pdf-reader-mcp.git
cd pdf-reader-mcp
go build -o pdf-reader-mcp-server .
```

### From Binary Releases (Other platforms)

Download pre-built binaries from the [releases page](https://github.com/Alcova-AI/pdf-reader-mcp/releases).

## Usage

This server supports only the `stdio` protocol for MCP communication.

### Setup with Claude Code

Adding to Claude Code:

```sh
claude mcp add-json --scope user pdf-reader-mcp '{"type":"stdio","command":"pdf-reader-mcp","env":{"GEMINI_API_KEY":"YOUR-GEMINI-API-KEY-HERE"}}'
```

That's it! You can now read and analyze PDFs in Claude Code.

### Setup with Claude Desktop

Adding to Claude Desktop:

1. Edit the Claude Desktop MCP config:

```sh
code ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

2. Add the PDF Reader MCP server:

```diff
  {
    "mcpServers": {
+        "pdf-reader-mcp": {
+            "command": "pdf-reader-mcp",
+            "args": [
+                "--model",
+                "gemini-1.5-flash"
+            ],
+            "env": {
+                "GEMINI_API_KEY": "YOUR-GEMINI-API-KEY-HERE"
+            }
+        }
    }
  }
```

### Command Line Options

- `--model, -m`: Specify the Gemini model to use for PDF analysis (default: "gemini-1.5-flash")
  - Can also be set with the `GEMINI_MODEL` environment variable
- `--max-pages`: Maximum number of pages to process at once (default: 10)
- `--temp-dir`: Directory for temporary file storage (default: system temp)

Example:

```sh
pdf-reader-mcp --model gemini-1.5-pro --max-pages 20
```

### Direct Execution

If you want to run the server directly (not recommended for most users):

1. Set your Gemini API key as an environment variable:

   ```sh
   export GEMINI_API_KEY=your-api-key-here
   ```

2. Run the server:

   ```sh
   pdf-reader-mcp
   ```

## License

MIT
