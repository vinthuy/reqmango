package service

import (
	"sort"
	"strings"
	"unicode"
)

const (
	duplicateCheckLimit      = 5
	duplicateCandidateLimit  = 50
	duplicateMinSimilarity   = 0.35
	duplicateMinNameLenRunes = 2
)

// TextSimilarity returns a score in [0, 1] comparing two issue titles/descriptions.
// Uses exact/substring match plus token or character-bigram Jaccard.
func TextSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}

	ra, rb := []rune(a), []rune(b)
	if strings.Contains(a, b) || strings.Contains(b, a) {
		shorter, longer := len(ra), len(rb)
		if shorter > longer {
			shorter, longer = longer, shorter
		}
		if longer == 0 {
			return 0
		}
		return float64(shorter) / float64(longer) * 0.95
	}

	tokensA := tokenizeIssueText(a)
	tokensB := tokenizeIssueText(b)
	if len(tokensA) > 1 || len(tokensB) > 1 {
		return jaccard(tokensA, tokensB)
	}
	return jaccard(charNgrams(ra, 2), charNgrams(rb, 2))
}

func tokenizeIssueText(s string) map[string]struct{} {
	out := make(map[string]struct{})
	var cur strings.Builder
	flush := func() {
		t := strings.ToLower(cur.String())
		cur.Reset()
		if len([]rune(t)) < 2 {
			return
		}
		out[t] = struct{}{}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			flush()
		}
	}
	if cur.Len() > 0 {
		flush()
	}
	return out
}

func charNgrams(runes []rune, n int) map[string]struct{} {
	out := make(map[string]struct{})
	if len(runes) < n {
		if len(runes) > 0 {
			out[string(runes)] = struct{}{}
		}
		return out
	}
	for i := 0; i <= len(runes)-n; i++ {
		out[string(runes[i:i+n])] = struct{}{}
	}
	return out
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

type scoredDuplicate struct {
	ID         uint64
	SequenceID int
	Name       string
	Priority   string
	StateID    uint64
	Similarity float64
}

func rankDuplicates(name, description string, candidates []scoredDuplicate) []scoredDuplicate {
	type scored struct {
		item scoredDuplicate
		score float64
	}
	var scoredList []scored
	for _, c := range candidates {
		score := TextSimilarity(name, c.Name)
		if description != "" {
			descScore := TextSimilarity(description, c.Name) * 0.5
			if descScore > score {
				score = descScore
			}
		}
		if score < duplicateMinSimilarity {
			continue
		}
		c.Similarity = score
		scoredList = append(scoredList, scored{item: c, score: score})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score == scoredList[j].score {
			return scoredList[i].item.ID > scoredList[j].item.ID
		}
		return scoredList[i].score > scoredList[j].score
	})
	limit := duplicateCheckLimit
	if len(scoredList) < limit {
		limit = len(scoredList)
	}
	out := make([]scoredDuplicate, limit)
	for i := 0; i < limit; i++ {
		out[i] = scoredList[i].item
	}
	return out
}
