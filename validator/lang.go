package validator

import (
	"strings"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	playground "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// 内置语言码（基础码）。区域变体（如 zh-CN）会归一到基础码。
const (
	LangZH = "zh"
	LangEN = "en"
)

// defaultLocales 默认注册的语言列表。
var defaultLocales = []string{LangZH, LangEN}

// NormalizeLang 将语言码归一为内置基础码。
//
//	zh-CN / zh_CN → zh
//	en-US → en
//	空字符串 → 空（由调用方决定默认语言）
func NormalizeLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	if i := strings.IndexAny(lang, ",;"); i >= 0 {
		lang = strings.TrimSpace(lang[:i])
	}
	lang = strings.ReplaceAll(lang, "_", "-")
	if i := strings.Index(lang, "-"); i > 0 {
		lang = lang[:i]
	}
	return strings.ToLower(strings.TrimSpace(lang))
}

func newLocale(lang string) locales.Translator {
	switch NormalizeLang(lang) {
	case LangZH:
		return zh.New()
	case LangEN:
		return en.New()
	default:
		return nil
	}
}

func registerLocaleTranslations(validate *playground.Validate, lang string, trans ut.Translator) error {
	switch NormalizeLang(lang) {
	case LangZH:
		return zhTranslations.RegisterDefaultTranslations(validate, trans)
	case LangEN:
		return enTranslations.RegisterDefaultTranslations(validate, trans)
	default:
		return nil
	}
}

// setupTranslators 按 locales 列表初始化 universal-translator，并注册内置 tag 翻译。
// 返回 map[lang]Translator；未知语言会被跳过。
func setupTranslators(validate *playground.Validate, localeCodes []string) map[string]ut.Translator {
	if len(localeCodes) == 0 {
		localeCodes = append([]string(nil), defaultLocales...)
	}

	seen := make(map[string]struct{}, len(localeCodes))
	var codes []string
	var localeList []locales.Translator
	var fallback locales.Translator

	for _, code := range localeCodes {
		code = NormalizeLang(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		loc := newLocale(code)
		if loc == nil {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
		localeList = append(localeList, loc)
		if fallback == nil {
			fallback = loc
		}
	}

	if fallback == nil {
		// 保底：至少中文
		fallback = zh.New()
		codes = []string{LangZH}
		localeList = []locales.Translator{fallback}
	}

	uni := ut.New(fallback, localeList...)
	result := make(map[string]ut.Translator, len(codes))
	for _, code := range codes {
		trans, found := uni.GetTranslator(code)
		if !found {
			continue
		}
		_ = registerLocaleTranslations(validate, code, trans)
		result[code] = trans
	}
	return result
}
