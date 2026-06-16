package mask

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		prefix int
		suffix int
		mask   string
		want   string
	}{
		{name: "empty", value: "", prefix: 1, suffix: 1, mask: "****", want: ""},
		{name: "short", value: "1234", prefix: 2, suffix: 2, mask: "****", want: "1234"},
		{name: "standard", value: "1234567890", prefix: 3, suffix: 2, mask: "****", want: "123****90"},
		{name: "negative prefix", value: "123456", prefix: -1, suffix: 2, mask: "****", want: "****56"},
		{name: "negative suffix", value: "123456", prefix: 2, suffix: -1, mask: "****", want: "12****"},
		{name: "unicode", value: "一二三四五六", prefix: 2, suffix: 2, mask: "****", want: "一二****五六"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.value, tt.prefix, tt.suffix, tt.mask); got != tt.want {
				t.Fatalf("String(%q, %d, %d, %q) = %q, want %q", tt.value, tt.prefix, tt.suffix, tt.mask, got, tt.want)
			}
		})
	}
}

func TestPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{name: "empty", phone: "", want: ""},
		{name: "short", phone: "1234567", want: "1234567"},
		{name: "standard", phone: "13812345678", want: "138****5678"},
		{name: "unicode", phone: "一二三四五六七八", want: "一二三****五六七八"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Phone(tt.phone); got != tt.want {
				t.Fatalf("Phone(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestIDCard(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "empty", id: "", want: ""},
		{name: "short", id: "1234567890", want: "1234567890"},
		{name: "standard", id: "110101199001011234", want: "110101********1234"},
		{name: "unicode", id: "一二三四五六七八九十十一十二", want: "一二三四五六********十一十二"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IDCard(tt.id); got != tt.want {
				t.Fatalf("IDCard(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestBankCard(t *testing.T) {
	tests := []struct {
		name string
		card string
		want string
	}{
		{name: "empty", card: "", want: ""},
		{name: "short", card: "12345678", want: "12345678"},
		{name: "standard", card: "6222021234567890123", want: "6222********0123"},
		{name: "unicode", card: "一二三四五六七八九", want: "一二三四********六七八九"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BankCard(tt.card); got != tt.want {
				t.Fatalf("BankCard(%q) = %q, want %q", tt.card, got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "single rune", value: "张", want: "张"},
		{name: "two runes", value: "张三", want: "张*"},
		{name: "three runes", value: "张三丰", want: "张**"},
		{name: "ascii", value: "Alice", want: "A****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Name(tt.value); got != tt.want {
				t.Fatalf("Name(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "empty", email: "", want: ""},
		{name: "missing at", email: "alice.example.com", want: "alice.example.com"},
		{name: "empty local", email: "@example.com", want: "@example.com"},
		{name: "short local", email: "ab@example.com", want: "a****@example.com"},
		{name: "three rune local", email: "abc@example.com", want: "a****@example.com"},
		{name: "long local", email: "alice@example.com", want: "ali****@example.com"},
		{name: "unicode local", email: "用户@example.com", want: "用****@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Email(tt.email); got != tt.want {
				t.Fatalf("Email(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestCompatibilityAliases(t *testing.T) {
	if got := MaskString("1234567890", 3, 2, "****"); got != String("1234567890", 3, 2, "****") {
		t.Fatalf("MaskString alias = %q, want %q", got, String("1234567890", 3, 2, "****"))
	}
	if got := MaskPhone("13812345678"); got != Phone("13812345678") {
		t.Fatalf("MaskPhone alias = %q, want %q", got, Phone("13812345678"))
	}
	if got := MaskEmail("alice@example.com"); got != Email("alice@example.com") {
		t.Fatalf("MaskEmail alias = %q, want %q", got, Email("alice@example.com"))
	}
	if got := MaskIDCard("110101199001011234"); got != IDCard("110101199001011234") {
		t.Fatalf("MaskIDCard alias = %q, want %q", got, IDCard("110101199001011234"))
	}
	if got := MaskBankCard("6222021234567890123"); got != BankCard("6222021234567890123") {
		t.Fatalf("MaskBankCard alias = %q, want %q", got, BankCard("6222021234567890123"))
	}
	if got := MaskName("张三丰"); got != Name("张三丰") {
		t.Fatalf("MaskName alias = %q, want %q", got, Name("张三丰"))
	}
}
