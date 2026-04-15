/*
Package validator 提供数据验证功能，封装 go-playground/validator/v10。

# 概述

validator 包提供结构体标签验证、自定义验证规则和中文错误消息支持。
基于 go-playground/validator/v10，提供更简洁的 API 和开箱即用的中文支持。

# 核心功能

  - 结构体标签验证
  - 自定义验证规则
  - 中文错误消息
  - 字段别名支持
  - 并发安全

# 使用示例

## 基本验证

	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"min=18,max=100"`
	}

	user := User{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	if err := validator.Validate(user); err != nil {
		log.Fatal(err)
	}

## 自定义字段名

	type User struct {
		Name  string `validate:"required" label:"姓名"`
		Email string `validate:"required,email" label:"邮箱"`
		Age   int    `validate:"min=18,max=100" label:"年龄"`
	}

	// 错误消息会显示：姓名为必填字段

## 创建自定义验证器

	v := validator.New()

	// 注册自定义验证规则
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		// 用户名只能包含字母、数字和下划线
		return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)
	})

	type User struct {
		Username string `validate:"required,username"`
	}

## 验证单个字段

	email := "invalid-email"
	err := validator.Var(email, "required,email")
	if err != nil {
		fmt.Println("邮箱格式不正确")
	}

## 条件验证

	type User struct {
		Role     string `validate:"required,oneof=admin user guest"`
		Password string `validate:"required_if=Role admin,min=8"`
	}

	// 如果 Role 是 admin，则 Password 必填且至少 8 位

# 常用验证标签

## 必填验证

	required              // 必填
	required_if=Field Val // 如果 Field 等于 Val，则必填
	required_unless       // 除非...否则必填
	required_with         // 如果其他字段存在，则必填
	required_without      // 如果其他字段不存在，则必填

## 字符串验证

	min=10                // 最小长度
	max=100               // 最大长度
	len=20                // 固定长度
	email                 // 邮箱格式
	url                   // URL 格式
	alpha                 // 只包含字母
	alphanum              // 只包含字母和数字
	numeric               // 只包含数字

## 数字验证

	min=18                // 最小值
	max=100               // 最大值
	gt=0                  // 大于
	gte=0                 // 大于等于
	lt=100                // 小于
	lte=100               // 小于等于
	eq=10                 // 等于
	ne=0                  // 不等于
	oneof=1 2 3           // 枚举值

## 日期验证

	datetime=2006-01-02   // 日期格式

## 比较验证

	eqfield=Password      // 等于另一个字段
	nefield=OldPassword   // 不等于另一个字段
	gtfield=StartDate     // 大于另一个字段
	ltfield=EndDate       // 小于另一个字段

## 切片/数组验证

	min=1                 // 最小元素数
	max=10                // 最大元素数
	unique                // 元素唯一
	dive                  // 验证切片中的每个元素

# 复杂验证示例

## 嵌套结构体

	type Address struct {
		City    string `validate:"required" label:"城市"`
		Street  string `validate:"required" label:"街道"`
		ZipCode string `validate:"required,len=6" label:"邮编"`
	}

	type User struct {
		Name    string  `validate:"required" label:"姓名"`
		Address Address `validate:"required,dive"`
	}

## 切片验证

	type User struct {
		Tags []string `validate:"required,min=1,max=5,dive,min=2,max=20"`
	}

	// Tags 必填，至少 1 个，最多 5 个
	// 每个 tag 长度在 2-20 之间

## 自定义错误消息

	v := validator.New()

	// 注册自定义翻译
	v.RegisterTranslation("username", func(ut ut.Translator) error {
		return ut.Add("username", "{0}只能包含字母、数字和下划线", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("username", fe.Field())
		return t
	})

# 最佳实践

## 1. 使用默认验证器

	// 推荐：使用包级函数
	err := validator.Validate(user)

	// 不推荐：每次创建新验证器
	v := validator.New()
	err := v.Validate(user)

## 2. 使用 label 标签自定义字段名

	type User struct {
		Name string `validate:"required" label:"用户名"`
	}

	// 错误消息：用户名为必填字段

## 3. 组合验证规则

	type User struct {
		Email string `validate:"required,email,max=100"`
	}

## 4. 在 HTTP 处理器中使用

	func CreateUser(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := validator.Validate(user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 创建用户
	}

## 5. 验证请求参数

	type CreateUserRequest struct {
		Username string `json:"username" validate:"required,min=3,max=20"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Age      int    `json:"age" validate:"required,min=18,max=100"`
	}

# 错误处理

	err := validator.Validate(user)
	if err != nil {
		// 类型断言获取详细错误
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				fmt.Printf("字段: %s, 错误: %s\n", e.Field(), e.Translate(trans))
			}
		}
	}

# 性能考虑

  - 验证器实例是线程安全的，应该复用
  - 使用默认验证器避免重复初始化
  - 验证规则会被缓存，重复验证性能很好

# 线程安全

所有验证器实例都是线程安全的，可以在多个 goroutine 中并发使用。
*/
package validator
