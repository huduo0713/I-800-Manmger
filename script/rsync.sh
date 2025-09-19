#!/usr/bin/env bash
# ==================================================
# 使用 scp 将指定文件/目录传到远端主机（目标主机无 rsync）
# 要求：远端可通过 SSH 登录；scp 命令在本地可用；端口默认为 22。
# 目标已固定为：192.168.11.204:/root/
# 脚本行为：收集存在的项（app_mng 文件，data/、manifest/ 目录），
# 使用单次 scp 调用传输所有项，从而只提示一次密码。
# ==================================================

set -euo pipefail

# 配置（固定为你的要求）
SSH_PORT=22
REMOTE_USER="root"
REMOTE_HOST="192.168.11.204"
REMOTE_PATH="/root/"

ROOT_DIR="$(pwd)"
APP_FILE="app_mng"
DIRS=("data" "manifest" "script")

SOURCES=(app_mng deploy.sh app_service.sh data/ manifest/)

echo "工作目录: $ROOT_DIR"
echo "准备收集要传输的项（将一次性通过 scp 上传）："

# 处理 app_mng（期望为文件）
SRC_APP="${ROOT_DIR}/${APP_FILE}"
if [ -e "$SRC_APP" ]; then
  if [ -f "$SRC_APP" ]; then
    echo " - 文件: $APP_FILE"
    SOURCES+=("$SRC_APP")
  else
    echo "警告：$APP_FILE 存在但不是普通文件，已跳过。"
  fi
else
  echo "警告：源文件不存在，跳过： $SRC_APP"
fi

# 处理目录
for d in "${DIRS[@]}"; do
  SRC_DIR="${ROOT_DIR}/${d}"
  if [ -e "$SRC_DIR" ]; then
    if [ -d "$SRC_DIR" ]; then
      echo " - 目录: $d"
      SOURCES+=("$SRC_DIR")
    else
      echo "警告：$SRC_DIR 存在但不是目录，已跳过。"
    fi
  else
    echo "警告：源目录不存在，跳过： $SRC_DIR"
  fi
done

if [ ${#SOURCES[@]} -eq 0 ]; then
  echo "没有找到任何要传输的项目，退出。"
  exit 0
fi

echo "将要一次性传输以下项（会提示一次密码）："
for s in "${SOURCES[@]}"; do
  echo " - $s"


echo "目标: ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH} (ssh端口 ${SSH_PORT})"

# 检查 scp 是否存在
if ! command -v scp >/dev/null 2>&1; then
  echo "错误：未检测到 scp。请先安装 OpenSSH 客户端（例如 apt install openssh-client）。"
  exit 2
fi

echo "开始传输（scp -r -P ${SSH_PORT} ...）..."
scp -r -P ${SSH_PORT} "${SOURCES[@]}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}"
rc=$?
if [ $rc -ne 0 ]; then
  echo "传输失败，scp 返回码：$rc"
  exit $rc
fi

echo "全部传输完成。"
echo "注意：若收到 SSH 密钥或主机指纹提示，请按提示操作以建立信任（首次连接时会出现）。"

exit 0
done
