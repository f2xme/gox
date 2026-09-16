package httpx

// Cookie 返回请求中指定 Cookie 的值；Cookie 不存在或读取失败时返回空字符串。
func Cookie(ctx Context, name string) string {
	cookie, err := ctx.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
