package validator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 封装 go-playground/validator 实例，提供数据验证功能。
// 实例是并发安全的，可以在多个 goroutine 中共享使用。
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
	mu       sync.RWMutex
}

var (
	// 默认验证器实例，使用 sync.Once 确保只初始化一次
	defaultValidator     *Validator
	defaultValidatorOnce sync.Once
)

// New 创建一个新的验证器实例，默认支持中文错误消息和中国本地化验证规则。
//
// 返回的验证器实例是并发安全的，可以在多个 goroutine 中共享使用。
//
// 示例：
//
//	v := validator.New()
//	err := v.Validate(user)
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

	// 设置中文翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	trans, _ := uni.GetTranslator("zh")

	// 注册中文翻译
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)

	v := &Validator{
		validate: validate,
		trans:    trans,
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

// Validate 验证结构体字段是否符合标签定义的规则。
//
// 如果验证失败，返回包含所有错误信息的 error（中文描述）。
// 如果验证成功，返回 nil。
//
// 参数 i 应该是一个结构体或结构体指针。
//
// 示例：
//
//	type User struct {
//	    Name string `validate:"required"`
//	}
//	user := User{Name: ""}
//	err := v.Validate(user) // 返回错误：Name为必填字段
func (v *Validator) Validate(i any) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	// 转换验证错误为友好的中文消息
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	return v.formatErrors(validationErrs)
}

// formatErrors 将验证错误格式化为友好的中文消息
func (v *Validator) formatErrors(errs validator.ValidationErrors) error {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Translate(v.trans))
	}
	return fmt.Errorf("%s", strings.Join(messages, "; "))
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

// RegisterTranslation 注册自定义验证规则的翻译消息。
//
// tag 是验证标签名称，message 是错误消息模板。
//
// 示例：
//
//	v.RegisterTranslation("custom_username", "用户名格式不正确")
func (v *Validator) RegisterTranslation(tag, message string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.validate.RegisterTranslation(
		tag,
		v.trans,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field())
			return t
		},
	)
}

// Validate 使用默认验证器验证结构体。
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

// RegisterTranslation 在默认验证器上注册自定义翻译消息。
//
// 这是一个便捷函数，内部使用全局默认验证器实例。
//
// 示例：
//
//	validator.RegisterTranslation("phone", "手机号格式不正确")
func RegisterTranslation(tag, message string) error {
	return getDefaultValidator().RegisterTranslation(tag, message)
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
	_ = v.RegisterTranslation("phone", "{0}手机号格式不正确")

	_ = v.validate.RegisterValidation("id_card", validateIDCard)
	_ = v.RegisterTranslation("id_card", "{0}身份证号格式不正确")

	_ = v.validate.RegisterValidation("bank_card", validateBankCard)
	_ = v.RegisterTranslation("bank_card", "{0}银行卡号格式不正确")
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
