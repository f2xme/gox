package captcha_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f2xme/gox/cache"
	cachememory "github.com/f2xme/gox/cache/adapter/memory"
	"github.com/f2xme/gox/captcha"
	cacheadapter "github.com/f2xme/gox/captcha/adapter/cache"
	"github.com/f2xme/gox/captcha/adapter/memory"
)

type blockedGenerator struct {
	entered chan struct{}
	resume  chan struct{}
}

func (*blockedGenerator) Type() string { return "test" }
func (g *blockedGenerator) Generate(context.Context) (captcha.ChallengeData, error) {
	if g.entered != nil {
		close(g.entered)
		<-g.resume
	}
	return captcha.ChallengeData{Data: "image", Answer: "1234"}, nil
}

// 固定两次 Get 都读到旧答案后再继续，避免依赖调度概率触发竞态。
type barrierStore struct {
	captcha.AtomicStore
	ready  chan struct{}
	resume chan struct{}
}

func (s *barrierStore) Get(ctx context.Context, id string) (string, error) {
	value, err := s.AtomicStore.Get(ctx, id)
	s.ready <- struct{}{}
	<-s.resume
	return value, err
}

func TestConcurrentConsumptionAndRegeneration(t *testing.T) {
	ctx := context.Background()
	backend, err := cachememory.New(cachememory.WithMaxSize(100))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backend.(cache.Closer).Close() })
	stores := map[string]captcha.Store{
		"memory": memory.New(memory.WithCleanupInterval(0)),
		"cache":  cacheadapter.New(backend),
	}
	for name, base := range stores {
		t.Run(name, func(t *testing.T) {
			if _, err := captcha.New(struct{ captcha.Store }{base}, captcha.WithGenerator(&blockedGenerator{})); !errors.Is(err, captcha.ErrAtomicStoreRequired) {
				t.Fatalf("non-atomic store: %v", err)
			}
			t.Run("verify once across services", func(t *testing.T) {
				store := &barrierStore{AtomicStore: base.(captcha.AtomicStore), ready: make(chan struct{}, 2), resume: make(chan struct{})}
				if err := base.Set(ctx, "id", "1234", time.Minute); err != nil {
					t.Fatal(err)
				}
				var wg sync.WaitGroup
				var successes atomic.Int32
				for range 2 {
					svc, err := captcha.New(store, captcha.WithGenerator(&blockedGenerator{}))
					if err != nil {
						t.Fatal(err)
					}
					wg.Go(func() {
						ok, err := svc.Verify(ctx, "id", " 1234 ")
						if err != nil {
							t.Error(err)
						}
						if ok {
							successes.Add(1)
						}
					})
				}
				<-store.ready
				<-store.ready
				close(store.resume)
				wg.Wait()
				if successes.Load() != 1 {
					t.Fatalf("successes = %d, want 1", successes.Load())
				}
			})
			for _, action := range []string{"consume", "delete", "replace"} {
				t.Run("regenerate after "+action, func(t *testing.T) {
					gen := &blockedGenerator{entered: make(chan struct{}), resume: make(chan struct{})}
					svc, err := captcha.New(base, captcha.WithGenerator(gen), captcha.WithConsumeMode(captcha.ConsumeAlways))
					if err != nil {
						t.Fatal(err)
					}
					if err := base.Set(ctx, "id", "1234", time.Minute); err != nil {
						t.Fatal(err)
					}
					finished := make(chan error, 1)
					go func() { _, err := svc.Regenerate(ctx, "id"); finished <- err }()
					<-gen.entered
					switch action {
					case "consume":
						if ok, err := svc.Verify(ctx, "id", "1234"); err != nil || !ok {
							t.Errorf("Verify() = %v, %v", ok, err)
						}
					case "delete":
						err = svc.Delete(ctx, "id")
					case "replace":
						err = base.Set(ctx, "id", "new-answer", time.Minute)
					}
					close(gen.resume)
					if err != nil {
						t.Fatal(err)
					}
					if err := <-finished; !errors.Is(err, captcha.ErrNotFound) {
						t.Fatalf("Regenerate() = %v, want ErrNotFound", err)
					}
					value, err := base.Get(ctx, "id")
					if action == "replace" {
						if err != nil || value != "new-answer" {
							t.Fatalf("replacement changed: %q, %v", value, err)
						}
					} else if !errors.Is(err, captcha.ErrNotFound) {
						t.Fatalf("consumed ID resurrected: %q, %v", value, err)
					}
				})
			}
		})
	}
}
