package alipay

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/go-pay/gopay"
)

// PayeeIdentityType 表示支付宝收款方标识类型。
type PayeeIdentityType string

const (
	// PayeeIdentityAlipayUserID 表示支付宝用户 ID。
	PayeeIdentityAlipayUserID PayeeIdentityType = "ALIPAY_USER_ID"
	// PayeeIdentityAlipayLogonID 表示支付宝登录号（手机号或邮箱）。
	PayeeIdentityAlipayLogonID PayeeIdentityType = "ALIPAY_LOGON_ID"
	// PayeeIdentityAlipayOpenID 表示支付宝 OpenID。
	PayeeIdentityAlipayOpenID PayeeIdentityType = "ALIPAY_OPEN_ID"
)

// TransferSceneReportInfo 表示转账场景上报信息。
type TransferSceneReportInfo struct {
	// InfoType 是支付宝为转账场景规定的信息类型。
	InfoType string `json:"info_type"`
	// InfoContent 是该信息类型对应的业务说明。
	InfoContent string `json:"info_content"`
}

// TransferRequest 表示支付宝单笔转账请求。
type TransferRequest struct {
	// TransferID 是商户转账单号，必须唯一。
	TransferID string
	// Amount 是转账金额，单位为分。
	Amount int64
	// PayeeIdentity 是收款方支付宝用户 ID、手机号或邮箱。
	PayeeIdentity string
	// PayeeIdentityType 指定 PayeeIdentity 的类型。
	PayeeIdentityType PayeeIdentityType
	// PayeeName 是收款方实名；非空时支付宝将校验姓名。
	PayeeName string
	// Title 是转账标题。
	Title string
	// Remark 是转账备注。
	Remark string
	// TransferSceneName 是商家平台申明的转账场景；可选，为空时不发送。
	TransferSceneName string
	// TransferSceneReportInfos 是转账场景上报信息；可选，为空时不发送。
	// 两个场景字段独立透传，是否需要及允许的值由支付宝商户场景决定。
	TransferSceneReportInfos []TransferSceneReportInfo
}

// TransferResult 表示支付宝单笔转账结果。
type TransferResult struct {
	// TransferID 是商户转账单号。
	TransferID string
	// OrderID 是支付宝转账单号。
	OrderID string
	// FundOrderID 是支付宝资金流水号。
	FundOrderID string
	// Amount 是转账金额，单位为分。
	Amount int64
	// Status 是支付宝返回的原始转账状态。
	Status string
	// FailureCode 是转账失败码。
	FailureCode string
	// FailureReason 是转账失败原因。
	FailureReason string
	// TransferredAt 是支付宝记录的转账时间。
	TransferredAt *time.Time
	// ArrivedAt 是预计或实际到账时间。
	ArrivedAt *time.Time
}

// Transfer 发起支付宝账户单笔转账。
func (a *Alipay) Transfer(ctx context.Context, req *TransferRequest) (*TransferResult, error) {
	if err := validateTransferRequest(ctx, req); err != nil {
		return nil, err
	}
	payee := gopay.BodyMap{
		"identity":      req.PayeeIdentity,
		"identity_type": string(req.PayeeIdentityType),
	}
	bm := gopay.BodyMap{
		"out_biz_no":   req.TransferID,
		"trans_amount": centsToYuan(req.Amount),
		"product_code": "TRANS_ACCOUNT_NO_PWD",
		"biz_scene":    "DIRECT_TRANSFER",
		"payee_info":   payee,
	}
	if req.TransferSceneName != "" {
		bm.Set("transfer_scene_name", req.TransferSceneName)
	}
	if len(req.TransferSceneReportInfos) > 0 {
		bm.Set("transfer_scene_report_infos", req.TransferSceneReportInfos)
	}
	if req.PayeeName != "" {
		payee.Set("name", req.PayeeName)
	}
	if req.Title != "" {
		bm.Set("order_title", req.Title)
	}
	if req.Remark != "" {
		bm.Set("remark", req.Remark)
	}
	resp, err := a.gateway.transfer(ctx, bm)
	if err != nil {
		return nil, providerError("transfer", err)
	}
	if resp == nil || resp.Response == nil {
		return nil, providerError("transfer", fmt.Errorf("empty transfer response"))
	}
	transferredAt, err := parseAlipayTime(resp.Response.TransDate)
	if err != nil {
		return nil, providerError("transfer", err)
	}
	return &TransferResult{
		TransferID:    resp.Response.OutBizNo,
		OrderID:       resp.Response.OrderId,
		FundOrderID:   resp.Response.PayFundOrderId,
		Amount:        req.Amount,
		Status:        resp.Response.Status,
		TransferredAt: transferredAt,
	}, nil
}

// QueryTransfer 按商户转账单号查询转账状态。
// 请求结果不明时必须复用原 TransferID 查询或重试，不能创建新单号。
func (a *Alipay) QueryTransfer(ctx context.Context, transferID string) (*TransferResult, error) {
	if err := payment.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if transferID == "" {
		return nil, fmt.Errorf("%w: transfer ID cannot be empty", payment.ErrInvalidRequest)
	}
	resp, err := a.gateway.queryTransfer(ctx, gopay.BodyMap{
		"out_biz_no":   transferID,
		"product_code": "TRANS_ACCOUNT_NO_PWD",
		"biz_scene":    "DIRECT_TRANSFER",
	})
	if err != nil {
		return nil, providerError("query_transfer", err)
	}
	if resp == nil || resp.Response == nil {
		return nil, providerError("query_transfer", fmt.Errorf("empty transfer query response"))
	}
	amount, err := yuanToCents(resp.Response.TransAmount)
	if err != nil {
		return nil, providerError("query_transfer", err)
	}
	transferredAt, err := parseAlipayTime(resp.Response.PayDate)
	if err != nil {
		return nil, providerError("query_transfer", err)
	}
	arrivedAt, err := parseAlipayTime(resp.Response.ArrivalTimeEnd)
	if err != nil {
		return nil, providerError("query_transfer", err)
	}
	return &TransferResult{
		TransferID:    resp.Response.OutBizNo,
		OrderID:       resp.Response.OrderId,
		FundOrderID:   resp.Response.PayFundOrderId,
		Amount:        amount,
		Status:        resp.Response.Status,
		FailureCode:   resp.Response.ErrorCode,
		FailureReason: resp.Response.FailReason,
		TransferredAt: transferredAt,
		ArrivedAt:     arrivedAt,
	}, nil
}

func validateTransferRequest(ctx context.Context, req *TransferRequest) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("%w: transfer request cannot be nil", payment.ErrInvalidRequest)
	}
	if req.TransferID == "" {
		return fmt.Errorf("%w: transfer ID cannot be empty", payment.ErrInvalidRequest)
	}
	if req.Amount < 10 || req.Amount > 10_000_000_000 {
		return fmt.Errorf("%w: transfer amount must be between 10 and 10000000000 cents", payment.ErrInvalidRequest)
	}
	if req.PayeeIdentity == "" {
		return fmt.Errorf("%w: payee identity cannot be empty", payment.ErrInvalidRequest)
	}
	if req.PayeeIdentityType == "" {
		return fmt.Errorf("%w: payee identity type cannot be empty", payment.ErrInvalidRequest)
	}
	switch req.PayeeIdentityType {
	case PayeeIdentityAlipayUserID, PayeeIdentityAlipayLogonID, PayeeIdentityAlipayOpenID:
	default:
		return fmt.Errorf("%w: unsupported payee identity type %q", payment.ErrInvalidRequest, req.PayeeIdentityType)
	}
	if req.PayeeIdentityType == PayeeIdentityAlipayLogonID && req.PayeeName == "" {
		return fmt.Errorf("%w: payee name is required for ALIPAY_LOGON_ID", payment.ErrInvalidRequest)
	}
	for _, info := range req.TransferSceneReportInfos {
		if info.InfoType == "" || info.InfoContent == "" {
			return fmt.Errorf("%w: transfer scene report info cannot be empty", payment.ErrInvalidRequest)
		}
	}
	return nil
}
