# Real-time User Ban Status Check

## Overview

The auth middleware supports real-time user status checking through the `UserStatusChecker` interface. This ensures that when a user is banned, the ban takes effect immediately on their next request, without waiting for their token to expire.

## Why Real-time Checking?

**Problem with token-based ban status:**
- User logs in and gets a valid token
- Risk control system detects suspicious activity and bans the user
- User's token is still valid and contains old ban status (not banned)
- User can continue accessing the API until token expires

**Solution with real-time checking:**
- On every request, after token validation, query the current user status
- If user is banned in the database, deny access immediately
- Ban takes effect on the next request, regardless of token validity

## Implementation

### 1. Implement UserStatusChecker

```go
package main

import (
    "database/sql"
    "github.com/f2xme/gox/httpx/middleware/auth"
)

type DBUserStatusChecker struct {
    db *sql.DB
}

func (c *DBUserStatusChecker) IsBanned(userID string) (bool, error) {
    var banned bool
    err := c.db.QueryRow(
        "SELECT banned FROM users WHERE id = ?", 
        userID,
    ).Scan(&banned)
    
    if err == sql.ErrNoRows {
        // User not found - treat as not banned
        return false, nil
    }
    if err != nil {
        // Database error - fail open (allow access)
        // You can change this to fail closed if preferred
        return false, err
    }
    
    return banned, nil
}
```

### 2. Configure Middleware

```go
package main

import (
    "github.com/f2xme/gox/httpx"
    "github.com/f2xme/gox/httpx/middleware/auth"
)

func main() {
    app := httpx.New()
    
    // Create status checker
    statusChecker := &DBUserStatusChecker{db: db}
    
    // Configure auth middleware with real-time status checking
    app.Use(auth.New(
        auth.WithValidator(jwtValidator),
        auth.WithUserStatusChecker(statusChecker),
        auth.WithBanHandler(func(c httpx.Context) {
            c.Status(403)
            _ = c.JSON(403, map[string]any{
                "error": "Your account has been suspended",
                "code":  "ACCOUNT_BANNED",
            })
        }),
    ))
    
    app.Run(":8080")
}
```

## Performance Considerations

### Caching

For high-traffic applications, consider caching the ban status:

```go
type CachedUserStatusChecker struct {
    db    *sql.DB
    cache *redis.Client
    ttl   time.Duration
}

func (c *CachedUserStatusChecker) IsBanned(userID string) (bool, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("user:banned:%s", userID)
    cached, err := c.cache.Get(context.Background(), cacheKey).Result()
    if err == nil {
        return cached == "1", nil
    }
    
    // Cache miss - query database
    var banned bool
    err = c.db.QueryRow(
        "SELECT banned FROM users WHERE id = ?", 
        userID,
    ).Scan(&banned)
    
    if err != nil && err != sql.ErrNoRows {
        return false, err
    }
    
    // Cache the result
    value := "0"
    if banned {
        value = "1"
    }
    _ = c.cache.Set(context.Background(), cacheKey, value, c.ttl).Err()
    
    return banned, nil
}
```

### Cache Invalidation

When banning a user, invalidate their cache:

```go
func BanUser(userID string) error {
    // Update database
    _, err := db.Exec("UPDATE users SET banned = true WHERE id = ?", userID)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:banned:%s", userID)
    _ = cache.Del(context.Background(), cacheKey).Err()
    
    return nil
}
```

## Error Handling

The middleware uses a **fail-open** strategy by default:
- If `IsBanned()` returns an error, the request is allowed
- This prevents database outages from blocking all users

To use **fail-closed** (deny access on error):

```go
type FailClosedChecker struct {
    inner auth.UserStatusChecker
}

func (c *FailClosedChecker) IsBanned(userID string) (bool, error) {
    banned, err := c.inner.IsBanned(userID)
    if err != nil {
        // Treat errors as banned
        return true, err
    }
    return banned, nil
}
```

## Testing

```go
func TestBanTakesEffectImmediately(t *testing.T) {
    // User logs in
    token := loginUser("user123")
    
    // User makes a request - should succeed
    resp := makeRequest(token)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Ban the user
    banUser("user123")
    
    // User makes another request with same token - should fail
    resp = makeRequest(token)
    assert.Equal(t, 403, resp.StatusCode)
}
```

## Migration from Token-based Ban Check

If you previously used the `User.IsBanned()` method from token claims:

**Before:**
```go
// Ban status was in the token
type MyUser struct {
    ID     string
    Banned bool
}

func (u *MyUser) IsBanned() bool {
    return u.Banned
}
```

**After:**
```go
// Implement UserStatusChecker for real-time checking
type MyStatusChecker struct {
    db *sql.DB
}

func (c *MyStatusChecker) IsBanned(userID string) (bool, error) {
    var banned bool
    err := c.db.QueryRow("SELECT banned FROM users WHERE id = ?", userID).Scan(&banned)
    return banned, err
}

// Configure middleware
auth.New(
    auth.WithValidator(validator),
    auth.WithUserStatusChecker(&MyStatusChecker{db: db}),
)
```

The `User.IsBanned()` method is still available for backward compatibility, but it's no longer used by the middleware for ban checking.
