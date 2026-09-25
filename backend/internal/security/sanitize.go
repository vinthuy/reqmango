package security

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// SanitizeHTML sanitizes HTML content using bluemonday's UGC policy,
// which allows common user-generated content tags while stripping
// dangerous elements like scripts and event handlers.
func SanitizeHTML(html string) string {
	if html == "" {
		return "<p></p>"
	}
	policy := bluemonday.UGCPolicy()
	return policy.Sanitize(html)
}

// StripHTMLTags removes all HTML tags from the input, returning
// the plain text content. Useful for building search-index fields
// from rich-text descriptions.
func StripHTMLTags(html string) string {
	if html == "" {
		return ""
	}
	return htmlTagRegex.ReplaceAllString(html, "")
}
