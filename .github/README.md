# GitHub 配置文件

## Release 管理

### 自动化脚本

使用 `create-release.sh` 脚本创建标准化的 Release：

```bash
# 基本用法
.github/create-release.sh v0.2.0

# 带自定义说明
.github/create-release.sh v0.2.0 "Bug 修复版本"
```

脚本会自动：
1. 验证版本号格式
2. 检查未提交的更改
3. 生成变更日志
4. 使用模板创建 Release 说明
5. 创建并推送标签
6. 创建 GitHub Release

### 手动创建 Release

如果需要手动创建，请参考 `RELEASE_TEMPLATE.md` 模板。

### Release 配置

- `release.yml` - GitHub Release 自动生成变更日志的配置
- `RELEASE_TEMPLATE.md` - Release 说明模板
- `create-release.sh` - Release 创建脚本

## 标签规范

使用语义化版本号：`vMAJOR.MINOR.PATCH`

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向后兼容的功能新增
- **PATCH**: 向后兼容的问题修复

## PR 标签

为了更好地生成变更日志，请为 PR 添加以下标签：

- `feature` / `enhancement` - 新功能
- `bug` / `fix` - Bug 修复
- `optimization` / `refactor` / `performance` - 优化改进
- `documentation` / `docs` - 文档更新
- `breaking-change` - 破坏性变更
- `security` - 安全更新
