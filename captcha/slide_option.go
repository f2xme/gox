package captcha

import "time"

// SlideOptions 定义滑动验证规则；轨迹坐标与 Distance、Tolerance 使用相同单位。
type SlideOptions struct {
	TTL              time.Duration // 默认 5 分钟
	Distance         float64       // 终点距离，默认 280
	Tolerance        float64       // 距离允许误差，默认 5
	MinDuration      time.Duration // 最短滑动时间，默认 300ms
	MaxDuration      time.Duration // 最长滑动时间，默认 10s
	MinSpeedVariance float64       // 最小速度方差，默认 0.001；0 关闭此启发式
}

// SlideOption 定义滑动验证配置函数。
type SlideOption func(*SlideOptions)

// WithSlideTTL 设置 token 有效期。
func WithSlideTTL(ttl time.Duration) SlideOption {
	return func(o *SlideOptions) { o.TTL = ttl }
}

// WithSlideDistance 设置终点距离和允许误差，均使用挑战的距离单位。
func WithSlideDistance(distance, tolerance float64) SlideOption {
	return func(o *SlideOptions) { o.Distance, o.Tolerance = distance, tolerance }
}

// WithSlideDuration 设置允许的滑动时长范围，精度为毫秒。
func WithSlideDuration(min, max time.Duration) SlideOption {
	return func(o *SlideOptions) { o.MinDuration, o.MaxDuration = min, max }
}

// WithSlideMinSpeedVariance 设置速度方差下限，0 关闭匀速过滤。
// 速度变化不能证明是真人操作；键盘等输入方式可能产生匀速轨迹。
func WithSlideMinSpeedVariance(min float64) SlideOption {
	return func(o *SlideOptions) { o.MinSpeedVariance = min }
}
