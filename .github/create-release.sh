#!/bin/bash
# Release 创建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [ $# -lt 1 ]; then
    echo -e "${RED}错误: 请提供版本号${NC}"
    echo "用法: $0 <version> [release-notes]"
    echo "示例: $0 v0.2.0"
    echo "示例: $0 v0.2.0 'Bug 修复版本'"
    exit 1
fi

VERSION=$1
NOTES=${2:-""}

# 验证版本号格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}错误: 版本号格式不正确，应为 vX.Y.Z${NC}"
    exit 1
fi

# 检查是否有未提交的更改
if [[ -n $(git status -s) ]]; then
    echo -e "${YELLOW}警告: 存在未提交的更改${NC}"
    git status -s
    read -p "是否继续? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 检查标签是否已存在
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}错误: 标签 $VERSION 已存在${NC}"
    exit 1
fi

# 获取上一个版本号
PREVIOUS_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

echo -e "${GREEN}准备创建 Release: $VERSION${NC}"
if [ -n "$PREVIOUS_VERSION" ]; then
    echo "上一个版本: $PREVIOUS_VERSION"
fi

# 生成变更日志
echo -e "\n${YELLOW}生成变更日志...${NC}"
if [ -n "$PREVIOUS_VERSION" ]; then
    CHANGELOG=$(git log --pretty=format:"- %s (%h)" "$PREVIOUS_VERSION"..HEAD)
else
    CHANGELOG=$(git log --pretty=format:"- %s (%h)")
fi

# 读取模板
TEMPLATE_FILE=".github/RELEASE_TEMPLATE.md"
if [ -f "$TEMPLATE_FILE" ]; then
    RELEASE_NOTES=$(cat "$TEMPLATE_FILE")
    # 替换占位符
    RELEASE_NOTES="${RELEASE_NOTES//\{\{VERSION\}\}/$VERSION}"
    RELEASE_NOTES="${RELEASE_NOTES//\{\{PREVIOUS_VERSION\}\}/$PREVIOUS_VERSION}"
else
    RELEASE_NOTES="## $VERSION\n\n### 变更内容\n\n$CHANGELOG"
fi

# 如果提供了自定义说明，添加到开头
if [ -n "$NOTES" ]; then
    RELEASE_NOTES="$NOTES\n\n$RELEASE_NOTES"
fi

# 创建临时文件
TEMP_FILE=$(mktemp)
echo -e "$RELEASE_NOTES" > "$TEMP_FILE"

# 显示预览
echo -e "\n${YELLOW}Release 说明预览:${NC}"
echo "----------------------------------------"
cat "$TEMP_FILE"
echo "----------------------------------------"

# 确认
read -p "是否创建 Release? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    rm "$TEMP_FILE"
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 创建标签
echo -e "\n${GREEN}创建标签 $VERSION...${NC}"
git tag -a "$VERSION" -m "Release $VERSION"

# 推送标签
echo -e "${GREEN}推送标签到远程...${NC}"
git push origin "$VERSION"

# 创建 GitHub Release
echo -e "${GREEN}创建 GitHub Release...${NC}"
gh release create "$VERSION" \
    --title "$VERSION" \
    --notes-file "$TEMP_FILE"

# 清理
rm "$TEMP_FILE"

echo -e "\n${GREEN}✅ Release $VERSION 创建成功!${NC}"
echo -e "查看: https://github.com/f2xme/gox/releases/tag/$VERSION"
