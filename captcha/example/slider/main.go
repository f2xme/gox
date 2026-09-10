// 滑动验证示例：go run ./captcha/example/slider，然后打开 http://127.0.0.1:8080。
package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/f2xme/gox/captcha"
	"github.com/f2xme/gox/captcha/adapter/memory"
)

//go:embed index.html
var page []byte

func main() {
	store := memory.New(memory.WithMaxSize(1000))
	defer store.(io.Closer).Close()
	svc, err := captcha.NewSlide(store)
	if err != nil {
		log.Print(err)
		return
	}
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           handler(svc),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       time.Minute,
	}
	log.Printf("滑动验证示例：http://%s", server.Addr)
	log.Print(server.ListenAndServe())
}

func handler(svc *captcha.SlideVerifier) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	mux.HandleFunc("GET /api/captcha/generate", func(w http.ResponseWriter, r *http.Request) {
		challenge, err := svc.Generate(r.Context())
		if err != nil {
			http.Error(w, "生成失败，请重试", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true, "data": challenge,
		})
	})
	mux.HandleFunc("DELETE /api/captcha/{token}", func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Delete(r.Context(), r.PathValue("token")); err != nil {
			http.Error(w, "刷新失败，请重试", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/captcha/verify", func(w http.ResponseWriter, r *http.Request) {
		var req captcha.SlideVerifyData
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil || req.Token == "" {
			http.Error(w, "验证参数无效", http.StatusBadRequest)
			return
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			http.Error(w, "验证参数无效", http.StatusBadRequest)
			return
		}
		ok, err := svc.Verify(r.Context(), req)
		if err != nil {
			http.Error(w, "验证失败，请刷新重试", http.StatusInternalServerError)
			return
		}
		result := struct {
			Success bool   `json:"success"`
			Message string `json:"message,omitempty"`
		}{Success: ok}
		if !ok {
			result.Message = "滑动未通过或已过期，请刷新后平稳拖动重试"
		}
		// 仅在 ok 为 true 时执行需要确认的业务操作；示例不实际发送短信。
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(result)
	})
	return mux
}
