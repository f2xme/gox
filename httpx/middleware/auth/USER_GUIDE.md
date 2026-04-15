# Auth User 使用指南

## 概述

`httpx/middleware/auth` 包现在提供了 `User` 接口和相关辅助函数，用于在 HTTP 上下文中管理用户认证信息。

## 核心接口

```go
type User interface {
    GetID() string
    GetRole() string
    GetRoles() []string
    HasRole(role string) bool
    HasPermission(perm string) bool
}
```

## 默认实现

```go
type DefaultUser struct {
    ID          string
    Role        string
    Roles       []string
    Permissions []string
    Extra       map[string]any  // 扩展字段
}
```

## 基本使用

### 1. 在中间件中设置用户

```go
func AuthMiddleware() httpx.Middleware {
    return func(next httpx.Handler) httpx.Handler {
        return func(ctx httpx.Context) error {
            // 验证 token 后创建用户
            user := &auth.DefaultUser{
                ID:          "user-123",
                Role:        "admin",
                Roles:       []string{"admin", "editor"},
                Permissions: []string{"user:read", "user:write", "user:delete"},
            }
            
            auth.SetUser(ctx, user)
            return next(ctx)
        }
    }
}
```

### 2. 在 Handler 中获取用户信息

```go
func MyHandler(ctx httpx.Context) error {
    // 获取用户对象
    user, ok := auth.GetUser(ctx)
    if !ok {
        return ctx.Fail("未认证")
    }
    
    // 访问用户信息
    uid := user.GetID()
    role := user.GetRole()
    
    // 检查角色
    if !user.HasRole("admin") {
        return ctx.Fail("权限不足")
    }
    
    // 检查权限
    if !user.HasPermission("user:delete") {
        return ctx.Fail("无删除权限")
    }
    
    return ctx.Success(map[string]any{
        "uid":  uid,
        "role": role,
    })
}
```

## 与 JWT 集成

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/f2xme/gox/httpx/middleware/auth"
)

// JWT Claims 实现 User 接口
type JWTClaims struct {
    UserID      string   `json:"uid"`
    Role        string   `json:"role"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}

func (c *JWTClaims) GetID() string              { return c.UserID }
func (c *JWTClaims) GetRole() string            { return c.Role }
func (c *JWTClaims) GetRoles() []string         { return c.Roles }
func (c *JWTClaims) HasRole(role string) bool {
    if c.Role == role {
        return true
    }
    return slices.Contains(c.Roles, role)
}
func (c *JWTClaims) HasPermission(perm string) bool {
    return slices.Contains(c.Permissions, perm)
}

// 同时实现 auth.Claims 接口
func (c *JWTClaims) GetSubject() string {
    return c.Subject
}
func (c *JWTClaims) Get(key string) (any, bool) {
    // 实现自定义字段获取
    return nil, false
}

// JWT Validator
type JWTValidator struct {
    secret []byte
}

func (v *JWTValidator) Validate(tokenString string) (auth.Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
        return v.secret, nil
    })
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*JWTClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    return claims, nil
}

// 使用
func main() {
    validator := &JWTValidator{secret: []byte("your-secret")}
    
    middleware := auth.New(
        auth.WithValidator(validator),
    )
    
    // 在 handler 中，Claims 自动也是 User
    handler := middleware(func(ctx httpx.Context) error {
        // 获取 Claims
        claims := auth.GetClaims(ctx)
        
        // Claims 也实现了 User 接口，可以直接使用
        if user, ok := claims.(auth.User); ok {
            auth.SetUser(ctx, user)
        }
        
        // 现在可以使用 User 对象
        user, _ := auth.GetUser(ctx)
        uid := user.GetID()
        hasAdmin := user.HasRole("admin")
        
        return ctx.Success(map[string]any{
            "uid":      uid,
            "hasAdmin": hasAdmin,
        })
    })
}
```

## 自定义 User 实现

```go
type CustomUser struct {
    Username    string
    Email       string
    IsActive    bool
    Roles       []string
    Permissions map[string]bool
}

func (u *CustomUser) GetID() string {
    return u.Email
}

func (u *CustomUser) GetRole() string {
    if len(u.Roles) > 0 {
        return u.Roles[0]
    }
    return ""
}

func (u *CustomUser) GetRoles() []string {
    return u.Roles
}

func (u *CustomUser) HasRole(role string) bool {
    return slices.Contains(u.Roles, role)
}

func (u *CustomUser) HasPermission(perm string) bool {
    return u.Permissions[perm]
}
```

## 辅助函数列表

| 函数 | 说明 |
|------|------|
| `SetUser(ctx, user)` | 设置用户到上下文 |
| `GetUser(ctx)` | 获取用户，返回 (User, bool) |

## 最佳实践

1. **使用 DefaultUser 快速开始**
   ```go
   user := &auth.DefaultUser{
       ID:   "123",
       Role: "admin",
   }
   ```

2. **缓存 User 对象避免重复查找**
   ```go
   // 好的做法：一次查找，多次使用
   user, ok := auth.GetUser(ctx)
   if !ok {
       return ctx.Fail("未认证")
   }
   
   uid := user.GetID()
   hasAdmin := user.HasRole("admin")
   canDelete := user.HasPermission("user:delete")
   
   // 避免：多次调用 GetUser
   // uid := auth.GetUserID(ctx)  // 第一次查找
   // hasAdmin := auth.HasRole(ctx, "admin")  // 第二次查找
   ```

3. **利用 Extra 字段存储额外信息**
   ```go
   user := &auth.DefaultUser{
       ID:   "123",
       Role: "admin",
       Extra: map[string]any{
           "email":    "user@example.com",
           "nickname": "Admin User",
       },
   }
   ```

4. **权限命名约定**
   ```go
   // 推荐使用 resource:action 格式
   Permissions: []string{
       "user:read",
       "user:write",
       "user:delete",
       "post:publish",
   }
   ```

5. **角色层级**
   ```go
   // 使用 Role 表示主角色，Roles 表示所有角色
   user := &auth.DefaultUser{
       Role:  "admin",           // 主角色
       Roles: []string{"admin", "editor", "viewer"}, // 所有角色
   }
   ```

6. **GetRoles 返回副本**
   ```go
   // GetRoles() 返回切片的副本，可以安全修改
   roles := user.GetRoles()
   roles[0] = "modified"  // 不会影响原始用户对象
   ```

## 与现有 Claims 的关系

- `User` 接口是独立的，不依赖 `Claims`
- `Claims` 可以选择实现 `User` 接口，实现统一
- 两者可以共存，通过不同的 context key 存储
- 推荐让 JWT Claims 同时实现两个接口，获得最佳灵活性
