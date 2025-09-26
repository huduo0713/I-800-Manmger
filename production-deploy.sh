#!/bin/bash
# 生产环境部署脚本 - 自动构建并部署Edge Device Manager
# 支持版本信息注入和服务管理

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置变量
APP_NAME="edge-device"
SERVICE_NAME="edge-device"
INSTALL_DIR="/opt/${APP_NAME}"
CONFIG_DIR="${INSTALL_DIR}/config"
LOG_DIR="${INSTALL_DIR}/logs"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# 显示帮助信息
show_help() {
    cat <<EOF
🚀 Edge Device Manager 部署脚本

用法: $0 <命令> [选项]

命令:
  build       仅构建程序（带版本信息）
  deploy      构建并部署到生产环境
  install     安装systemd服务
  start       启动服务
  stop        停止服务
  restart     重启服务
  status      查看服务状态
  logs        查看服务日志
  uninstall   卸载服务和程序
  update      更新程序（保留配置）

选项:
  --force     强制操作，跳过确认
  --dry-run   仅显示操作，不执行
  --help      显示此帮助

示例:
  $0 build              # 仅构建程序
  $0 deploy --force     # 强制部署
  $0 update            # 更新程序
EOF
}

# 检查权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}❌ 此操作需要root权限${NC}"
        echo "请使用: sudo $0 $*"
        exit 1
    fi
}

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装或不在PATH中${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Go环境: $(go version)${NC}"
}

# 构建程序
build_app() {
    echo -e "${BLUE}🔨 构建 ${APP_NAME}...${NC}"
    
    # 获取版本信息
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    GO_VERSION=$(go version | cut -d' ' -f3)
    
    echo -e "${CYAN}📋 版本信息:${NC}"
    echo "  • Git提交: ${GIT_COMMIT}"
    echo "  • Git分支: ${GIT_BRANCH}"
    echo "  • 构建时间: ${BUILD_TIME}"
    echo "  • Go版本: ${GO_VERSION}"
    
    # 构建
    LDFLAGS="-X 'demo/internal/consts.BuildTime=${BUILD_TIME}' \
             -X 'demo/internal/consts.GitCommit=${GIT_COMMIT}' \
             -X 'demo/internal/consts.GitBranch=${GIT_BRANCH}' \
             -X 'demo/internal/consts.GoVersion=${GO_VERSION}'"
    
    go build -ldflags "${LDFLAGS}" -o "${APP_NAME}" .
    
    if [ -f "${APP_NAME}" ]; then
        echo -e "${GREEN}✅ 构建成功: ${APP_NAME}${NC}"
        echo "📦 文件大小: $(ls -lh ${APP_NAME} | awk '{print $5}')"
    else
        echo -e "${RED}❌ 构建失败${NC}"
        exit 1
    fi
}

# 部署程序
deploy_app() {
    echo -e "${BLUE}📦 部署 ${APP_NAME}...${NC}"
    
    # 创建目录
    mkdir -p "${INSTALL_DIR}" "${CONFIG_DIR}" "${LOG_DIR}"
    
    # 复制程序
    cp "${APP_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${APP_NAME}"
    
    # 复制配置文件
    if [ -d "manifest/config" ]; then
        cp -r manifest/config/* "${CONFIG_DIR}/"
        echo -e "${GREEN}✅ 配置文件已复制到 ${CONFIG_DIR}${NC}"
    fi
    
    # 设置权限
    chown -R root:root "${INSTALL_DIR}"
    chmod 755 "${INSTALL_DIR}"
    
    echo -e "${GREEN}✅ 部署完成${NC}"
    echo "📁 安装目录: ${INSTALL_DIR}"
    echo "⚙️  配置目录: ${CONFIG_DIR}"
    echo "📄 日志目录: ${LOG_DIR}"
}

# 创建systemd服务
install_service() {
    echo -e "${BLUE}🔧 安装systemd服务...${NC}"
    
    cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=Edge Device Manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${APP_NAME}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 环境变量
Environment="GF_GCFG_PATH=${CONFIG_DIR}"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"
    
    echo -e "${GREEN}✅ 服务安装完成${NC}"
    echo "🔧 服务名称: ${SERVICE_NAME}"
    echo "📄 服务文件: ${SERVICE_FILE}"
}

# 主函数
main() {
    case "${1:-}" in
        "build")
            check_go
            build_app
            ;;
        "deploy")
            check_root
            check_go
            build_app
            deploy_app
            ;;
        "install")
            check_root
            install_service
            ;;
        "start")
            check_root
            systemctl start "${SERVICE_NAME}"
            echo -e "${GREEN}✅ 服务已启动${NC}"
            ;;
        "stop")
            check_root
            systemctl stop "${SERVICE_NAME}"
            echo -e "${YELLOW}⏹️  服务已停止${NC}"
            ;;
        "restart")
            check_root
            systemctl restart "${SERVICE_NAME}"
            echo -e "${GREEN}🔄 服务已重启${NC}"
            ;;
        "status")
            systemctl status "${SERVICE_NAME}"
            ;;
        "logs")
            journalctl -u "${SERVICE_NAME}" -f
            ;;
        "update")
            check_root
            check_go
            systemctl stop "${SERVICE_NAME}" 2>/dev/null || true
            build_app
            cp "${APP_NAME}" "${INSTALL_DIR}/"
            chmod +x "${INSTALL_DIR}/${APP_NAME}"
            systemctl start "${SERVICE_NAME}"
            echo -e "${GREEN}✅ 程序已更新并重启${NC}"
            ;;
        "uninstall")
            check_root
            systemctl stop "${SERVICE_NAME}" 2>/dev/null || true
            systemctl disable "${SERVICE_NAME}" 2>/dev/null || true
            rm -f "${SERVICE_FILE}"
            rm -rf "${INSTALL_DIR}"
            systemctl daemon-reload
            echo -e "${GREEN}✅ 卸载完成${NC}"
            ;;
        "--help"|"-h"|"help")
            show_help
            ;;
        *)
            echo -e "${RED}❌ 未知命令: ${1:-}${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"