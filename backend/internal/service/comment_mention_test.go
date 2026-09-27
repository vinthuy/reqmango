package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMentionsAgent(t *testing.T) {
	cases := []struct {
		body, name string
		want       bool
	}{
		{"@Assistant Agent 请分诊", "Assistant Agent", true},
		{"请 @请求分诊 看一下", "请求分诊", true},
		{"@Triage", "Triage", true},
		{"email@Triage.com", "Triage", false},
		{"@TriageBot 看下", "Triage", false},
		{"@Assistant 请看", "Assistant Agent", false},
		{"回复:\n@Triage，处理下", "Triage", true},
		{"没有提及", "Triage", false},
		{"@", "", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, mentionsAgent(c.body, c.name), "%q / %q", c.body, c.name)
	}
}
