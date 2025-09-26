#!/bin/bash
# Linux构建脚本 - 包含版本信息

set -e

# 可执行文件名变量
APP_NAME="app_mng"

echo "🔨 开始构建 Edge Device Manager (Linux)..."

# 获取版本信息
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

# 构建ldflags
LDFLAGS="-X 'demo/internal/consts.BuildTime=${BUILD_TIME}' \
         -X 'demo/internal/consts.GitCommit=${GIT_COMMIT}' \
         -X 'demo/internal/consts.GitBranch=${GIT_BRANCH}' \
         -X 'demo/internal/consts.GoVersion=${GO_VERSION}'"

echo "📋 版本信息:"
echo "  Git Commit: ${GIT_COMMIT}"
echo "  Git Branch: ${GIT_BRANCH}"
echo "  Build Time: ${BUILD_TIME}"
echo "  Go Version: ${GO_VERSION}"
echo ""

# 执行构建
echo "🔄 正在编译..."
go build -ldflags "${LDFLAGS}" -o "${APP_NAME}" .

if [ $? -eq 0 ]; then
    echo "✅ 构建成功: ${APP_NAME}"
    echo ""
    echo "🚀 运行程序请执行: ./${APP_NAME}"
    echo "🔍 查看版本信息: ./${APP_NAME} --help"
    echo "📁 当前目录: $(pwd)"
    echo "📦 文件大小: $(ls -lh ${APP_NAME} | awk '{print $5}')"
else
    echo "❌ 构建失败"
    exit 1
fi
