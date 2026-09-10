package base64

import (
	"context"
	"errors"
	"fmt"
	"testing"

	upstream "github.com/mojocn/base64Captcha"
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
			gen, err := New(WithType(tt.typ))
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			data, err := gen.Generate(context.Background())
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if data.Data == "" {
				t.Error("Generate() returned empty data")
			}

			if data.Answer == "" {
				t.Error("Generate() returned empty answer")
			}

			if gen.Type() != "base64" {
				t.Errorf("Type() = %v, want base64", gen.Type())
			}
		})
	}
}

type questionDriver struct {
	upstream.Driver
	drawn string
}

func (*questionDriver) GenerateIdQuestionAnswer() (string, string, string) {
	return "id", "1+2=?", "3"
}

func (d *questionDriver) DrawCaptcha(question string) (upstream.Item, error) {
	d.drawn = question
	return d.Driver.DrawCaptcha(question)
}

func TestDrawQuestion(t *testing.T) {
	opts := defaultOptions()
	opts.Type = TypeMath
	driver := &questionDriver{Driver: createDriver(opts)}
	gen := &base64Generator{driver: driver}
	data, err := gen.Generate(context.Background())
	if err != nil || driver.drawn != "1+2=?" || data.Answer != "3" {
		t.Fatalf("question=%q answer=%q error=%v", driver.drawn, data.Answer, err)
	}
}

func TestImageLayoutValidation(t *testing.T) {
	for _, tc := range []struct {
		typ                   CaptchaType
		width, height, length int
		valid                 bool
	}{
		{TypeDigit, 1, 1, 4, false},
		{TypeDigit, 10, 80, 4, false},
		{TypeDigit, 240, 10, 4, false},
		{TypeDigit, 240, 80, 1000, false},
		{TypeDigit, 240, 80, 4, true},
		{TypeDigit, 300, 100, 6, true},
		{TypeString, 19, 80, 4, false},
		{TypeString, 240, 15, 4, false},
		{TypeMath, 19, 80, 4, false},
		{TypeMath, 240, 15, 4, false},
		{TypeString, 20, 16, 4, true},
		{TypeMath, 20, 16, 4, true},
	} {
		t.Run(fmt.Sprintf("%s/%dx%d/%d", tc.typ, tc.width, tc.height, tc.length), func(t *testing.T) {
			gen, err := New(WithType(tc.typ), WithSize(tc.width, tc.height), WithLength(tc.length))
			if !tc.valid {
				if !errors.Is(err, ErrInvalidSize) {
					t.Fatalf("New() = %v, want ErrInvalidSize", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := gen.Generate(context.Background()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	gen, err := New(
		WithType(TypeDigit),
		WithSize(300, 100),
		WithLength(6),
		WithNoiseCount(5),
		WithLanguage("zh"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

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
