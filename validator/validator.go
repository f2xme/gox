package validator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// Validator 封装 go-playground/validator 实例，提供数据验证功能。
// 实例是并发安全的，可以在多个 goroutine 中共享使用。
//
// 默认注册中文（zh）与英文（en）错误消息；Validate() 使用默认语言（zh），
// ValidateWithLang() 可按请求语言输出对应文案。
type Validator struct {
	validate    *validator.Validate
	trans       map[string]ut.Translator // lang → translator
	defaultLang string
	mu          sync.RWMutex
}

var (
	// 默认验证器实例，使用 sync.Once 确保只初始化一次
	defaultValidator     *Validator
	defaultValidatorOnce sync.Once
)

// New 创建一个新的验证器实例。
//
// 默认：
//   - 字段名标签：label
//   - 默认语言：zh
//   - 已注册语言：zh、en
//   - 内置中国本地化规则：phone / id_card / bank_card
//
// 返回的验证器实例是并发安全的，可以在多个 goroutine 中共享使用。
//
// 示例：
//
//	v := validator.New()
//	err := v.Validate(user)                      // 中文
//	err = v.ValidateWithLang(user, validator.LangEN) // 英文
func New(opts ...Option) *Validator {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if options.FieldNameTag != "" {
			name := parseTagName(fld.Tag.Get(options.FieldNameTag))
			if name != "" {
				return name
			}
		}
		return fld.Name
	})

	transMap := setupTranslators(validate, options.Locales)
	defaultLang := NormalizeLang(options.DefaultLang)
	if defaultLang == "" {
		defaultLang = LangZH
	}
	if _, ok := transMap[defaultLang]; !ok {
		// 回退：优先 zh，否则取任意已注册语言
		if _, ok := transMap[LangZH]; ok {
			defaultLang = LangZH
		} else {
			for lang := range transMap {
				defaultLang = lang
				break
			}
		}
	}

	v := &Validator{
		validate:    validate,
		trans:       transMap,
		defaultLang: defaultLang,
	}
	v.registerBuiltinValidations()

	return v
}

// getDefaultValidator 获取默认验证器实例（懒加载，并发安全）
func getDefaultValidator() *Validator {
	defaultValidatorOnce.Do(func() {
		defaultValidator = New()
	})
	return defaultValidator
}

// Default 返回全局默认验证器实例。
//
// 默认验证器使用懒加载初始化，并且可以安全地在多个 goroutine 中共享。
//
// 示例：
//
//	v := validator.Default()
//	err := v.Validate(user)
func Default() *Validator {
	return getDefaultValidator()
}

// DefaultLang 返回 Validate() 使用的默认语言。
func (v *Validator) DefaultLang() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.defaultLang
}

// Locales 返回已注册的语言列表（无序）。
func (v *Validator) Locales() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]string, 0, len(v.trans))
	for lang := range v.trans {
		out = append(out, lang)
	}
	return out
}

// Validate 验证结构体字段是否符合标签定义的规则。
//
// 使用默认语言（见 WithDefaultLang，默认 zh）生成错误消息。
// 需要按请求语言输出时，请使用 ValidateWithLang。
//
// 如果验证失败，返回 *ValidationError；成功返回 nil。
//
// 参数 i 应该是一个结构体或结构体指针。
//
// 示例：
//
//	type User struct {
//	    Name string `validate:"required" label:"姓名"`
//	}
//	err := v.Validate(User{}) // 姓名为必填字段
func (v *Validator) Validate(i any) error {
	return v.ValidateWithLang(i, v.DefaultLang())
}

// ValidateWithLang 使用指定语言验证结构体。
//
// lang 会归一化（zh-CN → zh）。未注册语言回退到默认语言。
//
// 示例：
//
//	err := v.ValidateWithLang(user, "en")
//	err := v.ValidateWithLang(user, "zh-CN")
func (v *Validator) ValidateWithLang(i any, lang string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	return v.formatErrors(validationErrs, lang)
}

// formatErrors 将验证错误格式化为可识别的验证错误。
func (v *Validator) formatErrors(errs validator.ValidationErrors, lang string) error {
	trans := v.translatorLocked(lang)
	fields := make([]FieldError, 0, len(errs))
	for _, fe := range errs {
		fields = append(fields, FieldError{
			Namespace:   fe.Namespace(),
			Field:       fe.Field(),
			StructField: fe.StructField(),
			Tag:         fe.Tag(),
			Param:       fe.Param(),
			Message:     fe.Translate(trans),
		})
	}
	return &ValidationError{fields: fields}
}

// translatorLocked 在已持有读锁时解析 translator。
func (v *Validator) translatorLocked(lang string) ut.Translator {
	lang = NormalizeLang(lang)
	if lang != "" {
		if t, ok := v.trans[lang]; ok {
			return t
		}
	}
	if t, ok := v.trans[v.defaultLang]; ok {
		return t
	}
	for _, t := range v.trans {
		return t
	}
	// 不应发生：New 至少注册一种语言
	return nil
}

// RegisterValidation 注册自定义验证规则。
//
// tag 是验证标签名称，fn 是验证函数。
// 验证函数返回 true 表示验证通过，false 表示验证失败。
//
// 示例：
//
//	v.RegisterValidation("custom_username", func(fl validator.FieldLevel) bool {
//	    username := fl.Field().String()
//	    return len(username) >= 3 && unicode.IsLetter(rune(username[0]))
//	})
//
//	type User struct {
//	    Username string `validate:"custom_username"`
//	}
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.validate.RegisterValidation(tag, fn)
}

// RegisterTranslation 为默认语言注册自定义验证规则的翻译消息。
//
// tag 是验证标签名称，message 是错误消息模板（{0} 为字段名）。
//
// 需要多语言时请用 RegisterTranslationLang 或 RegisterTranslations。
//
// 示例：
//
//	v.RegisterTranslation("custom_username", "用户名格式不正确")
func (v *Validator) RegisterTranslation(tag, message string) error {
	return v.RegisterTranslationLang(tag, v.DefaultLang(), message)
}

// RegisterTranslationLang 为指定语言注册自定义验证规则的翻译消息。
//
// 示例：
//
//	v.RegisterTranslationLang("custom_username", "zh", "{0}用户名格式不正确")
//	v.RegisterTranslationLang("custom_username", "en", "{0} is not a valid username")
func (v *Validator) RegisterTranslationLang(tag, lang, message string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	lang = NormalizeLang(lang)
	if lang == "" {
		lang = v.defaultLang
	}
	trans, ok := v.trans[lang]
	if !ok {
		return errUnknownLocale(lang)
	}
	return v.registerTranslationLocked(tag, trans, message)
}

// RegisterTranslations 为多种语言批量注册同一 tag 的翻译。
//
// 示例：
//
//	v.RegisterTranslations("custom_username", map[string]string{
//	    "zh": "{0}用户名格式不正确",
//	    "en": "{0} is not a valid username",
//	})
func (v *Validator) RegisterTranslations(tag string, messages map[string]string) error {
	for lang, message := range messages {
		if err := v.RegisterTranslationLang(tag, lang, message); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) registerTranslationLocked(tag string, trans ut.Translator, message string) error {
	return v.validate.RegisterTranslation(
		tag,
		trans,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field())
			return t
		},
	)
}

// Validate 使用默认验证器验证结构体（默认语言）。
//
// 这是一个便捷函数，内部使用全局默认验证器实例。
// 默认验证器是并发安全的。
//
// 示例：
//
//	type User struct {
//	    Name string `validate:"required"`
//	}
//	user := User{Name: "张三"}
//	if err := validator.Validate(user); err != nil {
//	    log.Fatal(err)
//	}
func Validate(i any) error {
	return getDefaultValidator().Validate(i)
}

// ValidateWithLang 使用默认验证器、指定语言验证结构体。
//
// 示例：
//
//	err := validator.ValidateWithLang(user, "en")
func ValidateWithLang(i any, lang string) error {
	return getDefaultValidator().ValidateWithLang(i, lang)
}

// RegisterValidation 在默认验证器上注册自定义验证规则。
//
// 这是一个便捷函数，内部使用全局默认验证器实例。
//
// 示例：
//
//	validator.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
//	    phone := fl.Field().String()
//	    return len(phone) == 11
//	})
func RegisterValidation(tag string, fn validator.Func) error {
	return getDefaultValidator().RegisterValidation(tag, fn)
}

// RegisterTranslation 在默认验证器上为默认语言注册自定义翻译消息。
//
// 示例：
//
//	validator.RegisterTranslation("phone", "手机号格式不正确")
func RegisterTranslation(tag, message string) error {
	return getDefaultValidator().RegisterTranslation(tag, message)
}

// RegisterTranslationLang 在默认验证器上为指定语言注册自定义翻译消息。
func RegisterTranslationLang(tag, lang, message string) error {
	return getDefaultValidator().RegisterTranslationLang(tag, lang, message)
}

// RegisterTranslations 在默认验证器上批量注册多语言翻译。
func RegisterTranslations(tag string, messages map[string]string) error {
	return getDefaultValidator().RegisterTranslations(tag, messages)
}

func parseTagName(tag string) string {
	if tag == "" {
		return ""
	}
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		tag = tag[:idx]
	}
	if tag == "-" {
		return ""
	}
	return tag
}

func (v *Validator) registerBuiltinValidations() {
	_ = v.validate.RegisterValidation("phone", validatePhone)
	_ = v.registerBuiltinTranslations("phone", map[string]string{
		LangZH: "{0}手机号格式不正确",
		LangEN: "{0} is not a valid mobile phone number",
	})

	_ = v.validate.RegisterValidation("id_card", validateIDCard)
	_ = v.registerBuiltinTranslations("id_card", map[string]string{
		LangZH: "{0}身份证号格式不正确",
		LangEN: "{0} is not a valid ID card number",
	})

	_ = v.validate.RegisterValidation("bank_card", validateBankCard)
	_ = v.registerBuiltinTranslations("bank_card", map[string]string{
		LangZH: "{0}银行卡号格式不正确",
		LangEN: "{0} is not a valid bank card number",
	})
}

// registerBuiltinTranslations 仅向已注册语言写入内置文案。
func (v *Validator) registerBuiltinTranslations(tag string, messages map[string]string) error {
	filtered := make(map[string]string, len(messages))
	for lang, message := range messages {
		lang = NormalizeLang(lang)
		if _, ok := v.trans[lang]; ok {
			filtered[lang] = message
		}
	}
	return v.RegisterTranslations(tag, filtered)
}

// validatePhone 验证中国大陆手机号（11位，1开头）
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// 长度检查（防止 DoS）
	if len(phone) > 50 {
		return false
	}

	// 必须是 11 位
	if len(phone) != 11 {
		return false
	}

	// 第一位必须是 1
	if phone[0] != '1' {
		return false
	}

	return isDigitsOnly(phone)
}

// validateIDCard 验证中国大陆身份证号（18位，含校验位）
func validateIDCard(fl validator.FieldLevel) bool {
	idCard := fl.Field().String()

	// 长度检查（防止 DoS）
	if len(idCard) > 50 {
		return false
	}

	// 必须是 18 位
	if len(idCard) != 18 {
		return false
	}

	// 前 17 位必须是数字
	for i := 0; i < 17; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
	}

	// 第 18 位必须是数字或 X
	last := idCard[17]
	if !((last >= '0' && last <= '9') || last == 'X') {
		return false
	}

	// 校验位算法（GB 11643-1999）
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		digit := int(idCard[i] - '0')
		sum += digit * weights[i]
	}

	expectedCheck := checkCodes[sum%11]
	return idCard[17] == expectedCheck
}

// validateBankCard 验证银行卡号（Luhn 算法）
func validateBankCard(fl validator.FieldLevel) bool {
	cardNumber := fl.Field().String()

	// 长度检查（防止 DoS）
	if len(cardNumber) > 50 {
		return false
	}

	// 长度必须在 13-19 位之间
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	if !isDigitsOnly(cardNumber) {
		return false
	}

	// Luhn 算法校验
	sum := 0
	isEven := false

	// 从右往左遍历
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

func isDigitsOnly(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
