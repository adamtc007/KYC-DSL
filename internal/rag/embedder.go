package rag

import (
	"context"
	"fmt"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/adamtc007/KYC-DSL/internal/model"
)

// Embedder handles generation of vector embeddings using OpenAI API
type Embedder struct {
	client     *openai.Client
	model      openai.EmbeddingModel
	maxRetries int
	retryDelay time.Duration
	dimensions int
}

// EmbedderConfig configures the embedder
type EmbedderConfig struct {
	APIKey     string
	Model      openai.EmbeddingModel
	MaxRetries int
	RetryDelay time.Duration
}

// NewEmbedder creates a new embedder with OpenAI client
func NewEmbedder() *Embedder {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable not set")
	}

	return &Embedder{
		client:     openai.NewClient(apiKey),
		model:      openai.LargeEmbedding3, // text-embedding-3-large
		maxRetries: 3,
		retryDelay: 2 * time.Second,
		dimensions: 1536,
	}
}

// NewEmbedderWithConfig creates an embedder with custom configuration
func NewEmbedderWithConfig(config EmbedderConfig) *Embedder {
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if config.Model == "" {
		config.Model = openai.LargeEmbedding3
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}

	return &Embedder{
		client:     openai.NewClient(config.APIKey),
		model:      config.Model,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
		dimensions: 1536,
	}
}

// GenerateEmbedding generates a vector embedding for attribute metadata
func (e *Embedder) GenerateEmbedding(ctx context.Context, m model.AttributeMetadata) ([]float32, error) {
	input := m.ToEmbeddingText()

	if input == "" {
		return nil, fmt.Errorf("cannot generate embedding for empty text")
	}

	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(e.retryDelay)
			fmt.Printf("Retrying embedding generation for %s (attempt %d/%d)\n",
				m.AttributeCode, attempt, e.maxRetries)
		}

		resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: e.model,
			Input: []string{input},
		})

		if err != nil {
			lastErr = err
			continue
		}

		if len(resp.Data) == 0 {
			lastErr = fmt.Errorf("no embedding data returned")
			continue
		}

		return resp.Data[0].Embedding, nil
	}

	return nil, fmt.Errorf("failed to generate embedding after %d attempts: %w",
		e.maxRetries, lastErr)
}

// GenerateEmbeddingFromText generates an embedding from raw text
func (e *Embedder) GenerateEmbeddingFromText(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot generate embedding for empty text")
	}

	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(e.retryDelay)
		}

		resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: e.model,
			Input: []string{text},
		})

		if err != nil {
			lastErr = err
			continue
		}

		if len(resp.Data) == 0 {
			lastErr = fmt.Errorf("no embedding data returned")
			continue
		}

		return resp.Data[0].Embedding, nil
	}

	return nil, fmt.Errorf("failed to generate embedding after %d attempts: %w",
		e.maxRetries, lastErr)
}

// GenerateBatchEmbeddings generates embeddings for multiple attributes
func (e *Embedder) GenerateBatchEmbeddings(ctx context.Context, metadata []model.AttributeMetadata) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(metadata))

	for i, m := range metadata {
		if i > 0 && i%10 == 0 {
			// Rate limiting: pause every 10 requests
			time.Sleep(1 * time.Second)
		}

		embedding, err := e.GenerateEmbedding(ctx, m)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for %s: %w", m.AttributeCode, err)
		}

		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimension size
func (e *Embedder) GetDimensions() int {
	return e.dimensions
}

// GetModel returns the model being used
func (e *Embedder) GetModel() openai.EmbeddingModel {
	return e.model
}
