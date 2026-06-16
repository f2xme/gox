package mask

import "strings"

// MaskString 对字符串进行通用脱敏，保留前缀和后缀，中间替换为 mask。
func MaskString(s string, prefix, suffix int, mask string) string {
	if s == "" {
		return ""
	}
	if prefix < 0 {
		prefix = 0
	}
	if suffix < 0 {
		suffix = 0
	}

	runes := []rune(s)
	length := len(runes)
	if prefix+suffix >= length {
		return s
	}
	return string(runes[:prefix]) + mask + string(runes[length-suffix:])
}

// MaskPhone 对手机号进行脱敏，保留前三位和后四位。
func MaskPhone(phone string) string {
	return maskPhone(phone)
}

func maskPhone(phone string) string {
	return MaskString(phone, 3, 4, "****")
}

// MaskEmail 对邮箱地址进行脱敏，保留部分本地名称和完整域名。
func MaskEmail(email string) string {
	return maskEmail(email)
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	atIndex := strings.IndexRune(email, '@')
	if atIndex <= 0 {
		return email
	}

	local := email[:atIndex]
	domain := email[atIndex:]
	runes := []rune(local)
	if len(runes) <= 3 {
		return string(runes[:1]) + "****" + domain
	}
	return string(runes[:3]) + "****" + domain
}

// MaskIDCard 对身份证号进行脱敏，保留前六位和后四位。
func MaskIDCard(id string) string {
	return MaskString(id, 6, 4, "********")
}

// MaskBankCard 对银行卡号进行脱敏，保留前四位和后四位。
func MaskBankCard(card string) string {
	return MaskString(card, 4, 4, "********")
}

// MaskName 对姓名进行脱敏，保留首个字符。
func MaskName(name string) string {
	if name == "" {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	return string(runes[:1]) + strings.Repeat("*", len(runes)-1)
}
