package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPMAgentDefs_HasFour(t *testing.T) {
	assert.Len(t, PMAgentDefs(), 4)
	names := map[string]bool{}
	for _, d := range PMAgentDefs() {
		names[d.Name] = true
		assert.NotEmpty(t, d.SystemPrompt)
	}
	assert.True(t, names["请求分诊"])
	assert.True(t, names["交付风险"])
	assert.True(t, names["Sprint 总结"])
	assert.True(t, names["Spec 草稿"])
}
