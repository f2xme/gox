package pager

import (
	"net/http"
	"strconv"
)

// NewOffsetFromRequest 从查询参数（limit、offset）创建 OffsetPage
func NewOffsetFromRequest(r *http.Request) OffsetPage {
	query := r.URL.Query()

	limit := parseIntParam(query.Get("limit"), DefaultLimit)
	offset := parseIntParam(query.Get("offset"), DefaultOffset)

	return NewOffset(limit, offset)
}

// NewPageFromRequest 从查询参数（page、size）创建 PageNumber
func NewPageFromRequest(r *http.Request) PageNumber {
	query := r.URL.Query()

	page := parseIntParam(query.Get("page"), DefaultPage)
	size := parseIntParam(query.Get("size"), DefaultSize)

	return NewPage(page, size)
}

// NewCursorFromRequest 从查询参数（cursor、limit）创建 CursorPage
func NewCursorFromRequest(r *http.Request) CursorPage {
	query := r.URL.Query()

	cursor := query.Get("cursor")
	limit := parseIntParam(query.Get("limit"), DefaultLimit)

	return NewCursor(cursor, limit)
}

// parseIntParam 将字符串参数解析为 int，如果无效则返回 defaultValue
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(param)
	if err != nil || value < 0 {
		return defaultValue
	}

	return value
}
