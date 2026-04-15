package auth

import (
	"slices"

	"github.com/f2xme/gox/httpx"
)

const UserContextKey = "auth.user"

// User 表示具有身份和权限的已认证用户
type User interface {
	GetID() string
	GetRole() string
	GetRoles() []string
	HasRole(role string) bool
	HasPermission(perm string) bool
	IsBanned() bool
}

// DefaultUser 是 User 接口的默认实现
type DefaultUser struct {
	ID          string
	Role        string
	Roles       []string
	Permissions []string
	Banned      bool
	Extra       map[string]any
}

// GetID 返回用户 ID
func (u *DefaultUser) GetID() string {
	return u.ID
}

// GetRole 返回主要角色
func (u *DefaultUser) GetRole() string {
	return u.Role
}

// GetRoles 返回所有角色的副本
func (u *DefaultUser) GetRoles() []string {
	if u.Roles == nil {
		return nil
	}
	return slices.Clone(u.Roles)
}

// HasRole 检查用户是否具有指定角色
func (u *DefaultUser) HasRole(role string) bool {
	if u.Role == role {
		return true
	}
	return slices.Contains(u.Roles, role)
}

// HasPermission 检查用户是否具有指定权限
func (u *DefaultUser) HasPermission(perm string) bool {
	return slices.Contains(u.Permissions, perm)
}

// IsBanned 检查用户是否被封禁
func (u *DefaultUser) IsBanned() bool {
	return u.Banned
}

// SetUser 将用户存储到上下文中
func SetUser(ctx httpx.Context, user User) {
	ctx.Set(UserContextKey, user)
}

// GetUser 从上下文中获取用户
func GetUser(ctx httpx.Context) (User, bool) {
	v, ok := ctx.Get(UserContextKey)
	if !ok {
		return nil, false
	}
	user, ok := v.(User)
	return user, ok
}
