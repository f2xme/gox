/*
Package httpx 提供统一的 HTTP 框架抽象层。

httpx 定义了 HTTP 服务器的标准接口,支持多种 HTTP 框架(Gin、Echo 等)。
通过这些接口,可以在不同框架之间切换而无需修改业务代码。

# 功能特性

  - 统一接口:Engine、Router、Context 三层抽象,屏蔽框架差异
  - Value 类型:Param/Query/Header 返回 Value,支持链式类型转换与默认值回退
  - 多值 Query:QueryAll 支持同名参数多值(如 ?tag=a&tag=b)
  - 自动验证:Bind 系列方法自动调用 Validator 接口,无需手动校验
  - 可选统一响应:httpx.Data/Done/Fail 输出标准 JSON 格式;ErrBadRequest 等语义化 StatusError 构造函数通过 ErrorHandler 统一处理
  - 中间件:洋葱模型,支持全局和路由组级别注册
  - 错误处理:可自定义 ErrorHandler,默认映射 StatusError 到状态码
  - 集成测试:httpx/testkit 可对 httpx.Engine 发起真实 HTTP 请求,覆盖路由、中间件、绑定、错误处理和响应

# 快速开始

基本使用:

	import (
		"github.com/f2xme/gox/httpx"
		"github.com/f2xme/gox/httpx/adapter/gin"
	)

	engine := gin.New()

	engine.GET("/users/:id", func(c httpx.Context) error {
		id, err := c.Param("id").Int64()
		if err != nil {
			return httpx.ErrBadRequest("invalid id")
		}
		return c.JSON(200, map[string]int64{"id": id})
	})

	engine.Start(":8080")

# Value 类型化取值

	id, err := c.Param("id").Int64()           // 必须是合法整数
	page     := c.Query("page").IntOr(1)        // 非法/缺失则回退默认值
	enabled  := c.Query("enabled").BoolOr(false)
	until, _ := c.Query("until").Time(time.RFC3339)
	tags     := c.Query("tags").Split(",")      // "a,b,c" -> ["a","b","c"]
	name     := c.Query("name").Or("guest")     // 缺失回退字符串默认值
	if c.Header("X-Admin").Exists() { ... }     // 存在性判断

# 请求绑定与自动验证

使用 validate tag 进行验证（默认使用 gox/validator，支持中文错误消息）:

	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required" label:"名字"`
		Age   int    `json:"age" validate:"required,min=1,max=150" label:"年龄"`
		Email string `json:"email" validate:"email" label:"邮箱"`
	}

	func createUser(c httpx.Context) error {
		var req CreateUserRequest
		if err := c.BindJSON(&req); err != nil {
			return httpx.ErrBadRequest(err.Error())
		}
		return c.JSON(200, req)
	}

label tag 用于自定义验证错误消息中的字段名(支持中文)。

支持的 validate 验证规则:
  - required: 必填
  - min/max: 最小/最大值(数字)或长度(字符串/数组)
  - len: 固定长度
  - email: 邮箱格式
  - url: URL 格式
  - oneof: 枚举值
  - alpha/alphanum/numeric: 字母/字母数字/纯数字
  - uuid/uuid3/uuid4/uuid5: UUID 格式
  - ip/ipv4/ipv6: IP 地址格式
  - contains/startswith/endswith: 包含/前缀/后缀
  - phone: 中国大陆手机号
  - id_card: 中国大陆身份证号
  - bank_card: 银行卡号
  - 更多规则见 github.com/go-playground/validator

自定义 validator（可选）:

	import ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	import "github.com/gin-gonic/gin/binding"

	// 使用自定义 validator
	engine := ginadapter.New(
		ginadapter.WithValidator(myCustomValidator),
	)

实现 Validator 接口进行额外的自定义验证:

	type CreateUserRequest struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"required,min=1,max=150"`
	}

	func (r *CreateUserRequest) Validate() error {
		// 这里可以添加 validate tag 无法表达的复杂验证逻辑
		if r.Name == "admin" {
			return fmt.Errorf("name cannot be admin")
		}
		return nil
	}

	func createUser(c httpx.Context) error {
		var req CreateUserRequest
		// BindJSON 会先执行 validate tag 验证，再调用 req.Validate()
		if err := c.BindJSON(&req); err != nil {
			return httpx.ErrBadRequest(err.Error())
		}
		// 此时 req 已通过所有验证
		return c.JSON(200, req)
	}

所有 Bind 方法(Bind/BindJSON/BindQuery/BindForm)都支持自动验证。

# 路由分组与中间件

	api := engine.Group("/api/v1")
	api.Use(authMiddleware)
	api.GET("/users", listUsers)
	api.POST("/users", createUser)

# 错误处理

	engine.SetErrorHandler(func(c httpx.Context, err error) {
		var he *httpx.StatusError
		if errors.As(err, &he) {
			c.JSON(he.Status, map[string]string{"message": he.Message})
			return
		}
		c.JSON(500, map[string]string{"message": "internal error"})
	})

# 优雅关闭

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	engine.Shutdown(ctx)

# 集成测试

使用 httpx/testkit 测试真实 HTTP 链路:

	import "github.com/f2xme/gox/httpx/testkit"

	client := testkit.New(t, engine)
	defer client.Close()

	client.POSTJSON("/users", CreateUserRequest{Name: "Alice"}).
		ExpectStatus(201).
		ExpectJSONValue("success", true)

	client.Do(http.MethodTrace, "/debug", nil).
		ExpectStatus(200)
*/
package httpx
