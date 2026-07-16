package mock

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/f2xme/gox/payment"
)

// PaymentRecord 保存一次支付的输入、结果和当前状态。
type PaymentRecord struct {
	// Order 是创建支付时的订单副本。
	Order payment.Order
	// Result 是创建支付后返回的结果副本。
	Result payment.PaymentResult
	// Status 是当前支付状态。
	Status payment.PaymentStatus
	// PaidAt 是支付成功时间。
	PaidAt *time.Time
}

// RefundRecord 保存一次退款的输入和当前结果。
type RefundRecord struct {
	// Request 是创建退款时的请求副本。
	Request payment.RefundRequest
	// Result 是当前退款结果。
	Result payment.RefundResult
}

// Call 保存一次有效 mock 操作调用。
type Call struct {
	// Operation 是调用的操作。
	Operation Operation
	// OrderID 是关联商户订单号。
	OrderID string
	// RefundID 是关联商户退款单号。
	RefundID string
	// CalledAt 是调用时间。
	CalledAt time.Time
}

// Client 是并发安全的内存支付测试客户端。
type Client struct {
	clockMu         sync.Mutex
	mu              sync.RWMutex
	options         Options
	operationErrors map[Operation]error
	payments        map[string]*PaymentRecord
	refunds         map[string]*RefundRecord
	calls           []Call
}

var (
	_ payment.Payment         = (*Client)(nil)
	_ payment.PaymentNotifier = (*Client)(nil)
	_ payment.RefundNotifier  = (*Client)(nil)
)

// New 创建内存支付测试客户端。
func New(opts ...Option) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	errorsByOperation := make(map[Operation]error, len(options.OperationErrors))
	for operation, err := range options.OperationErrors {
		if err != nil {
			errorsByOperation[operation] = err
		}
	}
	options.OperationErrors = nil
	return &Client{
		options:         options,
		operationErrors: errorsByOperation,
		payments:        make(map[string]*PaymentRecord),
		refunds:         make(map[string]*RefundRecord),
	}, nil
}

// Pay 创建一条 mock 支付记录。
func (c *Client) Pay(ctx context.Context, order *payment.Order) (*payment.PaymentResult, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateOrder(order); err != nil {
		return nil, err
	}
	if err := validateExtra(order.Extra); err != nil {
		return nil, err
	}
	if err := c.begin(ctx, OperationPay, order.OrderID, ""); err != nil {
		return nil, err
	}

	clonedOrder := cloneOrder(*order)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.payments[order.OrderID]; exists {
		return nil, fmt.Errorf("%w: mock order %q already exists", payment.ErrInvalidRequest, order.OrderID)
	}
	result := payment.PaymentResult{
		OrderID:       order.OrderID,
		TransactionID: "mock-pay-" + order.OrderID,
		PayURL:        "mock://pay/" + url.PathEscape(order.OrderID),
		Extra:         cloneMap(order.Extra),
	}
	record := &PaymentRecord{Order: clonedOrder, Result: clonePaymentResult(result), Status: c.options.PaymentStatus}
	if record.Status == payment.PaymentStatusSuccess || record.Status == payment.PaymentStatusRefunded {
		record.PaidAt = cloneTimeValue(c.now())
	}
	c.payments[order.OrderID] = record
	returned := clonePaymentResult(result)
	return &returned, nil
}

// Query 返回 mock 支付当前状态。
func (c *Client) Query(ctx context.Context, orderID string) (*payment.QueryResult, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return nil, err
	}
	if err := c.begin(ctx, OperationQuery, orderID, ""); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	record, exists := c.payments[orderID]
	if !exists {
		return nil, fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, orderID)
	}
	return &payment.QueryResult{
		OrderID:       record.Order.OrderID,
		TransactionID: record.Result.TransactionID,
		Status:        record.Status,
		Amount:        record.Order.Amount,
		PaidAt:        cloneTime(record.PaidAt),
	}, nil
}

// Refund 创建一条 mock 退款记录。
func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResult, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if err := payment.ValidateRefundRequest(req); err != nil {
		return nil, err
	}
	if err := c.begin(ctx, OperationRefund, req.OrderID, req.RefundID); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.refunds[req.RefundID]; exists {
		return nil, fmt.Errorf("%w: mock refund %q already exists", payment.ErrInvalidRequest, req.RefundID)
	}
	order, exists := c.payments[req.OrderID]
	if !exists {
		return nil, fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, req.OrderID)
	}
	if order.Status != payment.PaymentStatusSuccess && order.Status != payment.PaymentStatusRefunded {
		return nil, fmt.Errorf("%w: mock order %q is not paid", payment.ErrInvalidRequest, req.OrderID)
	}
	if req.OriginalAmount != order.Order.Amount {
		return nil, fmt.Errorf("%w: mock original amount does not match order", payment.ErrInvalidRequest)
	}
	activeRefundAmount := c.activeRefundAmountLocked(req.OrderID, "")
	if req.Amount > order.Order.Amount-activeRefundAmount {
		return nil, fmt.Errorf("%w: mock cumulative refund exceeds order amount", payment.ErrInvalidRequest)
	}
	result := payment.RefundResult{
		RefundID:      req.RefundID,
		TransactionID: "mock-refund-" + req.RefundID,
		Status:        c.options.RefundStatus,
	}
	if result.Status == payment.RefundStatusSuccess {
		result.RefundAt = cloneTimeValue(c.now())
	}
	record := &RefundRecord{Request: *req, Result: cloneRefundResult(result)}
	c.refunds[req.RefundID] = record
	c.reconcilePaymentRefundStatusLocked(req.OrderID)
	returned := cloneRefundResult(result)
	return &returned, nil
}

// Close 关闭 pending 状态的 mock 支付。
func (c *Client) Close(ctx context.Context, orderID string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}
	if err := c.begin(ctx, OperationClose, orderID, ""); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	record, exists := c.payments[orderID]
	if !exists {
		return fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, orderID)
	}
	switch record.Status {
	case payment.PaymentStatusPending:
		record.Status = payment.PaymentStatusClosed
		return nil
	case payment.PaymentStatusClosed:
		return nil
	default:
		return fmt.Errorf("%w: mock order %q cannot be closed from %q", payment.ErrInvalidRequest, orderID, record.Status)
	}
}

// SetPaymentStatus 修改指定支付状态。
func (c *Client) SetPaymentStatus(orderID string, status payment.PaymentStatus) error {
	if err := payment.ValidateOrderID(orderID); err != nil {
		return err
	}
	if !validPaymentStatus(status) {
		return fmt.Errorf("%w: invalid mock payment status %q", payment.ErrInvalidRequest, status)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	record, exists := c.payments[orderID]
	if !exists {
		return fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, orderID)
	}
	record.Status = status
	if status == payment.PaymentStatusSuccess || status == payment.PaymentStatusRefunded {
		if record.PaidAt == nil {
			record.PaidAt = cloneTimeValue(c.now())
		}
	} else {
		record.PaidAt = nil
	}
	return nil
}

// SetRefundStatus 修改指定退款状态。
func (c *Client) SetRefundStatus(refundID string, status payment.RefundStatus) error {
	if refundID == "" {
		return fmt.Errorf("%w: refund ID cannot be empty", payment.ErrInvalidRequest)
	}
	if !validRefundStatus(status) {
		return fmt.Errorf("%w: invalid mock refund status %q", payment.ErrInvalidRequest, status)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.setRefundStatusLocked(refundID, status)
}

// setRefundStatusLocked 在已持有 c.mu 时修改退款状态并协调订单退款汇总。
func (c *Client) setRefundStatusLocked(refundID string, status payment.RefundStatus) error {
	record, exists := c.refunds[refundID]
	if !exists {
		return fmt.Errorf("%w: mock refund %q not found", payment.ErrInvalidRequest, refundID)
	}
	if activeRefundStatus(status) {
		order, exists := c.payments[record.Request.OrderID]
		if !exists {
			return fmt.Errorf("%w: mock order %q not found", payment.ErrInvalidRequest, record.Request.OrderID)
		}
		activeAmount := c.activeRefundAmountLocked(record.Request.OrderID, refundID)
		if activeAmount > order.Order.Amount || record.Request.Amount > order.Order.Amount-activeAmount {
			return fmt.Errorf("%w: mock cumulative refund exceeds order amount", payment.ErrInvalidRequest)
		}
	}
	record.Result.Status = status
	if status == payment.RefundStatusSuccess {
		if record.Result.RefundAt == nil {
			record.Result.RefundAt = cloneTimeValue(c.now())
		}
	} else {
		record.Result.RefundAt = nil
	}
	c.reconcilePaymentRefundStatusLocked(record.Request.OrderID)
	return nil
}

// SetOperationError 设置或清除指定操作错误。
func (c *Client) SetOperationError(operation Operation, err error) error {
	if !validOperation(operation) {
		return fmt.Errorf("%w: invalid mock operation %q", payment.ErrInvalidConfig, operation)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		delete(c.operationErrors, operation)
	} else {
		c.operationErrors[operation] = err
	}
	return nil
}

// Payments 返回按订单号排序的支付记录副本。
func (c *Client) Payments() []PaymentRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	records := make([]PaymentRecord, 0, len(c.payments))
	for _, record := range c.payments {
		records = append(records, clonePaymentRecord(*record))
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Order.OrderID < records[j].Order.OrderID })
	return records
}

// Refunds 返回按退款单号排序的退款记录副本。
func (c *Client) Refunds() []RefundRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	records := make([]RefundRecord, 0, len(c.refunds))
	for _, record := range c.refunds {
		records = append(records, cloneRefundRecord(*record))
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Request.RefundID < records[j].Request.RefundID })
	return records
}

// Calls 返回按调用顺序保存的调用记录副本。
func (c *Client) Calls() []Call {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Call(nil), c.calls...)
}

// Reset 清空订单、退款和调用记录，保留配置与操作错误。
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.payments)
	clear(c.refunds)
	c.calls = nil
}

func (c *Client) begin(ctx context.Context, operation Operation, orderID, refundID string) error {
	calledAt := c.now()
	c.mu.Lock()
	c.calls = append(c.calls, Call{Operation: operation, OrderID: orderID, RefundID: refundID, CalledAt: calledAt})
	c.mu.Unlock()

	if c.options.Delay > 0 {
		timer := time.NewTimer(c.options.Delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return fmt.Errorf("mock payment %s: %w", operation, ctx.Err())
		case <-timer.C:
		}
	} else if err := ctx.Err(); err != nil {
		return fmt.Errorf("mock payment %s: %w", operation, err)
	}
	c.mu.RLock()
	err := c.operationErrors[operation]
	c.mu.RUnlock()
	return err
}

func (c *Client) activeRefundAmountLocked(orderID, excludedRefundID string) int64 {
	var total int64
	for refundID, refund := range c.refunds {
		if refundID == excludedRefundID || refund.Request.OrderID != orderID {
			continue
		}
		if activeRefundStatus(refund.Result.Status) {
			total += refund.Request.Amount
		}
	}
	return total
}

func activeRefundStatus(status payment.RefundStatus) bool {
	return status == payment.RefundStatusPending || status == payment.RefundStatusSuccess
}

func (c *Client) reconcilePaymentRefundStatusLocked(orderID string) {
	order, exists := c.payments[orderID]
	if !exists {
		return
	}
	var refunded int64
	for _, refund := range c.refunds {
		if refund.Request.OrderID == orderID && refund.Result.Status == payment.RefundStatusSuccess {
			refunded += refund.Request.Amount
		}
	}
	if refunded >= order.Order.Amount {
		order.Status = payment.PaymentStatusRefunded
	} else if order.Status == payment.PaymentStatusRefunded {
		order.Status = payment.PaymentStatusSuccess
	}
}

func validateContext(ctx context.Context) error {
	if err := payment.ValidateContext(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("mock payment context: %w", err)
	}
	return nil
}

func cloneTimeValue(value time.Time) *time.Time {
	return &value
}

func (c *Client) now() time.Time {
	c.clockMu.Lock()
	defer c.clockMu.Unlock()
	return c.options.Clock()
}
