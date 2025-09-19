#!/bin/bash
### BEGIN INIT INFO
# Provides:          app_service
# Required-Start:    $local_fs $network $named $time $syslog
# Required-Stop:     $local_fs $network $named $time $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: 管理 app_mng 服务
# Description:       启动、停止并报告 /root/app_mng 的状态
### END INIT INFO

#
# 这是一个用于管理 /root/app_mng 的 SysV 风格 init 脚本
# 假设应用程序的可执行文件或脚本位于 /root/app_mng
# 支持命令: start, stop, restart, status
#

APP_DIR="/root"
APP_BIN="$APP_DIR/app_mng"
DAEMON_NAME="app_mng"
PIDFILE="/var/run/${DAEMON_NAME}.pid"
LOGFILE="/var/log/${DAEMON_NAME}.log"

start() {
    if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" >/dev/null 2>&1; then
        echo "$DAEMON_NAME 已在运行 (pid $(cat "$PIDFILE"))"
        return 0
    fi

    if [ ! -x "$APP_BIN" ]; then
        if [ -f "$APP_BIN" ]; then
            echo "找到 $APP_BIN，但不可执行。尝试设置可执行权限..."
            chmod +x "$APP_BIN" || { echo "无法将 $APP_BIN 设为可执行"; return 1; }
        else
            echo "未找到 $APP_BIN。请确保应用二进制或脚本存在于 $APP_BIN"
            return 1
        fi
    fi

    echo "正在启动 $DAEMON_NAME..."
    mkdir -p "$(dirname "$PIDFILE")" || true
    nohup "$APP_BIN" >>"$LOGFILE" 2>&1 &
    echo $! > "$PIDFILE"
    sleep 1
    if kill -0 "$(cat "$PIDFILE")" >/dev/null 2>&1; then
        echo "$DAEMON_NAME 已启动，pid $(cat "$PIDFILE")"
        return 0
    else
        echo "启动 $DAEMON_NAME 失败。请检查 $LOGFILE"
        rm -f "$PIDFILE"
        return 1
    fi
}

stop() {
    if [ ! -f "$PIDFILE" ]; then
        echo "$DAEMON_NAME 未在运行（没有 pid 文件）"
        return 0
    fi
    PID=$(cat "$PIDFILE")
    if ! kill -0 "$PID" >/dev/null 2>&1; then
        echo "进程 $PID 未在运行，移除过时的 pid 文件"
        rm -f "$PIDFILE"
        return 0
    fi
    echo "正在停止 $DAEMON_NAME (pid $PID) ..."
    kill "$PID"
    # 等待最多 10 秒钟
    for i in {1..10}; do
        if kill -0 "$PID" >/dev/null 2>&1; then
            sleep 1
        else
            break
        fi
    done
    if kill -0 "$PID" >/dev/null 2>&1; then
        echo "$DAEMON_NAME 未正常退出，发送 SIGKILL"
        kill -9 "$PID" >/dev/null 2>&1 || true
        sleep 1
    fi
    rm -f "$PIDFILE"
    echo "$DAEMON_NAME 已停止"
    return 0
}

status() {
    if [ -f "$PIDFILE" ]; then
        PID=$(cat "$PIDFILE")
        if kill -0 "$PID" >/dev/null 2>&1; then
            echo "$DAEMON_NAME 正在运行 (pid $PID)"
            return 0
        else
            echo "$DAEMON_NAME 的 pid 文件存在，但进程未运行"
            return 1
        fi
    fi
    # 作为回退：尝试使用 pgrep
    if pgrep -f "$APP_BIN" >/dev/null 2>&1; then
        echo "$DAEMON_NAME 正在运行"
        return 0
    fi
    echo "$DAEMON_NAME 未在运行"
    return 3
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    status)
        status
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status}"
        exit 2
        ;;
esac

exit 0
