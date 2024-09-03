package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hako/durafmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// prediction multiplexer
func generate(model string, prompt string, seed int) (string, error) {

	llm, err := getLLM(model)
	if err != nil {
		return "", err
	}

	generated, err := complete(llm, prompt, seed)
	if err != nil {
		return "", err
	}
	return generated, nil
}

func getLLM(model string) (llms.Model, error) {
	switch {
	case strings.HasPrefix(model, "gpt-"):
		llm, err := openai.New(openai.WithModel(model))
		if err != nil {
			return nil, err
		}
		return llm, nil

	case strings.HasPrefix(model, "gemini-"):
		ctx := context.Background()
		llm, err := googleai.New(ctx, googleai.WithDefaultModel(model))
		if err != nil {
			return nil, err
		}
		return llm, nil

	case strings.HasPrefix(model, "claude-"):
		llm, err := anthropic.New(anthropic.WithModel(model))
		if err != nil {
			return nil, err
		}
		return llm, nil
	default:
		llm, err := ollama.New(ollama.WithModel(model))
		if err != nil {
			return nil, err
		}
		return llm, nil
	}
}

func complete(llm llms.Model, prompt string, seed int) (string, error) {
	t0 := time.Now()
	c := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(
		c,
		llm,
		prompt,
		llms.WithMaxTokens(1024*4),
		llms.WithMinLength(1024*2),
		llms.WithSeed(seed),
	)
	if err != nil {
		return "", err
	}
	elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(1)
	fmt.Println(elapsed)
	return completion, nil
}
