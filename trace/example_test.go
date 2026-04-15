package trace_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/f2xme/gox/trace"
)

// User 示例用户结构
type User struct {
	ID   int64
	Name string
}

// UserDAO 示例 DAO 层
type UserDAO struct{}

func (d *UserDAO) GetUserByID(ctx context.Context, id int64) (user *User, err error) {
	defer trace.DAO(ctx, "GetUserByID", "id", id)(&err)

	// 模拟数据库查询
	if id == 0 {
		return nil, errors.New("user not found")
	}

	return &User{ID: id, Name: "test user"}, nil
}

// UserService 示例 Service 层
type UserService struct {
	dao *UserDAO
}

func (s *UserService) GetUser(ctx context.Context, id int64) (user *User, err error) {
	defer trace.Service(ctx, "GetUser", "id", id)(&err)

	return s.dao.GetUserByID(ctx, id)
}

func Example_basic() {
	// 设置回调记录 Span 结果
	trace.SetCallback(func(r *trace.SpanResult) {
		fmt.Printf("kind=%s name=%s duration_ms=%d success=%v\n",
			r.Kind(), r.Name(), r.DurationMs(), r.Success())
	})

	ctx := context.Background()
	service := &UserService{dao: &UserDAO{}}

	// 调用服务
	user, err := service.GetUser(ctx, 123)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("user: id=%d name=%s\n", user.ID, user.Name)

	// Output:
	// kind=dao name=GetUserByID duration_ms=0 success=true
	// kind=service name=GetUser duration_ms=0 success=true
	// user: id=123 name=test user
}

func Example_withError() {
	trace.SetCallback(func(r *trace.SpanResult) {
		fmt.Printf("kind=%s name=%s success=%v error=%v\n",
			r.Kind(), r.Name(), r.Success(), r.Error)
	})

	ctx := context.Background()
	service := &UserService{dao: &UserDAO{}}

	// 调用失败的情况
	_, err := service.GetUser(ctx, 0)
	if err != nil {
		fmt.Println("expected error:", err)
	}

	// Output:
	// kind=dao name=GetUserByID success=false error=user not found
	// kind=service name=GetUser success=false error=user not found
	// expected error: user not found
}

func Example_contextOperations() {
	// 创建追踪信息
	info := &trace.Info{
		TraceID:   "trace-123",
		SpanID:    "span-456",
		RequestID: "req-789",
	}

	// 注入到 context
	ctx := trace.ToContext(context.Background(), info)

	// 从 context 提取
	extracted := trace.FromContext(ctx)
	fmt.Printf("trace_id=%s span_id=%s request_id=%s\n",
		extracted.TraceID, extracted.SpanID, extracted.RequestID)

	// Output:
	// trace_id=trace-123 span_id=span-456 request_id=req-789
}

func Example_manualSpan() {
	ctx := context.Background()

	// 手动创建和控制 Span
	span := trace.StartSpan(ctx, trace.SpanKindService, "BatchProcess")
	span.Set("count", 3)

	// 模拟批处理
	for range 3 {
		// 处理逻辑
	}

	result := span.End(nil)

	fmt.Printf("name=%s count=%v duration_ms=%d\n",
		result.Name(), result.Attrs()["count"], result.DurationMs())

	// Output:
	// name=BatchProcess count=3 duration_ms=0
}
