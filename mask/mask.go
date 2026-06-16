package mask

import "strings"

// String 对字符串进行通用脱敏，保留前缀和后缀，中间替换为 mask。
func String(s string, prefix, suffix int, mask string) string {
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

// MaskString 对字符串进行通用脱敏，保留前缀和后缀，中间替换为 mask。
//
// Deprecated: 请使用 String。
func MaskString(s string, prefix, suffix int, mask string) string {
	return String(s, prefix, suffix, mask)
}

// Phone 对手机号进行脱敏，保留前三位和后四位。
func Phone(phone string) string {
	return String(phone, 3, 4, "****")
}

// MaskPhone 对手机号进行脱敏，保留前三位和后四位。
//
// Deprecated: 请使用 Phone。
func MaskPhone(phone string) string {
	return Phone(phone)
}

// Email 对邮箱地址进行脱敏，保留部分本地名称和完整域名。
func Email(email string) string {
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

// MaskEmail 对邮箱地址进行脱敏，保留部分本地名称和完整域名。
//
// Deprecated: 请使用 Email。
func MaskEmail(email string) string {
	return Email(email)
}

// IDCard 对身份证号进行脱敏，保留前六位和后四位。
func IDCard(id string) string {
	return String(id, 6, 4, "********")
}

// MaskIDCard 对身份证号进行脱敏，保留前六位和后四位。
//
// Deprecated: 请使用 IDCard。
func MaskIDCard(id string) string {
	return IDCard(id)
}

// BankCard 对银行卡号进行脱敏，保留前四位和后四位。
func BankCard(card string) string {
	return String(card, 4, 4, "********")
}

// MaskBankCard 对银行卡号进行脱敏，保留前四位和后四位。
//
// Deprecated: 请使用 BankCard。
func MaskBankCard(card string) string {
	return BankCard(card)
}

// Name 对姓名进行脱敏，保留首个字符。
func Name(name string) string {
	if name == "" {
		return ""
	}

	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	return string(runes[:1]) + strings.Repeat("*", len(runes)-1)
}

// MaskName 对姓名进行脱敏，保留首个字符。
//
// Deprecated: 请使用 Name。
func MaskName(name string) string {
	return Name(name)
}
