#!/bin/bash
# 开发环境启动脚本

# 使用 gf run 命令启动应用（自动查找配置文件）
go build -o i800 .
./i800
#gf run main.go --- IGNORE ---
