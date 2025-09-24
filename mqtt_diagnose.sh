#!/bin/bash

# MQTT连接诊断脚本
echo "🔍 MQTT连接诊断工具"
echo "==================="

# 从配置文件提取MQTT服务器信息（简化版本，实际可以用更复杂的解析）
BROKER_IP="192.168.11.204"
BROKER_PORT="1883"

echo "📡 目标MQTT服务器: $BROKER_IP:$BROKER_PORT"
echo ""

# 1. 检查网络连通性
echo "1️⃣ 检查网络连通性..."
if ping -c 3 -W 3 $BROKER_IP > /dev/null 2>&1; then
    echo "   ✅ 网络连通正常"
else
    echo "   ❌ 网络不通，请检查："
    echo "      - 网络配置是否正确"
    echo "      - 防火墙是否阻挡"
    echo "      - MQTT服务器是否在运行"
    exit 1
fi

# 2. 检查端口连通性
echo "2️⃣ 检查MQTT端口连通性..."
if command -v nc > /dev/null 2>&1; then
    if nc -z -w5 $BROKER_IP $BROKER_PORT 2>/dev/null; then
        echo "   ✅ MQTT端口 $BROKER_PORT 可访问"
    else
        echo "   ❌ MQTT端口 $BROKER_PORT 无法访问"
        echo "      请检查MQTT服务是否启动"
        exit 1
    fi
elif command -v telnet > /dev/null 2>&1; then
    # 使用telnet检查（超时处理比较复杂，简化处理）
    timeout 5 telnet $BROKER_IP $BROKER_PORT < /dev/null 2>/dev/null | grep -q "Connected"
    if [ $? -eq 0 ]; then
        echo "   ✅ MQTT端口 $BROKER_PORT 可访问"
    else
        echo "   ⚠️ 无法确定端口状态（telnet测试不够准确）"
    fi
else
    echo "   ⚠️ 无nc或telnet工具，跳过端口检查"
fi

# 3. DNS解析检查（如果使用域名）
echo "3️⃣ 检查DNS解析..."
if nslookup $BROKER_IP > /dev/null 2>&1; then
    echo "   ✅ DNS解析正常"
else
    echo "   ℹ️ 使用IP地址，无需DNS解析"
fi

echo ""
echo "🎯 诊断建议："
echo "   如果网络和端口都正常但仍连接失败，请检查："
echo "   1. MQTT broker是否正确配置和运行"
echo "   2. 防火墙规则是否允许1883端口"
echo "   3. MQTT broker的认证设置"
echo "   4. 网络延迟是否过高"
echo ""
echo "📝 运行程序时建议："
echo "   - 观察日志中的详细错误信息"
echo "   - 检查是否出现'🟢 MQTT连接成功'日志"
echo "   - 如果长时间卡住，可能需要调整connectTimeout配置"