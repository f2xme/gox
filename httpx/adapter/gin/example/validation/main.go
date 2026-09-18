// 多语言校验示例：在仓库根目录运行 go run ./httpx/adapter/gin/example/validation。
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	"github.com/f2xme/gox/validator"
	"golang.org/x/text/language"
)

//go:embed messages.json
var messagesJSON []byte

// CreateUserRequest 演示原始 Go 字段路径与业务翻译键的对应关系。
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

func main() {
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(app.Start("127.0.0.1:8080"))
}

func newApp() (httpx.Engine, error) {
	var messages map[string]map[string]string
	if err := json.Unmarshal(messagesJSON, &messages); err != nil {
		return nil, fmt.Errorf("加载业务语言包: %w", err)
	}
	// 顺序与 matcher 一致；无匹配时回退到第一种语言。
	languages := []string{"zh", "en"}
	matcher := language.NewMatcher([]language.Tag{language.SimplifiedChinese, language.English})

	app := ginadapter.New()
	app.SetErrorHandler(func(c httpx.Context, err error) {
		ve, ok := validator.AsValidationError(err)
		if !ok {
			httpx.DefaultErrorHandler(c, err)
			return
		}
		_, index := language.MatchStrings(matcher, string(c.Header("Accept-Language")))
		lang := languages[index]
		catalog := messages[lang]
		var texts []string
		for _, field := range ve.Fields() {
			// 不依赖 label、json 标签或 validator 自带的消息语言。
			message := catalog[field.StructNamespace+"."+field.Tag]
			if message == "" {
				message = catalog["validation.failed"]
			}
			texts = append(texts, strings.ReplaceAll(message, "{param}", field.Param))
		}
		c.SetHeader("Content-Language", lang)
		c.ResponseWriter().Header().Add("Vary", "Accept-Language")
		httpx.WriteError(c, http.StatusBadRequest, strings.Join(texts, "; "))
	})
	app.POST("/users", func(c httpx.Context) error {
		var req CreateUserRequest
		if err := c.BindJSON(&req); err != nil {
			// 保留底层错误，让统一处理器提取 ValidationError。
			return httpx.ErrBadRequest("请求参数无效", err)
		}
		// 示例只校验，不实际创建用户。
		return c.NoContent(http.StatusNoContent)
	})
	return app, nil
}
