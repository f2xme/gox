package captcha

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockStore 是用于测试的简单内存存储。
type mockStore struct {
	data map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string]string),
	}
}

func (s *mockStore) Set(ctx context.Context, id string, answer string, ttl time.Duration) error {
	s.data[id] = answer
	return nil
}

func (s *mockStore) Get(ctx context.Context, id string) (string, error) {
	answer, ok := s.data[id]
	if !ok {
		return "", ErrNotFound
	}
	return answer, nil
}

func (s *mockStore) Take(ctx context.Context, id string) (string, error) {
	answer, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	delete(s.data, id)
	return answer, nil
}

func (s *mockStore) Delete(ctx context.Context, id string) error {
	delete(s.data, id)
	return nil
}

func (s *mockStore) Exists(id string) bool {
	_, ok := s.data[id]
	return ok
}

type mockGenerator struct {
	data   string
	answer string
	count  int
}

func (g *mockGenerator) Generate(ctx context.Context) (ChallengeData, error) {
	g.count++
	if err := ctx.Err(); err != nil {
		return ChallengeData{}, err
	}
	if g.data == "" {
		return ChallengeData{Data: "captcha-data", Answer: "answer"}, nil
	}
	return ChallengeData{Data: g.data, Answer: g.answer}, nil
}

func (g *mockGenerator) Type() string {
	return "mock"
}

func newTestService(t *testing.T, store Store, opts ...Option) Service {
	t.Helper()

	opts = append([]Option{WithGenerator(&mockGenerator{})}, opts...)
	svc, err := New(store, opts...)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return svc
}

func TestService_Generate(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if challenge.ID == "" {
		t.Error("Generate() returned empty id")
	}
	if challenge.Data == "" {
		t.Error("Generate() returned empty data")
	}
	if challenge.Type != "mock" {
		t.Errorf("Type = %v, want mock", challenge.Type)
	}
	if !store.Exists(challenge.ID) {
		t.Error("answer should be stored")
	}
}

func TestService_Verify(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	answer, err := store.Get(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	tests := []struct {
		name   string
		id     string
		answer string
		want   bool
	}{
		{name: "correct answer", id: challenge.ID, answer: answer, want: true},
		{name: "wrong answer", id: challenge.ID, answer: "wrong", want: false},
		{name: "empty id", id: "", answer: answer, want: false},
		{name: "empty answer", id: challenge.ID, answer: "", want: false},
		{name: "missing id", id: "missing", answer: answer, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "correct answer" {
				store.data[challenge.ID] = answer
			}

			got, err := svc.Verify(ctx, tt.id, tt.answer)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ConsumeOnSuccess(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	ok, err := svc.Verify(ctx, challenge.ID, "answer")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Fatal("Verify() = false, want true")
	}
	if store.Exists(challenge.ID) {
		t.Fatal("captcha should be deleted after successful verification")
	}
}

func TestService_ConsumeOnSuccessKeepsFailure(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	ok, err := svc.Verify(ctx, challenge.ID, "wrong")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Fatal("Verify() = true, want false")
	}
	if !store.Exists(challenge.ID) {
		t.Fatal("captcha should remain after failed verification")
	}
}

func TestService_ConsumeAlways(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store, WithConsumeMode(ConsumeAlways))

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	ok, err := svc.Verify(ctx, challenge.ID, "wrong")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Fatal("Verify() = true, want false")
	}
	if store.Exists(challenge.ID) {
		t.Fatal("captcha should be deleted when ConsumeAlways is used")
	}
}

func TestService_ConsumeNever(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store, WithConsumeMode(ConsumeNever))

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	ok, err := svc.Verify(ctx, challenge.ID, "answer")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Fatal("Verify() = false, want true")
	}
	if !store.Exists(challenge.ID) {
		t.Fatal("captcha should remain when ConsumeNever is used")
	}
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if err := svc.Delete(ctx, challenge.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.Exists(challenge.ID) {
		t.Fatal("captcha should not exist after deletion")
	}
}

func TestService_Regenerate(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := &mockGenerator{data: "first-data", answer: "first-answer"}
	svc, err := New(store, WithGenerator(gen))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	gen.data = "second-data"
	gen.answer = "second-answer"
	regenerated, err := svc.Regenerate(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("Regenerate() error = %v", err)
	}

	if regenerated.ID != challenge.ID {
		t.Errorf("Regenerate() ID = %v, want %v", regenerated.ID, challenge.ID)
	}
	if regenerated.Data != "second-data" {
		t.Errorf("Regenerate() Data = %v, want second-data", regenerated.Data)
	}
	answer, err := store.Get(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if answer != "second-answer" {
		t.Errorf("stored answer = %v, want second-answer", answer)
	}
}

func TestService_RegenerateInvalidID(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	svc := newTestService(t, store)

	if _, err := svc.Regenerate(ctx, ""); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Regenerate() error = %v, want ErrInvalidID", err)
	}
	if _, err := svc.Regenerate(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Regenerate() error = %v, want ErrNotFound", err)
	}
}

func TestService_WithTTL(t *testing.T) {
	store := newMockStore()
	customTTL := 10 * time.Minute
	svc := newTestService(t, store, WithTTL(customTTL))

	impl := svc.(*service)
	if impl.opts.TTL != customTTL {
		t.Errorf("TTL = %v, want %v", impl.opts.TTL, customTTL)
	}
}

func TestService_WithIDLength(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	customLength := 21
	svc := newTestService(t, store, WithIDLength(customLength))

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(challenge.ID) != customLength {
		t.Errorf("ID length = %v, want %v", len(challenge.ID), customLength)
	}
}

func TestService_WithGenerator(t *testing.T) {
	ctx := context.Background()
	store := newMockStore()
	gen := &mockGenerator{data: "custom-data", answer: "custom-answer"}

	svc, err := New(store, WithGenerator(gen))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if challenge.Data != "custom-data" {
		t.Errorf("Data = %v, want custom-data", challenge.Data)
	}
}
