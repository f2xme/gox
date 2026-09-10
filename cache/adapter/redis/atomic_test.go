package redis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
	"github.com/f2xme/gox/captcha"
	cachecaptcha "github.com/f2xme/gox/captcha/adapter/cache"
	"github.com/f2xme/gox/captcha/generator/base64"
)

func TestAtomicOperations(t *testing.T) {
	c, mr := setupTestRedis(t)
	defer c.(cache.Closer).Close()
	ctx := context.Background()
	a := c.(cache.AtomicStore)
	for _, tc := range []struct {
		name      string
		ttl, want time.Duration
	}{
		{"keep TTL", cache.KeepTTL, time.Minute},
		{"replace TTL", 2 * time.Minute, 2 * time.Minute},
		{"remove TTL", cache.NoExpiration, 0},
		{"submillisecond", time.Nanosecond, time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.Set(ctx, "key", []byte{}, time.Minute); err != nil {
				t.Fatal(err)
			}
			if ok, err := a.CompareAndDelete(ctx, "key", []byte("wrong")); err != nil || ok {
				t.Fatalf("wrong delete = %v, %v", ok, err)
			}
			if ok, err := a.CompareAndSwap(ctx, "key", []byte("wrong"), nil, tc.ttl); err != nil || ok {
				t.Fatalf("wrong CAS = %v, %v", ok, err)
			}
			if ok, err := a.CompareAndSwap(ctx, "key", []byte{}, []byte{0, 255}, tc.ttl); err != nil || !ok {
				t.Fatalf("CAS = %v, %v", ok, err)
			}
			if got := mr.TTL("key"); got != tc.want {
				t.Fatalf("TTL = %v, want %v", got, tc.want)
			}
			if got, err := a.Take(ctx, "key"); err != nil || string(got) != string([]byte{0, 255}) {
				t.Fatalf("Take = %v, %v", got, err)
			}
			if _, err := a.Take(ctx, "key"); !errors.Is(err, cache.ErrNotFound) {
				t.Fatalf("second Take = %v", err)
			}
		})
	}
	if ok, err := a.CompareAndSwap(ctx, "key", nil, nil, time.Minute); err != nil || ok {
		t.Fatalf("CAS created missing key = %v, %v", ok, err)
	}
	if _, err := a.CompareAndSwap(ctx, "key", nil, nil, -time.Second); !errors.Is(err, cache.ErrInvalidTTL) {
		t.Fatalf("invalid TTL = %v", err)
	}
	if err := c.Set(ctx, "key", nil, time.Second); err != nil {
		t.Fatal(err)
	}
	mr.FastForward(2 * time.Second)
	if ok, err := a.CompareAndDelete(ctx, "key", nil); err != nil || ok {
		t.Fatalf("expired delete = %v, %v", ok, err)
	}
	if ok, err := a.CompareAndSwap(ctx, "key", nil, nil, time.Minute); err != nil || ok {
		t.Fatalf("revived expired key = %v, %v", ok, err)
	}
}

type captchaReadBarrier struct {
	captcha.AtomicStore
	ready, resume chan struct{}
}

func (s *captchaReadBarrier) Get(ctx context.Context, id string) (string, error) {
	value, err := s.AtomicStore.Get(ctx, id)
	s.ready <- struct{}{}
	<-s.resume
	return value, err
}

func TestCaptchaSharedRedis(t *testing.T) {
	first, mr := setupTestRedis(t)
	defer first.(cache.Closer).Close()
	second, err := New(WithAddr(mr.Addr()))
	if err != nil {
		t.Fatal(err)
	}
	defer second.(cache.Closer).Close()
	ctx := context.Background()
	gen, err := base64.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []captcha.ConsumeMode{captcha.ConsumeOnSuccess, captcha.ConsumeAlways} {
		ready, resume := make(chan struct{}, 2), make(chan struct{})
		var successes atomic.Int32
		var wg sync.WaitGroup
		for _, backend := range []cache.Store{first, second} {
			store := cachecaptcha.New(backend, cachecaptcha.WithPrefix("atomic-test:"))
			if err := store.Set(ctx, "id", "1234", time.Minute); err != nil {
				t.Fatal(err)
			}
			wrapped := &captchaReadBarrier{AtomicStore: store.(captcha.AtomicStore), ready: ready, resume: resume}
			svc, err := captcha.New(wrapped, captcha.WithGenerator(gen), captcha.WithConsumeMode(mode))
			if err != nil {
				t.Fatal(err)
			}
			wg.Go(func() {
				if mode == captcha.ConsumeAlways {
					ready <- struct{}{}
					<-resume
				}
				ok, err := svc.Verify(ctx, "id", "1234")
				if err != nil {
					t.Error(err)
				}
				if ok {
					successes.Add(1)
				}
			})
		}
		<-ready
		<-ready
		close(resume)
		wg.Wait()
		if successes.Load() != 1 {
			t.Fatalf("mode %v: successes = %d, want 1", mode, successes.Load())
		}
	}
}
