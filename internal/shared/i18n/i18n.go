package i18n

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"
	"learning-go/internal/shared/logger"
)

const DefaultLang = "en"

var SupportedLangs = []string{"en", "vi", "th", "zh", "id"}

var supportedLangSet = func() map[string]bool {
	m := make(map[string]bool, len(SupportedLangs))
	for _, l := range SupportedLangs {
		m[l] = true
	}
	return m
}()

func Normalize(lang string) string {
	s := strings.TrimSpace(strings.ToLower(lang))
	if s == "" {
		return DefaultLang
	}
	// Extract 2-char language prefix
	prefix := s
	if len(prefix) > 2 {
		prefix = s[:2]
	}
	if supportedLangSet[prefix] {
		return prefix
	}
	return DefaultLang
}

func FromAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLang
	}
	first := strings.TrimSpace(strings.Split(header, ",")[0])
	first = strings.TrimSpace(strings.Split(first, ";")[0])
	return Normalize(first)
}

func Translate(lang string, key string) string {
	ensureLoaded()
	lang = Normalize(lang)
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if byLang, ok := catalog[key]; ok {
		if msg, ok := byLang[lang]; ok && msg != "" {
			return msg
		}
		if msg, ok := byLang[DefaultLang]; ok && msg != "" {
			return msg
		}
	}
	return key
}

func TranslateText(lang string, text string) string {
	s := strings.TrimSpace(text)
	if s == "" {
		return ""
	}
	if IsKey(s) {
		return Translate(lang, s)
	}
	return s
}

// keyPattern matches i18n keys like "common.not_found", "auth.email_already_exists"
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_.]*$`)

func IsKey(s string) bool {
	return keyPattern.MatchString(s)
}

var (
	catalog  = map[string]map[string]string{}
	loadOnce sync.Once
)

func ensureLoaded() {
	loadOnce.Do(func() {
		dir := os.Getenv("I18N_DIR")
		if strings.TrimSpace(dir) == "" {
			dir = filepath.Join("resources", "i18n")
		}

		for _, lang := range SupportedLangs {
			loadLangFiles(dir, lang)
		}
	})
}

func loadLangFiles(baseDir string, lang string) {
	langDir := filepath.Join(baseDir, lang)
	entries, err := os.ReadDir(langDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Warn("[I18N] failed to read language directory",
				zap.String("dir", langDir), zap.Error(err))
		}
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		filePath := filepath.Join(langDir, e.Name())
		loadFile(lang, filePath)
	}

	legacyFile := filepath.Join(baseDir, lang+".json")
	loadFile(lang, legacyFile)
}

func loadFile(lang string, filePath string) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Warn("[I18N] failed to read translation file",
				zap.String("file", filePath), zap.Error(err))
		}
		return
	}

	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		logger.Warn("[I18N] failed to parse translation file",
			zap.String("file", filePath), zap.Error(err))
		return
	}

	for k, v := range m {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := catalog[k]; !ok {
			catalog[k] = map[string]string{}
		}
		catalog[k][lang] = v
	}
}
