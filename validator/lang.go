package validator

import (
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	playground "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// 内置语言码。区域变体（如 zh-CN）归一到基础码。
const (
	LangZH = "zh"
	LangEN = "en"
)

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(lang), "_", "-"))
	if i := strings.IndexByte(lang, '-'); i > 0 {
		lang = lang[:i]
	}
	return lang
}

func setupTranslators(validate *playground.Validate) map[string]ut.Translator {
	zhLocale, enLocale := zh.New(), en.New()
	uni := ut.New(zhLocale, zhLocale, enLocale)
	zhTrans, _ := uni.GetTranslator(LangZH)
	enTrans, _ := uni.GetTranslator(LangEN)
	_ = zhTranslations.RegisterDefaultTranslations(validate, zhTrans)
	_ = enTranslations.RegisterDefaultTranslations(validate, enTrans)
	return map[string]ut.Translator{
		LangZH: zhTrans,
		LangEN: enTrans,
	}
}
