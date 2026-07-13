package onepay

import "html/template"

var bridgeTemplate = template.Must(template.New("wechat-pay").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>微信支付</title></head>
<body><main><p id="status">正在调起微信支付…</p></main>
<script nonce="{{.Nonce}}">
const pay = {{.Params}};
function invokePay(){
  WeixinJSBridge.invoke('getBrandWCPayRequest', pay, function(res){
    document.getElementById('status').textContent = res.err_msg === 'get_brand_wcpay_request:ok' ? '支付已提交，请等待结果确认' : '未完成支付，请返回重试';
  });
}
if(typeof WeixinJSBridge === 'undefined'){
  document.addEventListener('WeixinJSBridgeReady', invokePay, false);
}else{ invokePay(); }
</script></body></html>`))

var errorTemplate = template.Must(template.New("error").Parse(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>支付提示</title></head><body><main><h1>支付未能继续</h1><p>{{.Message}}</p><small>请求编号：{{.RequestID}}</small></main></body></html>`))
