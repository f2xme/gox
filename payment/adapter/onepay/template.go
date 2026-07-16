package onepay

import (
	"encoding/json"
	"html/template"
	"strings"

	"github.com/f2xme/gox/payment"
)

// 微信 JSAPI 调起页默认文案（Config.WechatPage 零值时使用）。
const (
	DefaultWechatLoadingText = "支付中…"
	DefaultWechatSuccessText = "支付已提交，请等待结果确认"
	DefaultWechatFailText    = "未完成支付，请返回重试"
)

// WechatBridgeData 是默认/自定义微信调起页模板的数据。
//
//   - LoadingText：HTML 正文，已由 html/template 转义
//   - SuccessText / FailText：JSON 字符串字面量（含引号），可安全嵌入 script
//   - Params：JSAPI 六元组；在 script 上下文中由 html/template 输出
type WechatBridgeData struct {
	Nonce       string
	Params      *payment.JSAPIResult
	LoadingText string
	SuccessText template.JS
	FailText    template.JS
}

// 默认微信调起页。文案字段来自 WechatBridgeData。
var defaultBridgeTemplate = template.Must(template.New("wechat-pay").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>微信支付</title></head>
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

func resolveWechatPage(page WechatPage) WechatPage {
	if strings.TrimSpace(page.LoadingText) == "" {
		page.LoadingText = DefaultWechatLoadingText
	}
	if strings.TrimSpace(page.SuccessText) == "" {
		page.SuccessText = DefaultWechatSuccessText
	}
	if strings.TrimSpace(page.FailText) == "" {
		page.FailText = DefaultWechatFailText
	}
	return page
}

func buildWechatBridgeData(nonce string, params *payment.JSAPIResult, page WechatPage) (WechatBridgeData, error) {
	page = resolveWechatPage(page)
	successJSON, err := json.Marshal(page.SuccessText)
	if err != nil {
		return WechatBridgeData{}, err
	}
	failJSON, err := json.Marshal(page.FailText)
	if err != nil {
		return WechatBridgeData{}, err
	}
	return WechatBridgeData{
		Nonce:       nonce,
		Params:      params,
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
