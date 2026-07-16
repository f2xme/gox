package onepay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/f2xme/gox/payment"
)

// 微信 JSAPI 调起页默认文案（Config.WechatPage 零值时使用）。
//
// 注意：DefaultWechatLoadingText 现为「支付中…」（旧版为「正在调起微信支付…」）。
// 依赖精确旧文案的 E2E/监控需同步更新。
const (
	DefaultWechatTitle       = "微信支付"
	DefaultWechatLoadingText = "支付中…"
	DefaultWechatSuccessText = "支付已提交，请等待结果确认"
	DefaultWechatFailText    = "未完成支付，请返回重试"
)

// WechatBridgeCSP 返回调起页固定 Content-Security-Policy。
// 自定义 Template 必须兼容该策略：无外链 script/style/img；inline script 仅 nonce 匹配。
func WechatBridgeCSP(nonce string) string {
	return "default-src 'none'; script-src 'nonce-" + nonce + "'; base-uri 'none'; frame-ancestors 'none'"
}

// WechatBridgeData 是默认/自定义微信调起页模板的数据。
//
//   - Title / LoadingText：仅用于 HTML 正文/标题，由 html/template 转义
//   - SuccessText / FailText：仅用于 <script> 内（JSON 字符串字面量，含引号）；勿写入 HTML 正文
//   - Params：JSAPI 六元组；在 script 上下文中由 html/template 按 json tag 输出
//   - Nonce：须写入 script 的 nonce 属性，并与响应 CSP 一致
type WechatBridgeData struct {
	Nonce       string
	Params      *payment.JSAPIResult
	Title       string
	LoadingText string
	SuccessText template.JS
	FailText    template.JS
}

// 默认微信调起页。文案字段来自 WechatBridgeData。
var defaultBridgeTemplate = template.Must(template.New("wechat-pay").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}}</title></head>
<body><main><p id="status">{{.LoadingText}}</p></main>
<script nonce="{{.Nonce}}">
const pay = {{.Params}};
function invokePay(){
  WeixinJSBridge.invoke('getBrandWCPayRequest', pay, function(res){
    document.getElementById('status').textContent = res.err_msg === 'get_brand_wcpay_request:ok' ? {{.SuccessText}} : {{.FailText}};
  });
}
if(typeof WeixinJSBridge === 'undefined'){
  document.addEventListener('WeixinJSBridgeReady', invokePay, false);
}else{ invokePay(); }
</script></body></html>`))

var errorTemplate = template.Must(template.New("error").Parse(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>支付提示</title></head><body><main><h1>支付未能继续</h1><p>{{.Message}}</p><small>请求编号：{{.RequestID}}</small></main></body></html>`))

// resolveWechatPage 规范化文案：TrimSpace；空则填默认。Template 指针原样保留。
// normalizeConfig 调用一次即可；buildWechatBridgeData 再调一次保证幂等安全。
func resolveWechatPage(page WechatPage) WechatPage {
	page.Title = strings.TrimSpace(page.Title)
	if page.Title == "" {
		page.Title = DefaultWechatTitle
	}
	page.LoadingText = strings.TrimSpace(page.LoadingText)
	if page.LoadingText == "" {
		page.LoadingText = DefaultWechatLoadingText
	}
	page.SuccessText = strings.TrimSpace(page.SuccessText)
	if page.SuccessText == "" {
		page.SuccessText = DefaultWechatSuccessText
	}
	page.FailText = strings.TrimSpace(page.FailText)
	if page.FailText == "" {
		page.FailText = DefaultWechatFailText
	}
	return page
}

func buildWechatBridgeData(nonce string, params *payment.JSAPIResult, page WechatPage) (WechatBridgeData, error) {
	page = resolveWechatPage(page)
	successJSON, err := json.Marshal(page.SuccessText)
	if err != nil {
		return WechatBridgeData{}, fmt.Errorf("marshal wechat success text: %w", err)
	}
	failJSON, err := json.Marshal(page.FailText)
	if err != nil {
		return WechatBridgeData{}, fmt.Errorf("marshal wechat fail text: %w", err)
	}
	return WechatBridgeData{
		Nonce:       nonce,
		Params:      params,
		Title:       page.Title,
		LoadingText: page.LoadingText,
		SuccessText: template.JS(successJSON),
		FailText:    template.JS(failJSON),
	}, nil
}

func wechatBridgeTemplate(page WechatPage) *template.Template {
	if page.Template != nil {
		return page.Template
	}
	return defaultBridgeTemplate
}

// renderWechatBridge 只渲染模板，不写响应。失败时调用方仍可 writeError。
func renderWechatBridge(data WechatBridgeData, tpl *template.Template) ([]byte, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// commitWechatBridge 写入已渲染的调起页。调用后响应可能已提交，不得再 writeError。
func commitWechatBridge(w http.ResponseWriter, nonce string, body []byte) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", WechatBridgeCSP(nonce))
	w.Header().Set("Cache-Control", "no-store")
	_, err := w.Write(body)
	return err
}
