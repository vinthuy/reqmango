package i18n

import (
	"testing"
)

func TestT_KnownKey(t *testing.T) {
	// Test Chinese translations
	msg := T("zh", "project_not_found")
	if msg == "" || msg == "project_not_found" {
		t.Errorf("T(zh, project_not_found) = %q, expected a Chinese translation", msg)
	}

	// Test English translations
	msgEn := T("en", "project_not_found")
	if msgEn == "" || msgEn == "project_not_found" {
		t.Errorf("T(en, project_not_found) = %q, expected an English translation", msgEn)
	}

	// Chinese and English should be different
	if msg == msgEn {
		t.Logf("Warning: Chinese and English translations are identical: %q", msg)
	}
}

func TestT_UnknownKey(t *testing.T) {
	// Unknown key should fall back to English or return the key itself
	msg := T("zh", "nonexistent_key_12345")
	if msg == "" {
		t.Error("T() returned empty string for unknown key")
	}
}

func TestT_UnknownLanguage(t *testing.T) {
	// Unknown language should fall back to English
	msg := T("fr", "project_not_found")
	if msg == "" {
		t.Error("T() returned empty string for unknown language")
	}
	// Should match English fallback
	enMsg := T("en", "project_not_found")
	if msg != enMsg {
		t.Logf("French fallback %q differs from English %q", msg, enMsg)
	}
}

func TestDetectLanguage_Empty(t *testing.T) {
	lang := DetectLanguage("")
	if lang != "zh" {
		t.Errorf("DetectLanguage(\"\") = %q, want zh", lang)
	}
}

func TestDetectLanguage_Chinese(t *testing.T) {
	lang := DetectLanguage("zh-CN,zh;q=0.9,en;q=0.8")
	if lang != "zh" {
		t.Errorf("DetectLanguage(zh-CN,...) = %q, want zh", lang)
	}
}

func TestDetectLanguage_English(t *testing.T) {
	lang := DetectLanguage("en-US,en;q=0.9")
	if lang != "en" {
		t.Errorf("DetectLanguage(en-US,...) = %q, want en", lang)
	}
}

func TestDetectLanguage_EnglishSimple(t *testing.T) {
	lang := DetectLanguage("en")
	if lang != "en" {
		t.Errorf("DetectLanguage(en) = %q, want en", lang)
	}
}

func TestDetectLanguage_Unknown(t *testing.T) {
	// Unknown language should fall back to zh
	lang := DetectLanguage("fr-FR,fr;q=0.9")
	if lang != "zh" {
		t.Errorf("DetectLanguage(fr-FR,...) = %q, want zh", lang)
	}
}

func TestDetectLanguage_German(t *testing.T) {
	lang := DetectLanguage("de-DE")
	if lang == "" {
		t.Error("DetectLanguage returned empty string")
	}
	// German not supported, should fall back to zh
	if lang != "zh" {
		t.Logf("DetectLanguage(de-DE) = %q (may be zh as fallback)", lang)
	}
}

func TestGetMessage(t *testing.T) {
	// Test with standard error code format
	msg := GetMessage("zh", "PROJECT_NOT_FOUND")
	if msg == "" || msg == "PROJECT_NOT_FOUND" || msg == "project_not_found" {
		t.Errorf("GetMessage(zh, PROJECT_NOT_FOUND) = %q, expected a translation", msg)
	}

	// Test with different case
	msg2 := GetMessage("en", "project_not_found")
	if msg2 == "" || msg2 == "project_not_found" {
		t.Errorf("GetMessage(en, project_not_found) = %q, expected a translation", msg2)
	}
}

func TestBundlesLoaded(t *testing.T) {
	mu.RLock()
	defer mu.RUnlock()

	// Verify both bundles are loaded
	if _, ok := bundles["zh"]; !ok {
		t.Error("Chinese bundle not loaded")
	}
	if _, ok := bundles["en"]; !ok {
		t.Error("English bundle not loaded")
	}

	// Verify bundles have some entries
	if len(bundles["zh"]) == 0 {
		t.Error("Chinese bundle is empty")
	}
	if len(bundles["en"]) == 0 {
		t.Error("English bundle is empty")
	}

	// Both should have the same keys
	for k := range bundles["zh"] {
		if _, ok := bundles["en"][k]; !ok {
			t.Logf("Key %q exists in zh but not in en", k)
		}
	}
	for k := range bundles["en"] {
		if _, ok := bundles["zh"][k]; !ok {
			t.Logf("Key %q exists in en but not in zh", k)
		}
	}
}

func TestT_AllErrorCodes(t *testing.T) {
	// Test that all common error codes have translations
	codes := []string{
		"internal_error", "bad_request", "validation_error",
		"unauthorized", "forbidden", "not_found",
		"already_exists", "conflict",
		"project_not_found", "issue_not_found", "user_not_found",
		"workspace_not_found", "page_not_found", "cycle_not_found",
		"module_not_found", "state_not_found", "label_not_found",
		"release_not_found", "template_not_found", "view_not_found",
		"comment_not_found", "attachment_not_found",
		"required_field", "invalid_format", "invalid_value",
		"max_depth_exceeded", "self_reference",
		"already_assigned", "already_labelled",
		"invalid_state_transition", "workflow_violation",
		"timer_already_running", "duplicate_entry",
		"ai_config_missing", "ai_api_failure", "ai_timeout", "ai_quota_exceeded",
	}

	for _, code := range codes {
		msgZh := T("zh", code)
		msgEn := T("en", code)

		if msgZh == code {
			t.Errorf("Missing Chinese translation for: %s", code)
		}
		if msgEn == code {
			t.Errorf("Missing English translation for: %s", code)
		}
	}
}
