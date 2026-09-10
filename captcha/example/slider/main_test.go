package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
)

func TestSlideHTTP(t *testing.T) {
	store := memory.New(memory.WithCleanupInterval(0))
	svc, err := captcha.NewSlide(store)
	if err != nil {
		t.Fatal(err)
	}
	h := handler(svc)
	for _, tc := range []struct {
		name string
		body string
		code int
		pass bool
	}{
		{"completed", `{"token":%q,"distance":280,"duration":700,"track":[{"x":0,"t":0},{"x":35,"t":100},{"x":160,"t":350},{"x":280,"t":700}]}`, 200, true},
		{"no track", `{"token":%q,"distance":280,"duration":700}`, 200, false},
		{"old fixed answer", `{"token":%q,"answer":"100"}`, 400, false},
		{"wrong type", `{"token":%q,"duration":"700"}`, 400, false},
		{"trailing JSON", `{"token":%q}{}`, 400, false},
		{"large body", `{"token":%q,"track":[` + strings.Repeat(`{"x":0,"t":0},`, 3000) + `{"x":280,"t":700}]}`, 400, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generated := httptest.NewRecorder()
			h.ServeHTTP(generated, httptest.NewRequest(http.MethodGet, "/api/captcha/generate", nil))
			var result struct {
				Success bool                   `json:"success"`
				Data    captcha.SlideChallenge `json:"data"`
			}
			if err := json.Unmarshal(generated.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if !result.Success || len(result.Data.Token) != 32 || result.Data.Distance != 280 {
				t.Fatalf("unexpected challenge: %+v", result)
			}
			created := strconv.FormatInt(time.Now().Add(-time.Second).UnixMilli(), 10)
			if err := store.Set(context.Background(), result.Data.Token, created, time.Minute); err != nil {
				t.Fatal(err)
			}
			body := fmt.Sprintf(tc.body, result.Data.Token)
			response := httptest.NewRecorder()
			h.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/captcha/verify", strings.NewReader(body)))
			if response.Code != tc.code {
				t.Fatalf("verify: %d %s, want %d", response.Code, response.Body.String(), tc.code)
			}
			if tc.code == http.StatusOK {
				var verified struct{ Success bool }
				if err := json.Unmarshal(response.Body.Bytes(), &verified); err != nil || verified.Success != tc.pass {
					t.Fatalf("verification = %s, %v", response.Body.String(), err)
				}
				if _, err := store.Get(context.Background(), result.Data.Token); err != captcha.ErrNotFound {
					t.Fatalf("token was not consumed: %v", err)
				}
			}
		})
	}
}
