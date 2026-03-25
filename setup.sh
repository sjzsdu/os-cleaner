#!/bin/bash

# OS Cleaner 安装脚本
# 编译并安装 os-cleaner CLI 到 ~/.local/bin

set -e

echo "========================================"
echo "OS Cleaner 安装脚本"
echo "========================================"
echo ""

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go"
    echo "下载地址: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Go 版本: $GO_VERSION"
echo ""

echo "构建中..."
go build -o os-cleaner .
echo ""

# 创建安装目录
mkdir -p "$HOME/.local/bin"

# 复制二进制文件
cp os-cleaner "$HOME/.local/bin/"
chmod +x "$HOME/.local/bin/os-cleaner"

echo "========================================"
echo "安装完成！"
echo "========================================"
echo ""
ls -lh "$HOME/.local/bin/os-cleaner"
echo ""
echo "使用方法:"
echo "  os-cleaner scan           # 扫描所有缓存"
echo "  os-cleaner list           # 列出所有类别"
echo "  os-cleaner clean <类别>   # 清理指定类别"
echo "  os-cleaner clean --safe   # 清理所有安全类别"
echo ""
echo "示例:"
echo "  os-cleaner scan                         # 扫描"
echo "  os-cleaner clean npm-cache --dry-run    # 预览清理 npm"
echo "  os-cleaner clean --safe                # 清理所有安全类别"
echo "  os-cleaner clean xcode --recoverable   # 可恢复删除"
echo ""
echo "提示:"
echo "  - 使用 --dry-run 预览而不实际删除"
echo "  - 使用 --recoverable 在删除前压缩备份"
echo "  - 如需永久生效，添加到 ~/.zshrc:"
echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
