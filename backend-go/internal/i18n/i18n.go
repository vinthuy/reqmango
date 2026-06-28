package i18n

import (
	"encoding/json"
	"strings"
	"sync"

	"embed"
)

//go:embed messages_*.json
var messagesFS embed.FS

var (
	bundles = make(map[string]map[string]string)
	mu      sync.RWMutex
)

func init() {
	loadBundle("zh", "messages_zh.json")
	loadBundle("en", "messages_en.json")
}

func loadBundle(lang, filename string) {
	data, err := messagesFS.ReadFile(filename)
	if err != nil {
		return
	}
	var msgs map[string]string
	if json.Unmarshal(data, &msgs) == nil {
		mu.Lock()
		bundles[lang] = msgs
		mu.Unlock()
	}
}

// T translates a key to the given language.
func T(lang, key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if b, ok := bundles[lang]; ok {
		if msg, ok2 := b[key]; ok2 {
			return msg
		}
	}
	// Fallback to English
	if b, ok := bundles["en"]; ok {
		if msg, ok2 := b[key]; ok2 {
			return msg
		}
	}
	return key
}

// DetectLanguage extracts the preferred language from Accept-Language header.
func DetectLanguage(header string) string {
	if header == "" {
		return "zh"
	}
	// Parse "zh-CN,zh;q=0.9,en;q=0.8"
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return "zh"
	}
	lang := strings.TrimSpace(parts[0])
	// Extract primary tag: zh-CN → zh
	if idx := strings.Index(lang, "-"); idx > 0 {
		lang = lang[:idx]
	}
	lang = strings.ToLower(lang)
	if _, ok := bundles[lang]; ok {
		return lang
	}
	return "zh"
}

// GetMessage returns a translated error message by error code.
func GetMessage(lang string, errorCode string) string {
	key := strings.ToLower(errorCode)
	// Convert ERROR_CODE → error_code format
	key = strings.ReplaceAll(key, "_", "_")
	return T(lang, strings.ToLower(key))
}
