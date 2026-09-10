package captcha

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

var (
	// ErrInvalidSlideOptions 表示滑动验证配置无效。
	ErrInvalidSlideOptions = errors.New("captcha: invalid slide options")
)

// TrackPoint 表示滑动轨迹点；X 使用挑战的距离单位，T 为自拖动开始的毫秒数。
type TrackPoint struct {
	X float64 `json:"x"`
	T int     `json:"t"`
}

// SlideChallenge 表示一次滑动挑战；前端将整个可滑动距离映射为 Distance。
type SlideChallenge struct {
	Token    string  `json:"token"`
	Distance float64 `json:"distance"`
}

// SlideVerifyData 表示客户端提交的滑动数据；Track 最多包含 512 个点。
type SlideVerifyData struct {
	Token    string       `json:"token"`
	Distance float64      `json:"distance"`
	Duration int          `json:"duration"` // 毫秒，必须等于轨迹最后一点的 T
	Track    []TrackPoint `json:"track"`
}

// SlideVerifier 验证一次性 token 和滑动行为，不持有独立的会话缓存或清理协程。
// ponytail: 轨迹校验仅是启发式，客户端可伪造；需要防机器人时接入服务端风控。
type SlideVerifier struct {
	store Store
	taker Taker
	opts  SlideOptions
}

// NewSlide 创建滑动验证器；store 必须同时实现 Store 和原子 Taker。
// 存储由调用方管理生命周期，应为滑动验证使用独立实例或 key 前缀。
func NewSlide(store Store, opts ...SlideOption) (*SlideVerifier, error) {
	if store == nil {
		return nil, ErrNilStore
	}
	taker, ok := store.(Taker)
	if !ok {
		return nil, ErrAtomicStoreRequired
	}
	o := SlideOptions{
		TTL: 5 * time.Minute, Distance: 280, Tolerance: 5,
		MinDuration: 300 * time.Millisecond, MaxDuration: 10 * time.Second,
		MinSpeedVariance: 0.001,
	}
	for _, opt := range opts {
		opt(&o)
	}
	if o.TTL <= 0 || !finite(o.Distance) || o.Distance <= 0 ||
		!finite(o.Tolerance) || o.Tolerance < 0 || o.Tolerance >= o.Distance ||
		o.MinDuration < time.Millisecond || o.MaxDuration < o.MinDuration ||
		!finite(o.MinSpeedVariance) || o.MinSpeedVariance < 0 {
		return nil, ErrInvalidSlideOptions
	}
	return &SlideVerifier{store: store, taker: taker, opts: o}, nil
}

// Generate 生成随机 token，存储创建时间并设置过期时间。
func (s *SlideVerifier) Generate(ctx context.Context) (SlideChallenge, error) {
	if err := ctx.Err(); err != nil {
		return SlideChallenge{}, err
	}
	token, err := generateID(32)
	if err != nil {
		return SlideChallenge{}, err
	}
	created := strconv.FormatInt(time.Now().UnixMilli(), 10)
	if err := s.store.Set(ctx, token, created, s.opts.TTL); err != nil {
		return SlideChallenge{}, err
	}
	return SlideChallenge{Token: token, Distance: s.opts.Distance}, nil
}

// Verify 验证滑动数据；token 无论验证成功或失败都只可尝试一次。
// token 不存在、已过期或行为不符合规则时返回 false，存储故障等返回错误。
func (s *SlideVerifier) Verify(ctx context.Context, data SlideVerifyData) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if data.Token == "" {
		return false, nil
	}
	stored, err := s.taker.Take(ctx, data.Token)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	created, err := strconv.ParseInt(stored, 10, 64)
	if err != nil {
		return false, fmt.Errorf("captcha: decode slide session: %w", err)
	}
	// 客户端声明的时长不能超过服务端会话已存在的时长。
	if created < 0 || created > time.Now().UnixMilli() ||
		int64(data.Duration) > time.Now().UnixMilli()-created {
		return false, nil
	}
	return s.validateBehavior(data), nil
}

// Delete 撤销尚未验证的 token，用于关闭或刷新验证。
func (s *SlideVerifier) Delete(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if token == "" {
		return ErrInvalidID
	}
	return s.store.Delete(ctx, token)
}

func (s *SlideVerifier) validateBehavior(data SlideVerifyData) bool {
	if int64(data.Duration) < s.opts.MinDuration.Milliseconds() ||
		int64(data.Duration) > s.opts.MaxDuration.Milliseconds() ||
		!finite(data.Distance) || math.Abs(data.Distance-s.opts.Distance) > s.opts.Tolerance ||
		len(data.Track) < 3 || len(data.Track) > 512 {
		return false
	}
	first, last := data.Track[0], data.Track[len(data.Track)-1]
	if first.X != 0 || first.T != 0 || last.T != data.Duration || !finite(last.X) ||
		math.Abs(last.X-data.Distance) > s.opts.Tolerance {
		return false
	}
	var mean, variance float64
	for i := 1; i < len(data.Track); i++ {
		prev, point := data.Track[i-1], data.Track[i]
		if !finite(point.X) || point.X < 0 || point.X > s.opts.Distance+s.opts.Tolerance ||
			point.T <= prev.T || point.T > data.Duration {
			return false
		}
		speed := (point.X - prev.X) / float64(point.T-prev.T)
		delta := speed - mean
		mean += delta / float64(i)
		variance += delta * (speed - mean)
	}
	variance /= float64(len(data.Track) - 1)
	return finite(variance) && variance >= s.opts.MinSpeedVariance
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
