package service

import (
	"reflect"
	"testing"

	"github.com/reqmango/backend/internal/model"
)

func TestParseIssueKeyTitleBranchBody(t *testing.T) {
	s := &GitService{}
	got := s.ParseIssueKey("MOBILE-507 login fix\nfeature/mobile-12-oauth\nAlso relates to Mobile-507 and WEB-3")
	want := []string{"MOBILE-507", "MOBILE-12", "WEB-3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseSmartCommitCaseInsensitive(t *testing.T) {
	s := &GitService{}
	got := s.ParseSmartCommit("Fixes MOBILE-1, closed mobile-2; ref MOBILE-3\nResolves WEB-9")
	want := []map[string]string{
		{"action": "fixes", "key": "MOBILE-1"},
		{"action": "closes", "key": "MOBILE-2"},
		{"action": "ref", "key": "MOBILE-3"},
		{"action": "resolves", "key": "WEB-9"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRepoFullName(t *testing.T) {
	cases := []struct {
		url, name, ctx, want string
	}{
		{"", "", "acme/app", "acme/app"},
		{"https://github.com/acme/app.git", "", "", "acme/app"},
		{"git@github.com:acme/app", "", "", "acme/app"},
		{"", "acme/app", "", "acme/app"},
		{"", "app", "", ""},
	}
	for _, c := range cases {
		got := repoFullName(&model.GitIntegration{RepoURL: c.url, RepoName: c.name}, c.ctx)
		if got != c.want {
			t.Errorf("repoFullName(%q,%q,%q) = %q, want %q", c.url, c.name, c.ctx, got, c.want)
		}
	}
}

func TestGithubPRURL(t *testing.T) {
	m := githubPRURLRe.FindStringSubmatch("https://github.com/acme/app/pull/42/files")
	if m == nil || m[1] != "acme" || m[2] != "app" || m[3] != "42" {
		t.Fatalf("unexpected match %v", m)
	}
	if githubPRURLRe.MatchString("https://github.com/acme/app/issues/42") {
		t.Fatal("issue URL must not match")
	}
}
