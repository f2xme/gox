package testkit_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/f2xme/gox/httpx"
	"github.com/f2xme/gox/httpx/testkit"
)

func TestClient_JSONRequestAndAssertions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("X-Default"); got != "default-value" {
			t.Errorf("X-Default = %q, want default-value", got)
		}
		if got := r.Header.Get("X-Request"); got != "request-value" {
			t.Errorf("X-Request = %q, want request-value", got)
		}
		if got := r.URL.Query().Get("trace"); got != "abc" {
			t.Errorf("trace query = %q, want abc", got)
		}
		cookie, err := r.Cookie("sid")
		if err != nil {
			t.Errorf("sid cookie error = %v", err)
		} else if cookie.Value != "s1" {
			t.Errorf("sid cookie = %q, want s1", cookie.Value)
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if req.Name != "Alice" {
			t.Errorf("request name = %q, want Alice", req.Name)
		}

		w.Header().Set("X-Result", "created")
		http.SetCookie(w, &http.Cookie{Name: "token", Value: "t1"})
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"success":true,"data":{"id":123,"tags":["go","http"]}}`)
	})

	client := testkit.NewWithHandler(t, mux, testkit.WithDefaultHeader("X-Default", "default-value"))
	defer client.Close()

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	client.POSTJSON("/users", map[string]string{"name": "Alice"},
		testkit.WithHeader("X-Request", "request-value"),
		testkit.WithQuery("trace", "abc"),
		testkit.WithCookie(&http.Cookie{Name: "sid", Value: "s1"}),
	).
		ExpectStatus(http.StatusCreated).
		ExpectHeader("X-Result", "created").
		ExpectCookie("token", "t1").
		ExpectJSONValue("success", true).
		ExpectJSONValue("data.id", 123).
		ExpectJSONValue("data.tags[1]", "http").
		DecodeJSON(&resp)

	if !resp.Success || resp.Data.ID != 123 {
		t.Fatalf("decoded response = %+v, want success true with id 123", resp)
	}
}

func TestClient_PostForm(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want form content type", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if got := r.Form.Get("username"); got != "alice" {
			t.Errorf("username = %q, want alice", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client := testkit.NewWithHandler(t, mux)
	defer client.Close()

	client.POSTForm("/login", mapValues("username", "alice")).
		ExpectStatus(http.StatusNoContent)
}

func TestClient_HTTPMethods(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/method", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, r.Method)
	})

	client := testkit.NewWithHandler(t, mux)
	defer client.Close()

	client.GET("/method").ExpectStatus(http.StatusOK).ExpectBody(http.MethodGet)
	client.POST("/method", strings.NewReader("body")).ExpectStatus(http.StatusOK).ExpectBody(http.MethodPost)
	client.PUT("/method", strings.NewReader("body")).ExpectStatus(http.StatusOK).ExpectBody(http.MethodPut)
	client.PATCH("/method", strings.NewReader("body")).ExpectStatus(http.StatusOK).ExpectBody(http.MethodPatch)
	client.DELETE("/method").ExpectStatus(http.StatusOK).ExpectBody(http.MethodDelete)
	client.OPTIONS("/method").ExpectStatus(http.StatusOK).ExpectBody(http.MethodOptions)
	client.HEAD("/method").ExpectStatus(http.StatusNoContent).ExpectHeader("X-Method", http.MethodHead)
	client.Do(http.MethodTrace, "/method", nil).ExpectStatus(http.StatusOK).ExpectBody(http.MethodTrace)
}

func TestClient_PostMultipart(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Errorf("Content-Type = %q, want multipart/form-data", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1024); err != nil {
			t.Errorf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("kind"); got != "avatar" {
			t.Errorf("kind = %q, want avatar", got)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile() error = %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("ReadAll(file) error = %v", err)
		}
		if header.Filename != "avatar.txt" {
			t.Errorf("filename = %q, want avatar.txt", header.Filename)
		}
		if string(data) != "file-body" {
			t.Errorf("file body = %q, want file-body", string(data))
		}
		w.WriteHeader(http.StatusOK)
	})

	client := testkit.NewWithHandler(t, mux)
	defer client.Close()

	client.POSTMultipart("/upload",
		map[string]string{"kind": "avatar"},
		[]testkit.File{{
			FieldName:   "file",
			FileName:    "avatar.txt",
			ContentType: "text/plain",
			Reader:      strings.NewReader("file-body"),
		}},
	).ExpectStatus(http.StatusOK)
}

func TestClient_NewWithEngine(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "pong")
	})

	client := testkit.New(t, &fakeEngine{handler: mux})
	defer client.Close()

	client.GET("/ping").ExpectStatus(http.StatusOK).ExpectBody("pong")
}

func TestRunEngineSuite(t *testing.T) {
	calls := 0
	factories := []testkit.EngineFactory{
		{Name: "one", New: func(t testing.TB) httpx.Engine { return &fakeEngine{handler: okHandler()} }},
		{Name: "two", New: func(t testing.TB) httpx.Engine { return &fakeEngine{handler: okHandler()} }},
	}

	testkit.RunEngineSuite(t, factories, func(t *testing.T, client *testkit.Client) {
		calls++
		client.GET("/ok").ExpectStatus(http.StatusOK).ExpectBody("ok")
	})

	if calls != 2 {
		t.Fatalf("suite calls = %d, want 2", calls)
	}
}

func okHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})
	return mux
}

func mapValues(key, value string) map[string][]string {
	return map[string][]string{key: {value}}
}

type fakeEngine struct {
	handler http.Handler
}

func (e *fakeEngine) GET(string, httpx.Handler, ...httpx.Middleware)     {}
func (e *fakeEngine) POST(string, httpx.Handler, ...httpx.Middleware)    {}
func (e *fakeEngine) PUT(string, httpx.Handler, ...httpx.Middleware)     {}
func (e *fakeEngine) DELETE(string, httpx.Handler, ...httpx.Middleware)  {}
func (e *fakeEngine) PATCH(string, httpx.Handler, ...httpx.Middleware)   {}
func (e *fakeEngine) HEAD(string, httpx.Handler, ...httpx.Middleware)    {}
func (e *fakeEngine) OPTIONS(string, httpx.Handler, ...httpx.Middleware) {}
func (e *fakeEngine) Any(string, httpx.Handler, ...httpx.Middleware)     {}
func (e *fakeEngine) Group(string, ...httpx.Middleware) httpx.Router     { return e }
func (e *fakeEngine) Use(...httpx.Middleware)                            {}
func (e *fakeEngine) Static(string, string)                              {}
func (e *fakeEngine) StaticFile(string, string)                          {}
func (e *fakeEngine) Start(string) error                                 { return nil }
func (e *fakeEngine) Shutdown(context.Context) error                     { return nil }
func (e *fakeEngine) SetErrorHandler(httpx.ErrorHandler)                 {}
func (e *fakeEngine) SetNotFoundHandler(httpx.Handler)                   {}
func (e *fakeEngine) Raw() any                                           { return e.handler }
