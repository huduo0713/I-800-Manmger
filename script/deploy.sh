#!/bin/bash
# 简易部署脚本：用于安装 app_mng 的 SysV init 脚本
# 用法：
#   ./deploy.sh install    # 将 app_service.sh 复制到 /etc/init.d/S99app_service 并启用
#   ./deploy.sh remove     # 停止服务并移除 init 脚本
#   ./deploy.sh start|stop|restart|status  # 代理控制已安装的服务

set -euo pipefail

SCRIPT_SRC="$(cd "$(dirname "$0")" && pwd)/app_service.sh"
TARGET_DIR="/etc/init.d"
TARGET_NAME="S99app_service"
TARGET_PATH="$TARGET_DIR/$TARGET_NAME"

help() {
    cat <<EOF
用法: $0 <命令>

命令列表:
  install    将 app_service.sh 复制到 $TARGET_PATH 并设置为可执行
  remove     停止服务（如果在运行）并移除 $TARGET_PATH
  start|stop|restart|status  控制已安装的服务
  --dry-run  仅打印将执行的操作，不做任何修改
  -h|--help  显示此帮助信息
EOF
}

DRY_RUN=0
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN=1
    shift
fi

cmd=${1:-}
if [ -z "$cmd" ]; then
    help
    exit 2
fi

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        echo "DRY-RUN: $*"
    else
        echo "+ $*"
        bash -c "$*"
    fi
}

case "$cmd" in
    install)
        if [ ! -f "$SCRIPT_SRC" ]; then
            echo "未找到源脚本 $SCRIPT_SRC。请将 app_service.sh 放在与 deploy.sh 相同目录下。"
            exit 1
        fi
        run "cp -f '$SCRIPT_SRC' '$TARGET_PATH'"
        run "chmod 755 '$TARGET_PATH'"
        echo "已安装 $TARGET_PATH"
        run "'$TARGET_PATH' start || true"
        echo "已启动 $TARGET_PATH start"
        echo "查看日志：tail -f /var/log/app_mng.log"

        ;;
    remove)
        if [ -f "$TARGET_PATH" ]; then
            run "'$TARGET_PATH' stop || true"
            run "rm -f '$TARGET_PATH'"
            echo "已移除 $TARGET_PATH"
        else
            echo "$TARGET_PATH 不存在"
            exit 0
        fi
        ;;
    -h|--help)
        help
        ;;
    *)
        echo "未知命令: $cmd"
        help
        exit 2
        ;;
esac

exit 0
