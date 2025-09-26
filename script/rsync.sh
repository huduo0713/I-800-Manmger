#!/usr/bin/env bash
# ==================================================
# 使用 scp 将指定文件/目录传到远端主机（目标主机无 rsync）
# 要求：远端可通过 SSH 登录；scp 命令在本地可用；端口默认为 22。
# 目标已固定为：192.168.11.204:/root/
# 脚本行为：收集存在的项（app_mng 文件，data/init.sql、manifest/、script/ 目录），
# 排除 data/sqlite.db 数据库文件，只同步数据库初始化文件。
#
# 使用方法：
#   ./rsync.sh              # 自动检测远程环境
#   FORCE_SCP=1 ./rsync.sh  # 强制使用scp模式（适用于BusyBox环境）
# ==================================================

set -euo pipefail

# 配置（固定为你的要求）
SSH_PORT=22
REMOTE_USER="root"
REMOTE_HOST="192.168.11.203"
REMOTE_PATH="/root/"

# SSH连接复用配置
SSH_CONTROL_PATH="/tmp/rsync_ssh_$$"
SSH_OPTS="-p ${SSH_PORT} -o ControlMaster=auto -o ControlPath=${SSH_CONTROL_PATH} -o ControlPersist=60"

ROOT_DIR="$(pwd)"
APP_FILE="app_mng"
DIRS=("manifest" "script")  # 移除data，将单独处理
DATA_FILES=("data/init.sql")  # 只同步数据库初始化文件

echo "工作目录: $ROOT_DIR"
echo "准备传输文件到远程主机（排除 data/sqlite.db 数据库文件）"
echo ""

# 检查要传输的文件是否存在
echo "检查源文件："
if [ -f "${ROOT_DIR}/${APP_FILE}" ]; then
    echo " ✓ 主程序: $APP_FILE"
else
    echo " ✗ 主程序文件不存在: $APP_FILE"
fi

for d in "${DIRS[@]}"; do
    if [ -d "${ROOT_DIR}/${d}" ]; then
        echo " ✓ 目录: $d/"
    else
        echo " ✗ 目录不存在: $d/"
    fi
done

for f in "${DATA_FILES[@]}"; do
    if [ -f "${ROOT_DIR}/${f}" ]; then
        echo " ✓ 数据文件: $f"
    else
        echo " ✗ 数据文件不存在: $f"
    fi
done

echo ""

echo "目标: ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH} (ssh端口 ${SSH_PORT})"

# 检查必要工具
if ! command -v scp >/dev/null 2>&1; then
  echo "错误：未检测到 scp。请先安装 OpenSSH 客户端（例如 apt install openssh-client）。"
  exit 2
fi

if ! command -v ssh >/dev/null 2>&1; then
  echo "错误：未检测到 ssh。请先安装 OpenSSH 客户端。"
  exit 2
fi

# 检查是否配置了SSH密钥
echo "💡 优化提示："
if [ -f ~/.ssh/id_rsa ] || [ -f ~/.ssh/id_ed25519 ]; then
    echo "   ✓ 检测到SSH密钥文件，可能无需输入密码"
else
    echo "   建议配置SSH密钥认证以避免输入密码："
    echo "   1. 生成密钥: ssh-keygen -t ed25519"
    echo "   2. 复制公钥: ssh-copy-id ${REMOTE_USER}@${REMOTE_HOST}"
    echo "   3. 下次传输将无需输入密码"
fi
echo ""

echo "开始传输（使用SSH连接复用，只需要输入一次密码）..."

# 创建临时目录用于打包
TEMP_DIR="/tmp/deploy_package_$$"
echo "创建临时打包目录: $TEMP_DIR"
mkdir -p "$TEMP_DIR"

# 函数：清理临时文件
cleanup() {
    echo "清理临时文件..."
    rm -rf "$TEMP_DIR"
    # 关闭SSH连接复用
    ssh ${SSH_OPTS} -O exit "${REMOTE_USER}@${REMOTE_HOST}" 2>/dev/null || true
}
trap cleanup EXIT

# 复制需要传输的文件到临时目录，保持目录结构
echo "准备文件包..."

# 复制主程序文件
if [ -f "${ROOT_DIR}/${APP_FILE}" ]; then
    echo "  添加主程序: ${APP_FILE}"
    cp "${ROOT_DIR}/${APP_FILE}" "$TEMP_DIR/"
fi

# 复制目录
for d in "${DIRS[@]}"; do
    if [ -d "${ROOT_DIR}/${d}" ]; then
        echo "  添加目录: ${d}/"
        cp -r "${ROOT_DIR}/${d}" "$TEMP_DIR/"
    fi
done

# 复制data目录文件，保持目录结构
for f in "${DATA_FILES[@]}"; do
    if [ -f "${ROOT_DIR}/${f}" ]; then
        echo "  添加数据文件: ${f}"
        # 创建目标目录
        file_dir=$(dirname "${f}")
        mkdir -p "$TEMP_DIR/$file_dir"
        # 复制文件
        cp "${ROOT_DIR}/${f}" "$TEMP_DIR/${f}"
    fi
done

# 建立SSH连接（第一次输入密码）
echo ""
echo "建立SSH连接（请输入密码）..."
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "echo 'SSH连接已建立'"

# 创建远程目录结构
echo "创建远程目录结构..."
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "mkdir -p ${REMOTE_PATH}data"

# 打包并传输（兼容BusyBox tar）
echo "打包并传输文件..."
cd "$TEMP_DIR"

# 检查远程tar版本并选择合适的命令
echo "检测远程环境..."
REMOTE_TAR_VERSION=$(ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "tar --version 2>/dev/null | head -1 || echo 'BusyBox'")
echo "远程tar版本: $REMOTE_TAR_VERSION"

# 提供强制使用scp模式的选项
if [ "${FORCE_SCP:-}" = "1" ]; then
    echo "强制使用scp模式..."
    REMOTE_TAR_VERSION="BusyBox"
fi

if echo "$REMOTE_TAR_VERSION" | grep -qi "busybox"; then
    echo "检测到BusyBox环境，使用scp传输（兼容性最佳）..."

    # 使用scp传输，利用SSH连接复用
    scp_opts="-o ControlMaster=auto -o ControlPath=${SSH_CONTROL_PATH} -o ControlPersist=60 -P ${SSH_PORT}"

    # 预先创建所有需要的目录
    echo "  创建远程目录结构..."
    find . -type d | while read dir; do
        if [ "$dir" != "." ]; then
            relative_dir="${dir#./}"
            ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "mkdir -p ${REMOTE_PATH}${relative_dir}"
        fi
    done

    # 批量传输文件
    echo "  批量传输文件..."
    file_count=$(find . -type f | wc -l)
    current=0

    find . -type f | while read file; do
        current=$((current + 1))
        relative_file="${file#./}"
        echo "  [$current/$file_count] 传输: $relative_file"
        scp ${scp_opts} "$file" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}${relative_file}" 2>/dev/null || {
            echo "    警告: $relative_file 传输失败，重试..."
            scp ${scp_opts} "$file" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}${relative_file}"
        }
    done
else
    echo "使用GNU tar压缩传输..."
    # GNU tar 支持压缩，更高效
    tar -czf - . | ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "cd ${REMOTE_PATH} && tar -xzf -"
fi

# 检查传输是否成功
if [ $? -eq 0 ]; then
    echo "✓ 文件传输完成！"
else
    echo "✗ 传输过程中发生错误！"
    exit 1
fi

rc=$?
if [ $rc -ne 0 ]; then
  echo "传输失败，scp 返回码：$rc"
  exit $rc
fi

# 设置正确的文件权限（使用已建立的连接）
echo "设置文件权限..."
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "
    echo '正在设置文件权限...'

    # 设置主程序可执行权限
    if [ -f ${REMOTE_PATH}${APP_FILE} ]; then
        chmod +x ${REMOTE_PATH}${APP_FILE}
        echo '✓ 已设置 ${APP_FILE} 可执行权限'
    fi

    # 设置脚本文件可执行权限
    if [ -d ${REMOTE_PATH}script/ ]; then
        find ${REMOTE_PATH}script/ -name '*.sh' -type f -exec chmod +x {} \; 2>/dev/null || true
        echo '✓ 已设置脚本文件可执行权限'
    fi

    # 确保data目录下的文件不是可执行的（应该是数据文件）
    if [ -d ${REMOTE_PATH}data/ ]; then
        find ${REMOTE_PATH}data/ -type f -exec chmod 644 {} \; 2>/dev/null || true
        echo '✓ 已设置data目录文件权限为644'
    fi

    echo '文件权限设置完成'
"# 验证传输结果
echo ""
echo "验证传输结果..."
ssh ${SSH_OPTS} "${REMOTE_USER}@${REMOTE_HOST}" "
    echo '远程文件清单:'
    echo '主程序文件:'
    ls -la ${REMOTE_PATH}${APP_FILE} 2>/dev/null || echo '  ${APP_FILE} 不存在'

    echo '目录结构:'
    for dir in manifest script data; do
        if [ -d ${REMOTE_PATH}\${dir} ]; then
            echo \"  \${dir}/ (\$(ls ${REMOTE_PATH}\${dir} | wc -l) 个文件)\"
        else
            echo \"  \${dir}/ 不存在\"
        fi
    done

    echo 'data目录内容:'
    ls -la ${REMOTE_PATH}data/ 2>/dev/null || echo '  data目录为空或不存在'
"

echo ""
echo "🎉 全部传输完成！"
echo "📋 传输摘要:"
echo "   - 只需要输入一次SSH密码"
echo "   - 使用tar打包提高传输效率"
echo "   - 自动设置正确的文件权限"
echo "   - 排除了数据库文件 sqlite.db"
echo ""
echo "💡 提示：若收到 SSH 密钥或主机指纹提示，请按提示操作以建立信任（首次连接时会出现）"
echo "🚀 现在可以登录远程服务器运行程序：ssh ${REMOTE_USER}@${REMOTE_HOST}"

exit 0
