package gin

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/f2xme/gox/httpx"
	ginframework "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)


type ginContext struct {
	c *ginframework.Context
}

var _ httpx.Context = (*ginContext)(nil)

func (ctx *ginContext) Request() *http.Request       { return ctx.c.Request }
func (ctx *ginContext) Param(key string) httpx.Value  { return httpx.Value(ctx.c.Param(key)) }
func (ctx *ginContext) Query(key string) httpx.Value  { return httpx.Value(ctx.c.Query(key)) }
func (ctx *ginContext) QueryAll(key string) []string  { return ctx.c.QueryArray(key) }
func (ctx *ginContext) Header(key string) httpx.Value { return httpx.Value(ctx.c.GetHeader(key)) }
func (ctx *ginContext) Cookie(name string) (*http.Cookie, error) {
	return ctx.c.Request.Cookie(name)
}
func (ctx *ginContext) ClientIP() string { return ctx.c.ClientIP() }
func (ctx *ginContext) Method() string   { return ctx.c.Request.Method }
func (ctx *ginContext) Path() string     { return ctx.c.Request.URL.Path }

func (ctx *ginContext) Bind(v any) error {
	if err := ctx.c.ShouldBind(v); err != nil {
		return translateError(err, v)
	}
	return validate(v)
}

func (ctx *ginContext) BindJSON(v any) error {
	if err := ctx.c.ShouldBindJSON(v); err != nil {
		return translateError(err, v)
	}
	return validate(v)
}

func (ctx *ginContext) BindQuery(v any) error {
	if err := ctx.c.ShouldBindQuery(v); err != nil {
		return translateError(err, v)
	}
	return validate(v)
}

func (ctx *ginContext) BindForm(v any) error {
	if err := ctx.c.ShouldBindWith(v, binding.Form); err != nil {
		return translateError(err, v)
	}
	return validate(v)
}

// validate 检查 v 是否实现了 Validator 接口,如果实现则调用 Validate()。
func validate(v any) error {
	if validator, ok := v.(httpx.Validator); ok {
		return validator.Validate()
	}
	return nil
}

// translateError 将 validator 错误翻译为带 label 的中文错误消息。
func translateError(err error, obj any) error {
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	// 获取结构体类型和字段标签映射
	typ := reflect.TypeOf(obj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var errMsgs []string
	for _, fieldErr := range validationErrs {
		fieldName := fieldErr.Field()
		label := fieldName

		// 查找 label tag
		if field, found := typ.FieldByName(fieldName); found {
			if labelTag := field.Tag.Get("label"); labelTag != "" {
				label = labelTag
			}
		}

		// 翻译验证规则
		msg := translateValidationTag(label, fieldErr)
		errMsgs = append(errMsgs, msg)
	}

	if len(errMsgs) == 0 {
		return err
	}

	return fmt.Errorf("%s", strings.Join(errMsgs, "; "))
}

// translateValidationTag 翻译验证标签为中文错误消息。
func translateValidationTag(label string, fieldErr validator.FieldError) string {
	tag := fieldErr.Tag()
	param := fieldErr.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", label)
	case "min":
		switch fieldErr.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s长度不能小于%s", label, param)
		default:
			return fmt.Sprintf("%s不能小于%s", label, param)
		}
	case "max":
		switch fieldErr.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s长度不能大于%s", label, param)
		default:
			return fmt.Sprintf("%s不能大于%s", label, param)
		}
	case "email":
		return fmt.Sprintf("%s必须是有效的邮箱地址", label)
	case "url":
		return fmt.Sprintf("%s必须是有效的URL", label)
	case "len":
		return fmt.Sprintf("%s长度必须为%s", label, param)
	case "eq":
		return fmt.Sprintf("%s必须等于%s", label, param)
	case "ne":
		return fmt.Sprintf("%s不能等于%s", label, param)
	case "gt":
		return fmt.Sprintf("%s必须大于%s", label, param)
	case "gte":
		return fmt.Sprintf("%s必须大于等于%s", label, param)
	case "lt":
		return fmt.Sprintf("%s必须小于%s", label, param)
	case "lte":
		return fmt.Sprintf("%s必须小于等于%s", label, param)
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", label, param)
	case "alpha":
		return fmt.Sprintf("%s只能包含字母", label)
	case "alphanum":
		return fmt.Sprintf("%s只能包含字母和数字", label)
	case "numeric":
		return fmt.Sprintf("%s必须是数字", label)
	case "number":
		return fmt.Sprintf("%s必须是有效的数字", label)
	case "hexadecimal":
		return fmt.Sprintf("%s必须是十六进制字符串", label)
	case "hexcolor":
		return fmt.Sprintf("%s必须是十六进制颜色代码", label)
	case "rgb":
		return fmt.Sprintf("%s必须是RGB颜色代码", label)
	case "rgba":
		return fmt.Sprintf("%s必须是RGBA颜色代码", label)
	case "hsl":
		return fmt.Sprintf("%s必须是HSL颜色代码", label)
	case "hsla":
		return fmt.Sprintf("%s必须是HSLA颜色代码", label)
	case "uuid":
		return fmt.Sprintf("%s必须是有效的UUID", label)
	case "uuid3":
		return fmt.Sprintf("%s必须是有效的UUID v3", label)
	case "uuid4":
		return fmt.Sprintf("%s必须是有效的UUID v4", label)
	case "uuid5":
		return fmt.Sprintf("%s必须是有效的UUID v5", label)
	case "ascii":
		return fmt.Sprintf("%s只能包含ASCII字符", label)
	case "contains":
		return fmt.Sprintf("%s必须包含'%s'", label, param)
	case "containsany":
		return fmt.Sprintf("%s必须包含'%s'中的任意字符", label, param)
	case "excludes":
		return fmt.Sprintf("%s不能包含'%s'", label, param)
	case "excludesall":
		return fmt.Sprintf("%s不能包含'%s'中的任何字符", label, param)
	case "startswith":
		return fmt.Sprintf("%s必须以'%s'开头", label, param)
	case "endswith":
		return fmt.Sprintf("%s必须以'%s'结尾", label, param)
	case "ip":
		return fmt.Sprintf("%s必须是有效的IP地址", label)
	case "ipv4":
		return fmt.Sprintf("%s必须是有效的IPv4地址", label)
	case "ipv6":
		return fmt.Sprintf("%s必须是有效的IPv6地址", label)
	case "json":
		return fmt.Sprintf("%s必须是有效的JSON", label)
	case "datetime":
		return fmt.Sprintf("%s必须是有效的日期时间格式", label)
	default:
		return fmt.Sprintf("%s验证失败(%s)", label, tag)
	}
}

func (ctx *ginContext) JSON(code int, v any) error {
	ctx.c.JSON(code, v)
	return nil
}

func (ctx *ginContext) String(code int, s string) error {
	ctx.c.String(code, "%s", s)
	return nil
}

func (ctx *ginContext) HTML(code int, html string) error {
	ctx.c.Data(code, "text/html; charset=utf-8", []byte(html))
	return nil
}

func (ctx *ginContext) Blob(code int, contentType string, data []byte) error {
	ctx.c.Data(code, contentType, data)
	return nil
}

func (ctx *ginContext) NoContent(code int) error {
	ctx.c.Status(code)
	return nil
}

func (ctx *ginContext) Redirect(code int, url string) error {
	ctx.c.Redirect(code, url)
	return nil
}

func (ctx *ginContext) SetHeader(key, value string) { ctx.c.Header(key, value) }

func (ctx *ginContext) SetCookie(cookie *http.Cookie) {
	ctx.c.SetCookie(
		cookie.Name, cookie.Value, cookie.MaxAge,
		cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly,
	)
}

func (ctx *ginContext) Status(code int) { ctx.c.Status(code) }

func (ctx *ginContext) Set(key string, value any)    { ctx.c.Set(key, value) }
func (ctx *ginContext) Get(key string) (any, bool)   { return ctx.c.Get(key) }
func (ctx *ginContext) MustGet(key string) any       { return ctx.c.MustGet(key) }
func (ctx *ginContext) ResponseWriter() http.ResponseWriter { return ctx.c.Writer }
func (ctx *ginContext) Raw() any                     { return ctx.c }
