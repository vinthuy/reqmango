package llm

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestErrNotConfiguredMatchesAllVariants(t *testing.T) {
	c := NewLLMClient("", "", "", "openai")

	_, err := c.ChatSync(context.Background(), "", nil, nil)
	if !errors.Is(fmt.Errorf("wrapped: %w", err), ErrNotConfigured) {
		t.Fatalf("ChatSync error %v should match ErrNotConfigured", err)
	}

	_, err = c.GenerateEmbedding(context.Background(), "x")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("embedding error %v should match ErrNotConfigured", err)
	}

	if errors.Is(errors.New("AI 服务未配置"), ErrNotConfigured) {
		t.Fatal("unrelated error must not match")
	}
}
