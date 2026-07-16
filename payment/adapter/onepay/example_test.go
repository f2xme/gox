package onepay_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/f2xme/gox/payment"
	"github.com/f2xme/gox/payment/adapter/onepay"
)

// ExampleNew_createCodeAndMount 演示创建一码付服务、生成二维码并挂载 Handler。
// Resolver / Wechat 使用进程内假实现，保证示例可运行且不出网。
func ExampleNew_createCodeAndMount() {
	resolver := newMemoryResolver()
	wechat := &staticWechatOAuth{openID: "oTestOpenID"}

	svc, err := onepay.New(onepay.Config{
		BaseURL:  "https://pay.example.com",
		TokenKey: bytes.Repeat([]byte{7}, 32),
		Resolver: resolver,
		Wechat:   wechat,
	}, onepay.WithQRSize(256))
	if err != nil {
		log.Fatal(err)
	}

	code, err := svc.CreateCode(context.Background(), "intent-1001")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/pay/", svc.Handler())
	// Gin: r.Any("/pay/*path", gin.WrapH(svc.Handler()))

	fmt.Println(len(code.PNG) > 0)
	fmt.Println(len(code.URL) > len("https://pay.example.com/pay/"))
	// Output:
	// true
	// true
}

// ExampleCheckoutResolver 演示业务侧 CheckoutResolver 的内存骨架契约。
//
// 本 Example 仅断言：主意图取金额、同 intent+provider 复用、双平台不同订单号、微信 OpenID 冲突。
// 骨架细节（creating/retiring、slot uncertain、intent paid 单调、ctx 取消不落墓碑）由 TestMemoryResolver* 覆盖。
// 生产仍须：真实 DB 事务、主意图 CAS、成功回调后关另一平台 pending。
func ExampleCheckoutResolver() {
	alipayPay := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay&sign=demo",
	}
	wechatPay := &staticWechatCheckout{
		jsapi: &payment.JSAPIResult{
			AppID: "wx_app", Timestamp: "1710000000", NonceStr: "n",
			Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
		},
	}
	resolver := newMemoryResolverWithProviders(alipayPay, wechatPay)

	ali, err := resolver.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		log.Fatal(err)
	}
	ali2, err := resolver.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		log.Fatal(err)
	}

	wx, err := resolver.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderWechat, "openid-a")
	if err != nil {
		log.Fatal(err)
	}
	wx2, err := resolver.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderWechat, "openid-a")
	if err != nil {
		log.Fatal(err)
	}
	_, conflict := resolver.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderWechat, "openid-b")

	fmt.Println(ali.OrderID == ali2.OrderID)
	fmt.Println(ali.OrderID != wx.OrderID)
	fmt.Println(wx.OrderID == wx2.OrderID)
	fmt.Println(conflict != nil)
	// Output:
	// true
	// true
	// true
	// true
}

// --- 以下为示例用骨架，勿直接用于生产 ---

type staticWechatOAuth struct{ openID string }

func (s *staticWechatOAuth) OAuthURL(redirectURL, state string) (string, error) {
	u, err := url.Parse("https://open.weixin.qq.com/connect/oauth2/authorize")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *staticWechatOAuth) ExchangeOAuthCode(context.Context, string) (string, error) {
	return s.openID, nil
}

type alipayCheckoutProvider interface {
	WAPPay(ctx context.Context, order *payment.Order) (*payment.WAPResult, error)
}

type wechatCheckoutProvider interface {
	JSAPIPay(ctx context.Context, order *payment.Order, openID string) (*payment.JSAPIResult, error)
}

type orderLifecycle interface {
	Query(ctx context.Context, orderID string) (*payment.QueryResult, error)
	Close(ctx context.Context, orderID string) error
}

type staticAlipayCheckout struct {
	url string
	err error
}

func (s *staticAlipayCheckout) WAPPay(context.Context, *payment.Order) (*payment.WAPResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &payment.WAPResult{URL: s.url}, nil
}

type staticWechatCheckout struct {
	jsapi *payment.JSAPIResult
	err   error
}

func (s *staticWechatCheckout) JSAPIPay(context.Context, *payment.Order, string) (*payment.JSAPIResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	cp := *s.jsapi
	return &cp, nil
}

type staticLifecycle struct {
	mu             sync.Mutex
	paidOrderIDs   map[string]struct{}
	closedOrderIDs map[string]struct{}
	queryStatus    map[string]payment.PaymentStatus
	queryErr       map[string]error
	closeErr       error
}

func (s *staticLifecycle) Query(_ context.Context, orderID string) (*payment.QueryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.queryErr != nil {
		if err := s.queryErr[orderID]; err != nil {
			return nil, err
		}
	}
	if s.queryStatus != nil {
		if st, ok := s.queryStatus[orderID]; ok {
			return &payment.QueryResult{OrderID: orderID, Status: st}, nil
		}
	}
	if s.paidOrderIDs != nil {
		if _, ok := s.paidOrderIDs[orderID]; ok {
			return &payment.QueryResult{OrderID: orderID, Status: payment.PaymentStatusSuccess}, nil
		}
	}
	return &payment.QueryResult{OrderID: orderID, Status: payment.PaymentStatusPending}, nil
}

func (s *staticLifecycle) Close(_ context.Context, orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closeErr != nil {
		return s.closeErr
	}
	if s.closedOrderIDs == nil {
		s.closedOrderIDs = make(map[string]struct{})
	}
	s.closedOrderIDs[orderID] = struct{}{}
	return nil
}

type checkoutKey struct {
	intent   string
	provider payment.Provider
}

type slotPhase int

const (
	phaseReady slotPhase = iota
	phaseCreating
	phaseRetiring
	phaseUncertain // provider 槽位墓碑；不封锁 intent 其它 provider
)

type storedCheckout struct {
	phase       slotPhase
	checkout    *onepay.Checkout
	orderID     string
	payerDigest string
	ready       chan struct{}
}

// intent 级仅 open/paid（单调：paid 不可降级）。uncertain 只落在 provider 槽位。
type intentState int

const (
	intentOpen intentState = iota
	intentPaid
)

type intentRecord struct {
	Amount      int64
	Subject     string
	NotifyURL   string
	state       intentState
	paidOrderID string
}

type retireOutcome int

const (
	retireAllowRebuild retireOutcome = iota
	retireAlreadyPaid
	retireUncertain
	retireCtxAbort // ctx 取消/超时：不落墓碑
)

// memoryResolver 进程内骨架：
//   - creating/retiring 占位 + ready 等待
//   - intent paid 单调 CAS；uncertain 仅锁 provider 槽
//   - 创建失败先 Query；ctx 取消不落永久墓碑
type memoryResolver struct {
	mu        sync.Mutex
	store     map[checkoutKey]*storedCheckout
	intents   map[string]*intentRecord
	seq       int
	alipay    alipayCheckoutProvider
	wechat    wechatCheckoutProvider
	lifecycle orderLifecycle
	now       func() time.Time
}

func demoIntents() map[string]*intentRecord {
	return map[string]*intentRecord{
		"intent-1": {
			Amount: 9900, Subject: "会员订阅",
			NotifyURL: "https://merchant.example/payment/notify",
		},
		"intent-1001": {
			Amount: 9900, Subject: "会员订阅",
			NotifyURL: "https://merchant.example/payment/notify",
		},
	}
}

func newMemoryResolver() *memoryResolver {
	return newMemoryResolverWithProviders(
		&staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"},
		&staticWechatCheckout{jsapi: &payment.JSAPIResult{
			AppID: "app", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
		}},
	)
}

func newMemoryResolverWithProviders(ali alipayCheckoutProvider, wx wechatCheckoutProvider) *memoryResolver {
	return &memoryResolver{
		store:     make(map[checkoutKey]*storedCheckout),
		intents:   demoIntents(),
		alipay:    ali,
		wechat:    wx,
		lifecycle: &staticLifecycle{},
		now:       time.Now,
	}
}

func (m *memoryResolver) ResolveOrCreate(ctx context.Context, intentID string, provider payment.Provider, payerOpenID string) (*onepay.Checkout, error) {
	if intentID == "" {
		return nil, fmt.Errorf("%w: empty intent", payment.ErrInvalidRequest)
	}
	key := checkoutKey{intent: intentID, provider: provider}
	digest := ""
	if provider == payment.ProviderWechat {
		if payerOpenID == "" {
			return nil, fmt.Errorf("%w: wechat openid required", payment.ErrInvalidRequest)
		}
		sum := sha256.Sum256([]byte(payerOpenID))
		digest = hex.EncodeToString(sum[:])
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		m.mu.Lock()
		intent, ok := m.intents[intentID]
		if !ok {
			m.mu.Unlock()
			return nil, fmt.Errorf("%w: unknown intent %q", payment.ErrInvalidRequest, intentID)
		}
		if intent.Amount <= 0 || intent.Subject == "" || intent.NotifyURL == "" {
			m.mu.Unlock()
			return nil, fmt.Errorf("%w: intent %q missing amount/subject/notify", payment.ErrInvalidRequest, intentID)
		}
		// intent 级只拦 paid；uncertain 不封锁其它 provider。
		if intent.state == intentPaid {
			orderID := intent.paidOrderID
			m.mu.Unlock()
			return nil, fmt.Errorf("%w: intent already paid (order %s), do not recreate", payment.ErrInvalidRequest, orderID)
		}

		existing := m.store[key]
		if existing != nil {
			switch existing.phase {
			case phaseCreating, phaseRetiring:
				if provider == payment.ProviderWechat && existing.payerDigest != "" && existing.payerDigest != digest {
					m.mu.Unlock()
					return nil, fmt.Errorf("%w: wechat payer conflict", payment.ErrInvalidRequest)
				}
				ch := existing.ready
				m.mu.Unlock()
				select {
				case <-ch:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				continue

			case phaseUncertain:
				// 槽位墓碑：再次进入时重新 Query，而不是永久卡死。
				orderID := existing.orderID
				ready := make(chan struct{})
				existing.phase = phaseRetiring
				existing.ready = ready
				old := cloneStored(existing)
				old.orderID = orderID
				m.mu.Unlock()
				outcome, retireErr := m.retireExpired(ctx, old)
				m.finishRetire(key, ready, intentID, old, outcome)
				switch outcome {
				case retireAlreadyPaid:
					return nil, alreadyPaidError(old.orderID)
				case retireUncertain:
					return nil, fmt.Errorf("%w: slot uncertain for order %s; re-query later", payment.ErrInvalidRequest, orderID)
				case retireCtxAbort:
					return nil, contextAbortError(ctx, retireErr)
				case retireAllowRebuild:
					continue
				default:
					return nil, fmt.Errorf("%w: unknown retire outcome %v", payment.ErrInvalidRequest, outcome)
				}

			case phaseReady:
				if existing.checkout != nil && !existing.checkout.ExpiresAt.Before(m.now()) {
					if provider == payment.ProviderWechat && existing.payerDigest != digest {
						m.mu.Unlock()
						return nil, fmt.Errorf("%w: wechat payer conflict", payment.ErrInvalidRequest)
					}
					out := cloneCheckout(existing.checkout)
					m.mu.Unlock()
					return out, nil
				}
				// 过期 → retiring 占位（不提前 delete）；新建 ready 防 double-close。
				ready := make(chan struct{})
				existing.phase = phaseRetiring
				existing.ready = ready
				if existing.orderID == "" && existing.checkout != nil {
					existing.orderID = existing.checkout.OrderID
				}
				old := cloneStored(existing)
				m.mu.Unlock()
				outcome, retireErr := m.retireExpired(ctx, old)
				m.finishRetire(key, ready, intentID, old, outcome)
				switch outcome {
				case retireAlreadyPaid:
					return nil, alreadyPaidError(old.orderID)
				case retireUncertain:
					return nil, fmt.Errorf("%w: expire retire uncertain for order %s; re-query later", payment.ErrInvalidRequest, old.orderID)
				case retireCtxAbort:
					return nil, contextAbortError(ctx, retireErr)
				case retireAllowRebuild:
					continue
				default:
					return nil, fmt.Errorf("%w: unknown retire outcome %v", payment.ErrInvalidRequest, outcome)
				}
			}
		}

		// 空槽：creating 占位后锁外调网关。
		ready := make(chan struct{})
		m.seq++
		orderID := fmt.Sprintf("%s-%s-%d", intentID, provider, m.seq)
		intentSnap := *intent
		m.store[key] = &storedCheckout{phase: phaseCreating, ready: ready, payerDigest: digest, orderID: orderID}
		m.mu.Unlock()

		checkout, err := m.createPlatformCheckout(ctx, provider, orderID, intentSnap, payerOpenID)
		if err != nil {
			return nil, m.handleCreateFailure(ctx, key, ready, orderID, intentID, err)
		}

		m.mu.Lock()
		cur := m.store[key]
		if cur == nil || cur.phase != phaseCreating || cur.ready != ready {
			// 所有权丢失：必须 close 本路 ready，唤醒等待者。
			closeReady(ready)
			m.mu.Unlock()
			m.closeOrphan(ctx, orderID)
			continue
		}
		if provider == payment.ProviderWechat && cur.payerDigest != digest {
			delete(m.store, key)
			closeReady(ready)
			m.mu.Unlock()
			m.closeOrphan(ctx, orderID)
			return nil, fmt.Errorf("%w: wechat payer conflict", payment.ErrInvalidRequest)
		}
		// 创建期间他轨已 paid：关本路单。
		if intent.state == intentPaid {
			delete(m.store, key)
			closeReady(ready)
			paidOrder := intent.paidOrderID
			m.mu.Unlock()
			m.closeOrphan(ctx, orderID)
			return nil, fmt.Errorf("%w: intent already paid (order %s), do not recreate", payment.ErrInvalidRequest, paidOrder)
		}
		cur.phase = phaseReady
		cur.checkout = checkout
		cur.orderID = orderID
		closeReady(ready)
		out := cloneCheckout(checkout)
		m.mu.Unlock()
		return out, nil
	}
}

func (m *memoryResolver) createPlatformCheckout(ctx context.Context, provider payment.Provider, orderID string, intent intentRecord, payerOpenID string) (*onepay.Checkout, error) {
	order := &payment.Order{
		OrderID: orderID, Amount: intent.Amount, Subject: intent.Subject, NotifyURL: intent.NotifyURL,
	}
	expires := m.now().Add(15 * time.Minute)
	checkout := &onepay.Checkout{Provider: provider, OrderID: orderID, ExpiresAt: expires}
	switch provider {
	case payment.ProviderAlipay:
		wap, err := m.alipay.WAPPay(ctx, order)
		if err != nil {
			return nil, err
		}
		checkout.WAP = wap
	case payment.ProviderWechat:
		jsapi, err := m.wechat.JSAPIPay(ctx, order, payerOpenID)
		if err != nil {
			return nil, err
		}
		checkout.JSAPI = jsapi
	default:
		return nil, fmt.Errorf("%w: unsupported provider %q", payment.ErrInvalidRequest, provider)
	}
	return checkout, nil
}

func (m *memoryResolver) handleCreateFailure(ctx context.Context, key checkoutKey, ready chan struct{}, orderID, intentID string, createErr error) error {
	// create 本身因 ctx 取消/超时失败：骨架刻意 clear（订单未必已到平台）。
	// 生产应对预分配 orderID 先 Query，勿照抄为「永远 clear」。
	if isContextError(createErr) || isContextError(ctx.Err()) {
		m.clearCreating(key, ready)
		if isContextError(createErr) {
			return createErr
		}
		return ctx.Err()
	}
	if isDeterministicCreateError(createErr) {
		m.clearCreating(key, ready)
		return createErr
	}

	// 以下：create 已非确定性，orderID 可能已在平台存在 —— Query/Close 的 ctx abort
	// 必须保留 orderID 墓碑，禁止 clearCreating 后换号盲建（与 retireCtxAbort 对齐）。
	if m.lifecycle == nil {
		m.markSlotUncertain(key, ready, orderID)
		return fmt.Errorf("%w: create uncertain without lifecycle (order %s): %v", payment.ErrInvalidRequest, orderID, createErr)
	}
	q, err := m.lifecycle.Query(ctx, orderID)
	if err != nil {
		m.markSlotUncertain(key, ready, orderID)
		if isContextError(err) || isContextError(ctx.Err()) {
			return contextAbortError(ctx, err)
		}
		return fmt.Errorf("%w: create uncertain, query failed (order %s): %v", payment.ErrInvalidRequest, orderID, createErr)
	}
	switch q.Status {
	case payment.PaymentStatusSuccess:
		m.markIntentPaid(key, ready, orderID, intentID)
		return fmt.Errorf("%w: intent already paid (order %s), do not recreate", payment.ErrInvalidRequest, orderID)
	case payment.PaymentStatusRefunded:
		m.markIntentPaid(key, ready, orderID, intentID)
		return fmt.Errorf("%w: intent already paid-then-refunded (order %s), do not recreate", payment.ErrInvalidRequest, orderID)
	case payment.PaymentStatusPending:
		if closeErr := m.lifecycle.Close(ctx, orderID); closeErr != nil {
			// 平台已确认 pending：Close 失败（含 ctx）一律保留 orderID 墓碑。
			m.markSlotUncertain(key, ready, orderID)
			if isContextError(closeErr) {
				return contextAbortError(ctx, closeErr)
			}
			return fmt.Errorf("%w: create pending but close failed (order %s): %v", payment.ErrInvalidRequest, orderID, createErr)
		}
		m.clearCreating(key, ready)
		return createErr
	case payment.PaymentStatusClosed, payment.PaymentStatusFailed:
		// 终态：无需 Close，清占位后允许重建。
		m.clearCreating(key, ready)
		return createErr
	default:
		m.markSlotUncertain(key, ready, orderID)
		return fmt.Errorf("%w: create uncertain unknown status %q (order %s)", payment.ErrInvalidRequest, q.Status, orderID)
	}
}

func isDeterministicCreateError(err error) bool {
	return errors.Is(err, payment.ErrInvalidRequest) || errors.Is(err, payment.ErrInvalidConfig)
}

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func alreadyPaidError(orderID string) error {
	return fmt.Errorf("%w: intent already paid (order %s), do not recreate", payment.ErrInvalidRequest, orderID)
}

// contextAbortError 保证 retireCtxAbort 永不返回 (nil, nil)。
// 优先下游 context 错误；其次父 ctx.Err()；皆空时回落 Canceled。
func contextAbortError(ctx context.Context, retireErr error) error {
	if isContextError(retireErr) {
		return retireErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if retireErr != nil {
		return retireErr
	}
	return fmt.Errorf("%w: retire aborted", context.Canceled)
}

// closeReady 关闭 ready 并容忍重复 close。
// 调用约定：必须在持有 m.mu 时调用（与 store 所有权校验同一临界区）。
func closeReady(ready chan struct{}) {
	if ready == nil {
		return
	}
	select {
	case <-ready:
		// already closed
	default:
		close(ready)
	}
}

func (m *memoryResolver) clearCreating(key checkoutKey, ready chan struct{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur := m.store[key]; cur != nil && cur.phase == phaseCreating && cur.ready == ready {
		delete(m.store, key)
		closeReady(ready)
		return
	}
	// 非 owner：仍唤醒本路等待者，但不改他人槽位。
	closeReady(ready)
}

// markSlotUncertain 仅在 owner 时改槽位；从不写 intent 级 uncertain，不覆盖他人 store。
func (m *memoryResolver) markSlotUncertain(key checkoutKey, ready chan struct{}, orderID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur := m.store[key]; cur != nil && cur.ready == ready {
		cur.phase = phaseUncertain
		cur.checkout = nil
		cur.orderID = orderID
		closeReady(ready)
		return
	}
	// 非 owner：no-op 槽位，只 close 本路 ready。
	closeReady(ready)
}

// markIntentPaid 单调写入 intentPaid（不可被 uncertain 降级）；owner 清槽。
func (m *memoryResolver) markIntentPaid(key checkoutKey, ready chan struct{}, orderID, intentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if intent := m.intents[intentID]; intent != nil {
		// 单调：已 paid 保持 paid，可更新 paidOrderID 仅当空。
		if intent.state != intentPaid {
			intent.state = intentPaid
			intent.paidOrderID = orderID
		} else if intent.paidOrderID == "" {
			intent.paidOrderID = orderID
		}
	}
	if cur := m.store[key]; cur != nil && cur.ready == ready {
		delete(m.store, key)
		closeReady(ready)
		return
	}
	closeReady(ready)
}

func (m *memoryResolver) finishRetire(key checkoutKey, ready chan struct{}, intentID string, old *storedCheckout, outcome retireOutcome) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur := m.store[key]
	owns := cur != nil && cur.phase == phaseRetiring && cur.ready == ready

	if !owns {
		// 所有权丢失仍唤醒本路等待者。
		closeReady(ready)
		return
	}

	switch outcome {
	case retireAlreadyPaid:
		if intent := m.intents[intentID]; intent != nil {
			if intent.state != intentPaid {
				intent.state = intentPaid
				if old != nil {
					intent.paidOrderID = old.orderID
				}
			}
		}
		delete(m.store, key)
		closeReady(ready)
	case retireUncertain:
		// 仅槽位 uncertain，不碰 intent（避免封锁另一 provider）。
		cur.phase = phaseUncertain
		cur.checkout = nil
		if old != nil {
			cur.orderID = old.orderID
		}
		closeReady(ready)
	case retireCtxAbort:
		// 不落永久「空槽」：保留 orderID 以便下次先 Query，禁止盲目换号重建。
		// - 仍有 checkout（过期 ready 路径）→ 恢复 phaseReady，下次继续 retire
		// - 仅有 orderID（uncertain 重入）→ 恢复 phaseUncertain 墓碑
		// - 无任何平台单线索 → 才删槽
		if cur.checkout != nil {
			cur.phase = phaseReady
		} else {
			orderID := cur.orderID
			if orderID == "" && old != nil {
				orderID = old.orderID
			}
			if orderID != "" {
				cur.phase = phaseUncertain
				cur.checkout = nil
				cur.orderID = orderID
			} else {
				delete(m.store, key)
			}
		}
		closeReady(ready)
	case retireAllowRebuild:
		delete(m.store, key)
		closeReady(ready)
	default:
		// 未知 outcome：唤醒 waiters，恢复 uncertain 以免永久挂起 / 空槽盲建。
		orderID := cur.orderID
		if orderID == "" && old != nil {
			orderID = old.orderID
		}
		if orderID != "" {
			cur.phase = phaseUncertain
			cur.checkout = nil
			cur.orderID = orderID
		} else {
			delete(m.store, key)
		}
		closeReady(ready)
	}
}

func (m *memoryResolver) closeOrphan(ctx context.Context, orderID string) {
	if m.lifecycle == nil || orderID == "" {
		return
	}
	// Close 失败须入重试队列。
	if err := m.lifecycle.Close(ctx, orderID); err != nil {
		_ = err
	}
}

// retireExpired 退役过期/uncertain 槽位。ctx 类错误时返回 (retireCtxAbort, err)，err 供调用方避免 (nil,nil)。
func (m *memoryResolver) retireExpired(ctx context.Context, old *storedCheckout) (retireOutcome, error) {
	if old == nil || old.orderID == "" {
		return retireAllowRebuild, nil
	}
	if m.lifecycle == nil {
		return retireUncertain, nil
	}
	q, err := m.lifecycle.Query(ctx, old.orderID)
	if err != nil {
		if isContextError(err) {
			return retireCtxAbort, err
		}
		if isContextError(ctx.Err()) {
			return retireCtxAbort, ctx.Err()
		}
		return retireUncertain, nil
	}
	switch q.Status {
	case payment.PaymentStatusSuccess, payment.PaymentStatusRefunded:
		return retireAlreadyPaid, nil
	case payment.PaymentStatusPending:
		if err := m.lifecycle.Close(ctx, old.orderID); err != nil {
			if isContextError(err) {
				return retireCtxAbort, err
			}
			if isContextError(ctx.Err()) {
				return retireCtxAbort, ctx.Err()
			}
			return retireUncertain, nil
		}
		return retireAllowRebuild, nil
	case payment.PaymentStatusFailed, payment.PaymentStatusClosed:
		// 终态：不必 Close，允许重建。
		return retireAllowRebuild, nil
	default:
		return retireUncertain, nil
	}
}

func cloneStored(s *storedCheckout) *storedCheckout {
	if s == nil {
		return nil
	}
	out := *s
	if s.checkout != nil {
		out.checkout = cloneCheckout(s.checkout)
	}
	return &out
}

func cloneCheckout(c *onepay.Checkout) *onepay.Checkout {
	if c == nil {
		return nil
	}
	out := *c
	if c.WAP != nil {
		wap := *c.WAP
		out.WAP = &wap
	}
	if c.JSAPI != nil {
		js := *c.JSAPI
		out.JSAPI = &js
	}
	return &out
}

// --- tests ---

func TestMemoryResolverRetirePaidPersists(t *testing.T) {
	life := &staticLifecycle{paidOrderIDs: map[string]struct{}{}}
	ali := &staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life
	fixed := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	first, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	life.mu.Lock()
	life.paidOrderIDs[first.OrderID] = struct{}{}
	life.mu.Unlock()
	m.mu.Lock()
	if slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]; slot != nil && slot.checkout != nil {
		slot.checkout.ExpiresAt = fixed.Add(-time.Second)
	}
	m.mu.Unlock()

	if _, err = m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected already-paid error")
	}
	if _, err = m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected already-paid on second attempt")
	}
}

func TestMemoryResolverCreateUncertainBlocksSameProvider(t *testing.T) {
	life := &staticLifecycle{
		queryErr: map[string]error{"intent-1-alipay-1": errors.New("query timeout")},
	}
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: errors.New("gateway timeout"),
	}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected create uncertain error")
	}
	ali.err = nil
	// 同 provider 槽位 uncertain：再次 Query 仍失败则继续拒绝盲目重建。
	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected blocked by slot uncertain re-query failure")
	}
}

func TestMemoryResolverDualRailUncertainDoesNotBlockOtherProvider(t *testing.T) {
	life := &staticLifecycle{
		queryErr: map[string]error{"intent-1-alipay-1": errors.New("query timeout")},
	}
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: errors.New("gateway timeout"),
	}
	wx := &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}}
	m := newMemoryResolverWithProviders(ali, wx)
	m.lifecycle = life

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected alipay uncertain")
	}
	// 微信轨不应被支付宝 uncertain 封锁。
	got, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderWechat, "openid-a")
	if err != nil {
		t.Fatalf("wechat blocked by alipay uncertain: %v", err)
	}
	if got == nil || got.JSAPI == nil {
		t.Fatal("expected wechat jsapi checkout")
	}
}

func TestMemoryResolverPaidNotDowngradedByUncertain(t *testing.T) {
	m := newMemoryResolver()
	ready := make(chan struct{})
	key := checkoutKey{"intent-1", payment.ProviderAlipay}
	m.mu.Lock()
	m.store[key] = &storedCheckout{phase: phaseCreating, ready: ready, orderID: "o-paid"}
	m.mu.Unlock()

	// 先 paid，再 uncertain：intent 必须保持 paid（单调，不可降级）。
	m.markIntentPaid(key, ready, "o-paid", "intent-1")
	// 另一路 ready 模拟非 owner uncertain，不得覆盖 paid。
	other := make(chan struct{})
	m.markSlotUncertain(key, other, "o-other")

	m.mu.Lock()
	state := m.intents["intent-1"].state
	paidOrder := m.intents["intent-1"].paidOrderID
	m.mu.Unlock()
	if state != intentPaid || paidOrder != "o-paid" {
		t.Fatalf("state=%v order=%q, want paid/o-paid", state, paidOrder)
	}
}

func TestMemoryResolverCtxCancelNoTombstone(t *testing.T) {
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: context.DeadlineExceeded,
	}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline", err)
	}
	m.mu.Lock()
	_, hasSlot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	state := m.intents["intent-1"].state
	m.mu.Unlock()
	if hasSlot {
		t.Fatal("ctx cancel should not leave uncertain tombstone")
	}
	if state != intentOpen {
		t.Fatalf("intent state = %v, want open", state)
	}
	// 恢复后可重建。
	ali.err = nil
	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryResolverRetireClosedAllowsRebuild(t *testing.T) {
	life := &staticLifecycle{queryStatus: map[string]payment.PaymentStatus{}}
	ali := &staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life
	fixed := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	first, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	life.mu.Lock()
	life.queryStatus[first.OrderID] = payment.PaymentStatusClosed
	life.mu.Unlock()
	m.mu.Lock()
	if slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]; slot != nil && slot.checkout != nil {
		slot.checkout.ExpiresAt = fixed.Add(-time.Second)
	}
	m.mu.Unlock()

	second, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	if second.OrderID == first.OrderID {
		t.Fatalf("expected new order after closed retire, got same %s", second.OrderID)
	}
}

func TestMemoryResolverConcurrentSingleFlight(t *testing.T) {
	ali := &staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})

	const n = 16
	var wg sync.WaitGroup
	results := make([]*onepay.Checkout, n)
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i], errs[i] = m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
		}()
	}
	wg.Wait()
	var orderID string
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: %v", i, errs[i])
		}
		if orderID == "" {
			orderID = results[i].OrderID
		} else if results[i].OrderID != orderID {
			t.Fatalf("double create: %s vs %s", orderID, results[i].OrderID)
		}
	}
}

func TestMemoryResolverCreateQueryCtxAbortKeepsOrderID(t *testing.T) {
	// 非确定性 create 后 Query 因 DeadlineExceeded 失败 → 必须 uncertain 保留 orderID，禁止换号盲建。
	life := &staticLifecycle{
		queryErr: map[string]error{"intent-1-alipay-1": context.DeadlineExceeded},
	}
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: errors.New("gateway timeout"),
	}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life

	got, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if got != nil || err == nil {
		t.Fatalf("got (%v, %v), want (nil, err)", got, err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}

	m.mu.Lock()
	slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	m.mu.Unlock()
	if slot == nil || slot.phase != phaseUncertain || slot.orderID != "intent-1-alipay-1" {
		t.Fatalf("slot = %+v, want uncertain/intent-1-alipay-1", slot)
	}

	// create 恢复后仍不得空槽换号。
	ali.err = nil
	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected still blocked, not blind recreate")
	}
	m.mu.Lock()
	slot2 := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	m.mu.Unlock()
	if slot2 == nil || slot2.orderID != "intent-1-alipay-1" {
		t.Fatal("orderID tombstone lost after retry")
	}
}

func TestMemoryResolverCreatePendingCloseCtxAbortKeepsOrderID(t *testing.T) {
	// create 不确定 → Query=pending → Close 因 ctx 失败 → 保留 orderID，禁止盲建。
	life := &staticLifecycle{
		queryStatus: map[string]payment.PaymentStatus{"intent-1-alipay-1": payment.PaymentStatusPending},
		closeErr:    context.DeadlineExceeded,
	}
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: errors.New("gateway timeout"),
	}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	m.mu.Lock()
	slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	m.mu.Unlock()
	if slot == nil || slot.phase != phaseUncertain || slot.orderID != "intent-1-alipay-1" {
		t.Fatalf("slot = %+v, want uncertain tombstone", slot)
	}
}

func TestMemoryResolverRetireCtxAbortReturnsNonNilError(t *testing.T) {
	// Query 返回 DeadlineExceeded，但父 ctx 仍有效 → 不得 (nil, nil)。
	life := &staticLifecycle{
		queryErr: map[string]error{},
	}
	ali := &staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life
	fixed := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	first, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	life.mu.Lock()
	life.queryErr[first.OrderID] = context.DeadlineExceeded
	life.mu.Unlock()
	m.mu.Lock()
	if slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]; slot != nil && slot.checkout != nil {
		slot.checkout.ExpiresAt = fixed.Add(-time.Second)
	}
	m.mu.Unlock()

	got, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if got != nil {
		t.Fatalf("checkout = %v, want nil", got)
	}
	if err == nil {
		t.Fatal("expected non-nil error, got (nil, nil)")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
}

func TestMemoryResolverRetireCtxAbortKeepsUncertainOrderID(t *testing.T) {
	// uncertain 槽重入时 Query 因 ctx 类错误中止 → 必须保留 orderID 墓碑，禁止空槽重建。
	life := &staticLifecycle{
		queryErr: map[string]error{"intent-1-alipay-1": errors.New("query timeout")},
	}
	ali := &staticAlipayCheckout{
		url: "https://openapi.alipay.com/gateway.do?x=1",
		err: errors.New("gateway timeout"),
	}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected first create uncertain")
	}

	// 重入：Query 改为 DeadlineExceeded（父 ctx 仍 Background）。
	life.mu.Lock()
	life.queryErr = map[string]error{"intent-1-alipay-1": context.DeadlineExceeded}
	life.mu.Unlock()

	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected ctx abort error")
	}

	m.mu.Lock()
	slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	m.mu.Unlock()
	if slot == nil {
		t.Fatal("slot deleted; orderID tombstone lost")
	}
	if slot.phase != phaseUncertain {
		t.Fatalf("phase = %v, want uncertain", slot.phase)
	}
	if slot.orderID != "intent-1-alipay-1" {
		t.Fatalf("orderID = %q, want intent-1-alipay-1", slot.orderID)
	}

	// 即使 create 恢复，也不应空槽换新单号（仍 uncertain，Query 仍 deadline）。
	ali.err = nil
	if _, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, ""); err == nil {
		t.Fatal("expected still blocked / abort, not blind recreate")
	}
	m.mu.Lock()
	slot2 := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]
	m.mu.Unlock()
	if slot2 == nil || slot2.orderID != "intent-1-alipay-1" {
		t.Fatal("blind recreate replaced orderID tombstone")
	}
}

func TestMemoryResolverRetireFailedAllowsRebuildWithoutClose(t *testing.T) {
	life := &staticLifecycle{queryStatus: map[string]payment.PaymentStatus{}}
	ali := &staticAlipayCheckout{url: "https://openapi.alipay.com/gateway.do?x=1"}
	m := newMemoryResolverWithProviders(ali, &staticWechatCheckout{jsapi: &payment.JSAPIResult{
		AppID: "a", Timestamp: "1", NonceStr: "n", Package: "prepay_id=p", SignType: "RSA", PaySign: "s",
	}})
	m.lifecycle = life
	fixed := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	first, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	life.mu.Lock()
	life.queryStatus[first.OrderID] = payment.PaymentStatusFailed
	life.mu.Unlock()
	m.mu.Lock()
	if slot := m.store[checkoutKey{"intent-1", payment.ProviderAlipay}]; slot != nil && slot.checkout != nil {
		slot.checkout.ExpiresAt = fixed.Add(-time.Second)
	}
	m.mu.Unlock()

	second, err := m.ResolveOrCreate(context.Background(), "intent-1", payment.ProviderAlipay, "")
	if err != nil {
		t.Fatal(err)
	}
	if second.OrderID == first.OrderID {
		t.Fatal("expected rebuild after failed")
	}
	life.mu.Lock()
	_, closed := life.closedOrderIDs[first.OrderID]
	life.mu.Unlock()
	if closed {
		t.Fatal("failed status should not Close")
	}
}
