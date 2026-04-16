# Database 包改进日志

## 2026-04-16 - 接口重大改进

### 新增功能

#### 1. 完整的 CRUD 操作
- `Create(ctx, value)` - 创建记录
- `First(ctx, dest, conds...)` - 查询第一条记录
- `Find(ctx, dest, conds...)` - 查询所有记录
- `Save(ctx, value)` - 保存（插入或更新）
- `Update(ctx, model, column, value)` - 更新单个字段
- `Updates(ctx, model, values)` - 更新多个字段
- `Delete(ctx, value, conds...)` - 删除记录

#### 2. 链式查询构建
- `Where(query, args...)` - 添加查询条件
- `Order(value)` - 指定排序
- `Limit(limit)` - 限制返回记录数
- `Offset(offset)` - 设置偏移量
- `Select(query, args...)` - 指定查询字段
- `Omit(columns...)` - 忽略字段
- `Joins(query, args...)` - JOIN 查询
- `Group(name)` - GROUP BY
- `Having(query, args...)` - HAVING 条件
- `Preload(query, args...)` - 预加载关联
- `Model(value)` - 指定模型
- `Count(ctx, count)` - 统计记录数

#### 3. 手动事务控制
- `Begin(ctx)` - 开始事务
- `Commit()` - 提交事务
- `Rollback()` - 回滚事务

#### 4. 原生 SQL 支持
- `Exec(ctx, sql, values...)` - 执行原生 SQL
- `Raw(sql, values...)` - 创建原生查询
- `Scan(ctx, dest)` - 扫描查询结果

#### 5. 工具方法
- `WithContext(ctx)` - 设置 context

### 破坏性变更

- `Engine()` 方法重命名为 `Unwrap()`，语义更清晰
- 所有 CRUD 方法现在需要 `context.Context` 参数

### API 统一

- `mysqldb` 新增 `NewWithConfig()` 和 `MustNewWithConfig()` 方法
- `sqlitedb` 新增 `NewWithConfig()` 和 `MustNewWithConfig()` 方法
- 所有 adapter 现在提供一致的 API

### 文档改进

- 更新 `doc.go`，添加完整的使用示例
- 新增 `example_test.go`，包含 5 个实际示例
- 所有公开方法都有清晰的文档注释

### 优势

1. **无需类型断言** - 用户可以直接使用接口方法进行 CRUD 操作，不再需要 `db.Unwrap().(*gorm.DB)`
2. **类型安全** - 编译期检查，减少运行时错误
3. **更好的可测试性** - 接口方法更容易 mock
4. **链式调用** - 提供流畅的查询构建体验
5. **向后兼容** - 保留 `Unwrap()` 方法支持高级用法

### 迁移指南

#### 旧代码
```go
gormDB := db.Engine().(*gorm.DB)
gormDB.Create(&user)
gormDB.Where("age > ?", 18).Find(&users)
```

#### 新代码
```go
db.Create(ctx, &user)
db.Where("age > ?", 18).Find(ctx, &users)
```

如果需要 GORM 高级功能：
```go
gormDB := db.Unwrap().(*gorm.DB)
```

### 测试覆盖

- ✅ 所有现有测试通过
- ✅ 新增 5 个示例测试
- ✅ SQLite adapter 完整测试覆盖
