/*
Package httpx 提供统一的 HTTP 框架抽象层。

httpx 定义了 HTTP 服务器的标准接口,支持多种 HTTP 框架(Gin、Echo 等)。
通过这些接口,可以在不同框架之间切换而无需修改业务代码。

# 功能特性

  - 统一接口:Engine、Router、Context 三层抽象,屏蔽框架差异
  - Value 类型:Param/Query/Header 返回 Value,支持链式类型转换与默认值回退
  - 多值 Query:QueryAll 支持同名参数多值(如 ?tag=a&tag=b)
  - 自动验证:Bind 系列方法自动调用 Validator 接口,无需手动校验
  - 统一响应:httpx.Data/Done/Fail 输出标准 JSON 格式;ErrBadRequest 等语义化 HTTPError 构造函数通过 ErrorHandler 统一处理
  - 中间件:洋葱模型,支持全局和路由组级别注册
  - 错误处理:可自定义 ErrorHandler,默认映射 HTTPError 到状态码

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

实现 Validator 接口,Bind 系列方法会自动调用验证:

	type CreateUserRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	func (r *CreateUserRequest) Validate() error {
		if r.Name == "" {
			return fmt.Errorf("name is required")
		}
		if r.Age < 1 || r.Age > 150 {
			return fmt.Errorf("age must be between 1 and 150")
		}
		return nil
	}

	func createUser(c httpx.Context) error {
		var req CreateUserRequest
		// BindJSON 会自动调用 req.Validate()
		if err := c.BindJSON(&req); err != nil {
			return httpx.ErrBadRequest(err.Error())
		}
		// 此时 req 已通过验证
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
		var he *httpx.HTTPError
		if errors.As(err, &he) {
			c.JSON(he.Code, httpx.NewFailResponse(he.Message))
			return
		}
		c.JSON(500, httpx.NewFailResponse("internal error"))
	})

# 优雅关闭

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	engine.Shutdown(ctx)
*/
package httpx
