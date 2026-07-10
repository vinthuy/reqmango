package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

type GitWebhookHandler struct {
	gitSvc *service.GitService
}

func NewGitWebhookHandler(gitSvc *service.GitService) *GitWebhookHandler {
	return &GitWebhookHandler{gitSvc: gitSvc}
}

func (h *GitWebhookHandler) GitHubWebhook(c *gin.Context) {
	projectID, err := parseProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	integration, err := h.gitSvc.GetIntegration(projectID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Git integration not found"})
		return
	}

	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing signature"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to read body"})
		return
	}

	if !verifySignature(body, signature, integration.WebhookSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid signature"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")
	switch event {
	case "push":
		h.handleGitHubPush(c, projectID, body)
	case "pull_request":
		h.handleGitHubPR(c, projectID, body)
	case "issues":
		h.handleGitHubIssue(c, projectID, body)
	case "issue_comment":
		h.handleGitHubIssueComment(c, projectID, body)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event not handled"})
	}
}

func (h *GitWebhookHandler) handleGitHubPush(c *gin.Context, projectID uint64, body []byte) {
	var payload struct {
		Commits []map[string]interface{} `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload"})
		return
	}

	if err := h.gitSvc.HandlePushEvent(projectID, payload.Commits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to handle push event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Push event handled"})
}

func (h *GitWebhookHandler) handleGitHubPR(c *gin.Context, projectID uint64, body []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payload"})
		return
	}

	pr, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing pull_request in payload"})
		return
	}

	if err := h.gitSvc.HandlePullRequestEvent(projectID, pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to handle PR event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PR event handled"})
}

func (h *GitWebhookHandler) handleGitHubIssue(c *gin.Context, projectID uint64, body []byte) {
	c.JSON(http.StatusOK, gin.H{"message": "Issue event handled"})
}

func (h *GitWebhookHandler) handleGitHubIssueComment(c *gin.Context, projectID uint64, body []byte) {
	c.JSON(http.StatusOK, gin.H{"message": "Issue comment event handled"})
}

func verifySignature(body []byte, signature, secret string) bool {
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}

	expected := signature[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actual := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(actual))
}

func parseProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}