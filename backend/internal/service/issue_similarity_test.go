package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextSimilarity_ExactMatch(t *testing.T) {
	assert.InDelta(t, 1.0, TextSimilarity("Fix login bug", "Fix login bug"), 0.001)
	assert.InDelta(t, 1.0, TextSimilarity("  Fix Login Bug  ", "fix login bug"), 0.001)
}

func TestTextSimilarity_SimilarTitles(t *testing.T) {
	score := TextSimilarity("Fix login bug on mobile", "Fix login bug")
	assert.Greater(t, score, duplicateMinSimilarity)
	assert.Greater(t, score, TextSimilarity("Fix login bug", "Rewrite payment gateway"))
}

func TestTextSimilarity_CJK(t *testing.T) {
	score := TextSimilarity("修复登录失败问题", "修复登录失败")
	assert.Greater(t, score, duplicateMinSimilarity)
}

func TestTextSimilarity_Empty(t *testing.T) {
	assert.Equal(t, 0.0, TextSimilarity("", "anything"))
	assert.Equal(t, 0.0, TextSimilarity("anything", ""))
}

func TestRankDuplicates_ReturnsTopSimilar(t *testing.T) {
	candidates := []scoredDuplicate{
		{ID: 1, SequenceID: 1, Name: "Unrelated payment rewrite", Priority: "low", StateID: 1},
		{ID: 2, SequenceID: 2, Name: "Fix login bug", Priority: "high", StateID: 1},
		{ID: 3, SequenceID: 3, Name: "Fix login bug on mobile", Priority: "medium", StateID: 2},
		{ID: 4, SequenceID: 4, Name: "Update docs", Priority: "none", StateID: 1},
	}
	got := rankDuplicates("Fix login bug", "", candidates)
	assert.NotEmpty(t, got)
	assert.LessOrEqual(t, len(got), duplicateCheckLimit)
	assert.Equal(t, uint64(2), got[0].ID)
	assert.GreaterOrEqual(t, got[0].Similarity, duplicateMinSimilarity)
}
