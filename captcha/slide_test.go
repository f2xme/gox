package captcha_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
)

func ExampleNewSlide() {
	store := memory.New(memory.WithCleanupInterval(0))
	slide, err := captcha.NewSlide(store)
	if err != nil {
		fmt.Println(err)
		return
	}
	challenge, err := slide.Generate(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	// 将 token 和目标距离交给前端，收到轨迹后调用 slide.Verify(ctx, data)。
	fmt.Println(len(challenge.Token), challenge.Distance)
	// Output: 32 280
}

func slideData(t *testing.T, slide *captcha.SlideVerifier, store captcha.Store) captcha.SlideVerifyData {
	t.Helper()
	challenge, err := slide.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 模拟已创建一秒的会话，避免每个用例等待最短滑动时长。
	created := strconv.FormatInt(time.Now().Add(-time.Second).UnixMilli(), 10)
	if err := store.Set(context.Background(), challenge.Token, created, time.Minute); err != nil {
		t.Fatal(err)
	}
	return captcha.SlideVerifyData{
		Token: challenge.Token, Distance: challenge.Distance, Duration: 700,
		Track: []captcha.TrackPoint{{X: 0, T: 0}, {X: 35, T: 100}, {X: 160, T: 350}, {X: 280, T: 700}},
	}
}

func TestSlideVerify(t *testing.T) {
	store := memory.New(memory.WithCleanupInterval(0))
	slide, err := captcha.NewSlide(store)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*captcha.SlideVerifyData)
		want   bool
	}{
		{"valid", func(d *captcha.SlideVerifyData) {}, true},
		{"tolerance", func(d *captcha.SlideVerifyData) { d.Distance = 278 }, true},
		{"too fast", func(d *captcha.SlideVerifyData) { d.Duration = 299 }, false},
		{"padded duration", func(d *captcha.SlideVerifyData) {
			d.Duration = 300
			d.Track = []captcha.TrackPoint{{X: 0, T: 0}, {X: 1, T: 1}, {X: 280, T: 2}}
		}, false},
		{"too slow", func(d *captcha.SlideVerifyData) { d.Duration = 10001 }, false},
		{"longer than session", func(d *captcha.SlideVerifyData) { d.Duration = 10000 }, false},
		{"incomplete", func(d *captcha.SlideVerifyData) { d.Distance = 200 }, false},
		{"NaN", func(d *captcha.SlideVerifyData) { d.Distance = math.NaN() }, false},
		{"Inf", func(d *captcha.SlideVerifyData) { d.Track[1].X = math.Inf(1) }, false},
		{"missing track", func(d *captcha.SlideVerifyData) { d.Track = nil }, false},
		{"few points", func(d *captcha.SlideVerifyData) { d.Track = d.Track[:2] }, false},
		{"many points", func(d *captcha.SlideVerifyData) {
			d.Duration = 512
			d.Track = make([]captcha.TrackPoint, 513)
			for i := range d.Track {
				fraction := float64(i) / 512
				d.Track[i] = captcha.TrackPoint{X: 280 * fraction * fraction, T: i}
			}
		}, false},
		{"wrong start", func(d *captcha.SlideVerifyData) { d.Track[0].X = 100 }, false},
		{"wrong end", func(d *captcha.SlideVerifyData) { d.Track[3].X = 220 }, false},
		{"outside track", func(d *captcha.SlideVerifyData) { d.Track[1].X = 400 }, false},
		{"negative coordinate", func(d *captcha.SlideVerifyData) { d.Track[1].X = -1 }, false},
		{"wrong start time", func(d *captcha.SlideVerifyData) { d.Track[0].T = 1 }, false},
		{"duplicate time", func(d *captcha.SlideVerifyData) { d.Track[2].T = 100 }, false},
		{"reversed time", func(d *captcha.SlideVerifyData) { d.Track[2].T = 50 }, false},
		{"future point", func(d *captcha.SlideVerifyData) { d.Track[3].T = 701 }, false},
		{"uniform speed", func(d *captcha.SlideVerifyData) {
			d.Track = []captcha.TrackPoint{{X: 0, T: 0}, {X: 140, T: 350}, {X: 280, T: 700}}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := slideData(t, slide, store)
			tc.change(&data)
			if ok, err := slide.Verify(context.Background(), data); err != nil || ok != tc.want {
				t.Fatalf("Verify() = %v, %v, want %v", ok, err, tc.want)
			}
			if ok, err := slide.Verify(context.Background(), data); err != nil || ok {
				t.Fatalf("replay = %v, %v", ok, err)
			}
		})
	}
}

func TestSlideSession(t *testing.T) {
	ctx := context.Background()
	store := memory.New(memory.WithCleanupInterval(0))
	slide, err := captcha.NewSlide(store)
	if err != nil {
		t.Fatal(err)
	}
	data := slideData(t, slide, store)
	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			ok, err := slide.Verify(ctx, data)
			if err != nil {
				t.Error(err)
			}
			if ok {
				successes.Add(1)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successes = %d, want 1", successes.Load())
	}
	data = slideData(t, slide, store)
	if err := store.Set(ctx, data.Token, "0", time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	if ok, err := slide.Verify(ctx, data); err != nil || ok {
		t.Fatalf("expired = %v, %v", ok, err)
	}
	data = slideData(t, slide, store)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := slide.Generate(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("Generate(canceled) = %v", err)
	}
	if _, err := slide.Verify(canceled, data); !errors.Is(err, context.Canceled) {
		t.Fatalf("Verify(canceled) = %v", err)
	}
	if ok, err := slide.Verify(ctx, data); err != nil || !ok {
		t.Fatalf("canceled request consumed token: %v, %v", ok, err)
	}
	data = slideData(t, slide, store)
	if err := slide.Delete(ctx, data.Token); err != nil {
		t.Fatal(err)
	}
	if ok, err := slide.Verify(ctx, data); err != nil || ok {
		t.Fatalf("deleted = %v, %v", ok, err)
	}
	data = slideData(t, slide, store)
	if err := store.Set(ctx, data.Token, "broken", time.Minute); err != nil {
		t.Fatal(err)
	}
	if ok, err := slide.Verify(ctx, data); ok || err == nil {
		t.Fatalf("corrupt session = %v, %v", ok, err)
	}
}

func TestSlideOptions(t *testing.T) {
	store := memory.New(memory.WithCleanupInterval(0))
	if _, err := captcha.NewSlide(nil); !errors.Is(err, captcha.ErrNilStore) {
		t.Fatalf("nil store = %v", err)
	}
	if _, err := captcha.NewSlide(struct{ captcha.Store }{store}); !errors.Is(err, captcha.ErrAtomicStoreRequired) {
		t.Fatalf("non-atomic store = %v", err)
	}
	for _, opt := range []captcha.SlideOption{
		captcha.WithSlideTTL(0), captcha.WithSlideDistance(0, 0),
		captcha.WithSlideDistance(280, -1), captcha.WithSlideDistance(280, 280),
		captcha.WithSlideDistance(math.NaN(), 0), captcha.WithSlideDuration(0, time.Second),
		captcha.WithSlideDuration(time.Second, time.Millisecond),
		captcha.WithSlideMinSpeedVariance(-1), captcha.WithSlideMinSpeedVariance(math.Inf(1)),
	} {
		if _, err := captcha.NewSlide(store, opt); !errors.Is(err, captcha.ErrInvalidSlideOptions) {
			t.Fatalf("invalid options = %v", err)
		}
	}
	slide, err := captcha.NewSlide(store, captcha.WithSlideDistance(100, 2), captcha.WithSlideMinSpeedVariance(0))
	if err != nil {
		t.Fatal(err)
	}
	data := slideData(t, slide, store)
	data.Track = []captcha.TrackPoint{{X: 0, T: 0}, {X: 50, T: 350}, {X: 100, T: 700}}
	if ok, err := slide.Verify(context.Background(), data); err != nil || !ok {
		t.Fatalf("custom distance / disabled variance = %v, %v", ok, err)
	}
}
