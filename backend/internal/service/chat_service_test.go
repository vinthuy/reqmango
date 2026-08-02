package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/reqmango/backend/internal/dto/request"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NOTE: This test uses sqlmock to avoid a live DB dependency. If go-sqlmock is
// not yet a dependency, run: cd backend; go get github.com/DATA-DOG/go-sqlmock
// If adding the dep is undesirable, delete this test file and rely on the
// handler integration test (Task 11) + manual verification (Task 26).

func TestChatService_SendMessage_RejectsNonMember(t *testing.T) {
	// Use the in-memory sqlite helper (setupTestDB, defined in page_service_test.go)
	// so checkProjectMembership actually runs against an empty ProjectMember table
	// and returns Forbidden. Avoids nil-db panic.
	db := setupTestDB(t)
	s := &ChatService{db: db}
	err := s.checkProjectMembership(1, 2)
	if err == nil {
		t.Fatal("expected Forbidden error for non-member, got nil")
	}
	if !strings.Contains(err.Error(), "project member") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatService_parseAndResolveMentions_EmptyContent(t *testing.T) {
	s := &ChatService{}
	// Will return nil because parseMentions("") is empty — no DB hit.
	got := s.parseAndResolveMentions("", 1)
	if len(got) != 0 {
		t.Fatalf("expected 0 mentions, got %d", len(got))
	}
}

func TestChatService_toMessageResponses_Empty(t *testing.T) {
	s := &ChatService{}
	got := s.toMessageResponses(nil)
	if got == nil || len(got) != 0 {
		t.Fatalf("expected non-nil empty slice, got %v", got)
	}
}

func TestParseMentions_ExtractsAgentName(t *testing.T) {
	names := parseMentions("hello @leader-agent please review")
	if len(names) != 1 || names[0] != "leader-agent" {
		t.Fatalf("expected [leader-agent], got %v", names)
	}
}

// Ensure request DTOs round-trip JSON as expected.
func TestSendMessageRequest_JSONBinding(t *testing.T) {
	raw := `{"content":"hi @leader-agent","reply_to_id":null}`
	var r request.SendMessageRequest
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.Content != "hi @leader-agent" {
		t.Fatalf("unexpected content: %q", r.Content)
	}
}

// Suppress unused imports when sqlmock path is conditionally compiled.
var _ = sqlmock.New
var _ = postgres.Open
var _ = gorm.ErrRecordNotFound
