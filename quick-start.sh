#!/bin/bash
# 快速构建并运行 - Linux版本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Edge Device Manager - 快速构建运行${NC}"
echo "=================================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go未安装或不在PATH中${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go环境检查通过: $(go version)${NC}"

# 获取版本信息
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

# 显示构建信息
echo -e "${YELLOW}📋 构建信息:${NC}"
echo "  • Git提交: ${GIT_COMMIT}"
echo "  • Git分支: ${GIT_BRANCH}"  
echo "  • 构建时间: ${BUILD_TIME}"
echo "  • Go版本: ${GO_VERSION}"
echo ""

# 构建程序
echo -e "${BLUE}🔨 开始构建...${NC}"
LDFLAGS="-X 'demo/internal/consts.BuildTime=${BUILD_TIME}' -X 'demo/internal/consts.GitCommit=${GIT_COMMIT}' -X 'demo/internal/consts.GitBranch=${GIT_BRANCH}' -X 'demo/internal/consts.GoVersion=${GO_VERSION}'"

if go build -ldflags "${LDFLAGS}" -o edge-device .; then
    echo -e "${GREEN}✅ 构建成功!${NC}"
    echo "📦 可执行文件: $(pwd)/edge-device"
    echo "📏 文件大小: $(ls -lh edge-device | awk '{print $5}')"
    echo ""
    
    # 询问是否运行
    read -p "🤔 是否立即运行程序? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${GREEN}🚀 启动程序...${NC}"
        echo "=================================="
        ./edge-device
    else
        echo -e "${YELLOW}💡 手动运行请执行: ./edge-device${NC}"
    fi
else
    echo -e "${RED}❌ 构建失败${NC}"
    exit 1
fi