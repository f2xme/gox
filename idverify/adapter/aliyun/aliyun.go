package aliyun

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cloudauth "github.com/alibabacloud-go/cloudauth-20190307/v4/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"

	"github.com/f2xme/gox/idverify"
)

const (
	bizMatch    = "1"
	bizMismatch = "2"
	bizNoRecord = "3"
	apiOK       = "200"
)

// callResult 单节点调用解析结果。
type callResult struct {
	HTTPStatus int32
	Code       string
	Message    string
	BizCode    string
	RequestID  string
}

// caller 可注入，便于单测。
type caller func(ctx context.Context, endpoint, name, idNumber string) (callResult, error)

// Verifier 阿里云 Id2MetaVerify 实现。
type Verifier struct {
	options Options
	call    caller

	clientsMu sync.Mutex
	clients   map[string]*cloudauth.Client
}

var _ idverify.Verifier = (*Verifier)(nil)

// Provider 返回 aliyun。
func (v *Verifier) Provider() string { return idverify.ProviderAliyun }

// withCaller 覆盖底层调用，仅测试使用。
func (v *Verifier) withCaller(fn caller) *Verifier {
	if fn != nil {
		v.call = fn
	}
	return v
}

// Verify 主备 endpoint 轮询调用。
func (v *Verifier) Verify(ctx context.Context, req idverify.Request) (idverify.Result, error) {
	start := time.Now()
	if ctx == nil {
		return idverify.Result{Duration: time.Since(start)}, fmt.Errorf("%w: context is nil", idverify.ErrInvalidArgument)
	}
	if err := ctx.Err(); err != nil {
		return idverify.Result{Duration: time.Since(start)}, err
	}

	req = req.Normalize()
	if req.Name == "" || req.IDNumber == "" {
		return idverify.Result{Provider: idverify.ProviderAliyun, Duration: time.Since(start)},
			fmt.Errorf("%w: name and id number are required", idverify.ErrInvalidArgument)
	}

	callFn := v.call
	if callFn == nil {
		callFn = v.sdkCall
	}

	var lastSys error
	for _, ep := range v.options.Endpoints {
		if err := ctx.Err(); err != nil {
			return idverify.Result{Duration: time.Since(start)}, err
		}
		got, err := callFn(ctx, ep, req.Name, req.IDNumber)
		if err != nil {
			lastSys = err
			continue
		}
		if !httpOK(got.HTTPStatus) {
			lastSys = fmt.Errorf("endpoint %s http_status=%d code=%s", ep, got.HTTPStatus, got.Code)
			continue
		}
		if got.Code == apiOK {
			return mapBiz(got, start)
		}
		// 参数非法属于用户输入/业务侧问题，不重试其它节点，也不当系统不可用。
		if res, ok := mapParamIllegal(got, start); ok {
			return res, nil
		}
		if isClientFault(got.Code) {
			return idverify.Result{Provider: idverify.ProviderAliyun, Duration: time.Since(start)},
				idverify.Wrap(idverify.ProviderAliyun, "verify",
					fmt.Errorf("%w: code=%s message=%s requestId=%s", idverify.ErrUnavailable, got.Code, got.Message, got.RequestID))
		}
		lastSys = fmt.Errorf("endpoint %s code=%s message=%s", ep, got.Code, got.Message)
	}
	if lastSys != nil {
		return idverify.Result{Duration: time.Since(start)},
			idverify.Wrap(idverify.ProviderAliyun, "verify", fmt.Errorf("%w: %v", idverify.ErrUnavailable, lastSys))
	}
	return idverify.Result{Duration: time.Since(start)},
		idverify.Wrap(idverify.ProviderAliyun, "verify", idverify.ErrUnavailable)
}

func mapBiz(got callResult, start time.Time) (idverify.Result, error) {
	base := idverify.Result{
		Provider:     idverify.ProviderAliyun,
		ProviderCode: got.BizCode,
		RequestID:    got.RequestID,
		Duration:     time.Since(start),
	}
	switch got.BizCode {
	case bizMatch:
		base.Matched = true
		return base, nil
	case bizMismatch:
		base.ErrorCode = idverify.CodeNameMismatch
		base.ErrorMessage = "name and id number mismatch"
		return base, nil
	case bizNoRecord:
		base.ErrorCode = idverify.CodeIDInvalid
		base.ErrorMessage = "id number not found"
		return base, nil
	default:
		return base, idverify.Wrap(idverify.ProviderAliyun, "verify",
			fmt.Errorf("%w: unexpected biz_code=%s message=%s", idverify.ErrUnavailable, got.BizCode, got.Message))
	}
}

func httpOK(status int32) bool {
	return status == 0 || status == 200
}

func isClientFault(code string) bool {
	switch code {
	case "400", "401", "402", "403", "410", "411", "412":
		return true
	default:
		return false
	}
}

// mapParamIllegal 将阿里云「参数非法」类响应映射为业务不匹配结果。
// 例：code=401 message=参数非法(identifyNum) → CodeIDInvalid（非 ErrUnavailable）。
func mapParamIllegal(got callResult, start time.Time) (idverify.Result, bool) {
	msg := strings.TrimSpace(got.Message)
	if msg == "" {
		return idverify.Result{}, false
	}
	// 阿里云常见：参数非法(identifyNum) / 参数非法(userName)
	if !strings.Contains(msg, "参数非法") && !strings.Contains(strings.ToLower(msg), "invalid parameter") {
		return idverify.Result{}, false
	}

	base := idverify.Result{
		Provider:     idverify.ProviderAliyun,
		ProviderCode: got.Code,
		RequestID:    got.RequestID,
		Duration:     time.Since(start),
	}
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "identifynum"), strings.Contains(lower, "identify_num"), strings.Contains(msg, "身份证"):
		base.ErrorCode = idverify.CodeIDInvalid
		base.ErrorMessage = "invalid id number"
		return base, true
	case strings.Contains(lower, "username"), strings.Contains(lower, "user_name"), strings.Contains(msg, "姓名"):
		// 姓名参数非法：按信息不匹配处理（用户可改后重试）
		base.ErrorCode = idverify.CodeNameMismatch
		base.ErrorMessage = "invalid name"
		return base, true
	default:
		// 其它参数非法：证件侧兜底，避免误报系统错误
		base.ErrorCode = idverify.CodeIDInvalid
		base.ErrorMessage = msg
		return base, true
	}
}

func (v *Verifier) clientFor(endpoint string) (*cloudauth.Client, error) {
	v.clientsMu.Lock()
	defer v.clientsMu.Unlock()
	if v.clients == nil {
		v.clients = make(map[string]*cloudauth.Client)
	}
	if c, ok := v.clients[endpoint]; ok {
		return c, nil
	}
	timeoutMs := int(v.options.Timeout / time.Millisecond)
	if timeoutMs <= 0 {
		timeoutMs = int(defaultTimeout / time.Millisecond)
	}
	cfg := &openapiutil.Config{
		AccessKeyId:     dara.String(v.options.AccessKeyID),
		AccessKeySecret: dara.String(v.options.AccessKeySecret),
		Endpoint:        dara.String(endpoint),
		ConnectTimeout:  dara.Int(timeoutMs),
		ReadTimeout:     dara.Int(timeoutMs),
	}
	client, err := cloudauth.NewClient(cfg)
	if err != nil {
		return nil, idverify.Wrap(idverify.ProviderAliyun, "client", err)
	}
	v.clients[endpoint] = client
	return client, nil
}

func (v *Verifier) sdkCall(ctx context.Context, endpoint, name, idNumber string) (callResult, error) {
	client, err := v.clientFor(endpoint)
	if err != nil {
		return callResult{}, err
	}
	timeoutMs := int(v.options.Timeout / time.Millisecond)
	if timeoutMs <= 0 {
		timeoutMs = int(defaultTimeout / time.Millisecond)
	}
	req := &cloudauth.Id2MetaVerifyRequest{
		ParamType:   dara.String(paramTypeNormal),
		UserName:    dara.String(name),
		IdentifyNum: dara.String(idNumber),
	}
	runtime := &dara.RuntimeOptions{}
	runtime.SetConnectTimeout(timeoutMs)
	runtime.SetReadTimeout(timeoutMs)

	type out struct {
		resp *cloudauth.Id2MetaVerifyResponse
		err  error
	}
	ch := make(chan out, 1)
	go func() {
		resp, err := client.Id2MetaVerifyWithOptions(req, runtime)
		ch <- out{resp: resp, err: err}
	}()

	select {
	case <-ctx.Done():
		return callResult{}, idverify.Wrap(idverify.ProviderAliyun, "call", ctx.Err())
	case o := <-ch:
		if o.err != nil {
			return callResult{}, idverify.Wrap(idverify.ProviderAliyun, "call", fmt.Errorf("%w: %w", idverify.ErrUnavailable, o.err))
		}
		return parseResp(o.resp)
	}
}

func parseResp(resp *cloudauth.Id2MetaVerifyResponse) (callResult, error) {
	if resp == nil {
		return callResult{}, idverify.Wrap(idverify.ProviderAliyun, "call", fmt.Errorf("empty response"))
	}
	out := callResult{}
	if resp.StatusCode != nil {
		out.HTTPStatus = *resp.StatusCode
	}
	if resp.Body == nil {
		return out, idverify.Wrap(idverify.ProviderAliyun, "call", fmt.Errorf("empty body"))
	}
	if resp.Body.Code != nil {
		out.Code = strings.TrimSpace(*resp.Body.Code)
	}
	if resp.Body.Message != nil {
		out.Message = strings.TrimSpace(*resp.Body.Message)
	}
	if resp.Body.RequestId != nil {
		out.RequestID = strings.TrimSpace(*resp.Body.RequestId)
	}
	if resp.Body.ResultObject != nil && resp.Body.ResultObject.BizCode != nil {
		out.BizCode = strings.TrimSpace(*resp.Body.ResultObject.BizCode)
	}
	return out, nil
}
