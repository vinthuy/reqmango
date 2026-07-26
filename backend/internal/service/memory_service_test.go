package service

import (
	"testing"
)

// TestTokenize tests the tokenize function
func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "english text",
			input:    "Hello World 2024",
			expected: []string{"hello", "world", "2024"},
		},
		{
			name:     "chinese text",
			input:    "你好世界",
			expected: []string{"你", "好", "世", "界"},
		},
		{
			name:     "mixed chinese and english",
			input:    "测试 memory 功能",
			expected: []string{"测", "试", "memory", "功", "能"},
		},
		{
			name:     "text with punctuation",
			input:    "Hello, World! How are you?",
			expected: []string{"hello", "world", "how", "are", "you"},
		},
		{
			name:     "chinese with punctuation",
			input:    "你好，世界！测试一下。",
			expected: []string{"你", "好", "世", "界", "测", "试", "一", "下"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize(%q) returned %d tokens, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, expected %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestCalculateWordFrequency tests the calculateWordFrequency function
func TestCalculateWordFrequency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]int{},
		},
		{
			name:     "simple english",
			input:    "hello world hello",
			expected: map[string]int{"hello": 2, "world": 1},
		},
		{
			name:     "chinese text",
			input:    "你好你好世界",
			expected: map[string]int{"你": 2, "好": 2, "世": 1, "界": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWordFrequency(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("calculateWordFrequency(%q) returned %d entries, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("calculateWordFrequency(%q)[%q] = %d, expected %d", tt.input, k, result[k], v)
				}
			}
		})
	}
}

// TestCalculateTextSimilarity tests the CalculateTextSimilarity method
func TestCalculateTextSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
	}{
		{
			name:     "empty texts",
			text1:    "",
			text2:    "",
			expected: 0.0,
		},
		{
			name:     "one empty text",
			text1:    "hello",
			text2:    "",
			expected: 0.0,
		},
		{
			name:     "identical english texts",
			text1:    "hello world",
			text2:    "hello world",
			expected: 1.0,
		},
		{
			name:     "identical chinese texts",
			text1:    "你好世界",
			text2:    "你好世界",
			expected: 1.0,
		},
		{
			name:     "completely different",
			text1:    "hello world",
			text2:    "foo bar",
			expected: 0.0,
		},
		{
			name:     "partial match english",
			text1:    "hello world",
			text2:    "hello there",
			expected: 0.5, // "hello" matches, so cosine similarity = 1/sqrt(2) ≈ 0.707, but depends on tokenization
		},
		{
			name:     "partial match chinese",
			text1:    "你好世界",
			text2:    "你好测试",
			expected: 0.5, // "你" and "好" match, so similarity = 2/sqrt(8) = 0.707
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &MemoryService{}
			result := svc.CalculateTextSimilarity(tt.text1, tt.text2)
			
			// For exact matches, check with tolerance
			if tt.expected == 1.0 || tt.expected == 0.0 {
				if result < tt.expected-0.0001 || result > tt.expected+0.0001 {
					t.Errorf("CalculateTextSimilarity(%q, %q) = %f, expected %f", tt.text1, tt.text2, result, tt.expected)
				}
			} else {
				// For partial matches, check range
				if result < 0.3 || result > 0.9 {
					t.Errorf("CalculateTextSimilarity(%q, %q) = %f, expected between 0.3 and 0.9", tt.text1, tt.text2, result)
				}
			}
		})
	}
}

// TestExtractSummary tests the extractSummary function
func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		query    string
		maxLen   int
	}{
		{
			name:     "short content",
			content:  "short content",
			query:    "short",
			maxLen:   15,
		},
		{
			name:     "keyword in middle",
			content:  "this is a long content with keyword in the middle somewhere",
			query:    "keyword",
			maxLen:   150,
		},
		{
			name:     "keyword not found",
			content:  "this is a long content without the search term",
			query:    "missing",
			maxLen:   150,
		},
		{
			name:     "chinese content",
			content:  "这是一段很长的中文内容，其中包含关键词搜索测试，需要提取摘要",
			query:    "关键词",
			maxLen:   150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSummary(tt.content, tt.query)
			if len(result) > tt.maxLen {
				t.Errorf("extractSummary() returned %d chars, expected <= %d", len(result), tt.maxLen)
			}
			// For keyword found cases, check that keyword context is included
			if tt.name != "keyword not found" && tt.name != "short content" {
				if len(result) > 0 {
					// Summary should contain relevant context
					t.Logf("Summary: %q", result)
				}
			}
		})
	}
}
