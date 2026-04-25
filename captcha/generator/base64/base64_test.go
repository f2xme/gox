package base64

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		typ  CaptchaType
	}{
		{"digit", TypeDigit},
		{"string", TypeString},
		{"math", TypeMath},
		{"audio", TypeAudio},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := New(WithType(tt.typ))

			data, answer, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if data == "" {
				t.Error("Generate() returned empty data")
			}

			if answer == "" {
				t.Error("Generate() returned empty answer")
			}

			if gen.Type() != "base64" {
				t.Errorf("Type() = %v, want base64", gen.Type())
			}
		})
	}
}

func TestOptions(t *testing.T) {
	gen := New(
		WithType(TypeDigit),
		WithSize(300, 100),
		WithLength(6),
		WithNoiseCount(5),
		WithLanguage("zh"),
	)

	g := gen.(*base64Generator)

	if g.opts.Type != TypeDigit {
		t.Errorf("Type = %v, want %v", g.opts.Type, TypeDigit)
	}
	if g.opts.Width != 300 {
		t.Errorf("Width = %v, want 300", g.opts.Width)
	}
	if g.opts.Height != 100 {
		t.Errorf("Height = %v, want 100", g.opts.Height)
	}
	if g.opts.Length != 6 {
		t.Errorf("Length = %v, want 6", g.opts.Length)
	}
	if g.opts.NoiseCount != 5 {
		t.Errorf("NoiseCount = %v, want 5", g.opts.NoiseCount)
	}
	if g.opts.Language != "zh" {
		t.Errorf("Language = %v, want zh", g.opts.Language)
	}
}

func TestCaptchaTypeString(t *testing.T) {
	tests := []struct {
		typ  CaptchaType
		want string
	}{
		{TypeDigit, "digit"},
		{TypeString, "string"},
		{TypeMath, "math"},
		{TypeAudio, "audio"},
		{CaptchaType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
