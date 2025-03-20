package main

import (
	"os"
	"testing"
)

func TestPerformChatCompletion(t *testing.T) {
	// Skip this test if no API key is provided
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: PERPLEXITY_API_KEY environment variable not set")
	}

	// Test message
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "What is the capital of France?"},
	}

	// Test with default model
	result, err := performChatCompletion(apiKey, "sonar-pro", messages)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Fatalf("Expected non-empty result, got empty string")
	}

	// Additional checks can be added here based on expected response format
	t.Logf("API Response: %s", result)
}

func TestPerformReasoning(t *testing.T) {
	// Skip this test if no API key is provided
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: PERPLEXITY_API_KEY environment variable not set")
	}

	// Test message for reasoning task
	messages := []Message{
		{Role: "system", Content: "You are a reasoning assistant focused on solving complex problems through step-by-step reasoning."},
		{Role: "user", Content: "If a train travels at 120 km/h and another train travels at 80 km/h in the opposite direction, how long will it take for them to be 500 km apart if they start at the same location?"},
	}

	// Test with reasoning model
	result, err := performChatCompletion(apiKey, "sonar-reasoning-pro", messages)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Fatalf("Expected non-empty result, got empty string")
	}

	// Additional checks can be added here based on expected response format
	t.Logf("Reasoning API Response: %s", result)
}
