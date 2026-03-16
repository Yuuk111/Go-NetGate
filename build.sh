#!/bin/bash
# 构建 Go 项目
# --- 配置区 ---
APP_NAME="NetGate"
SOURCE_PATH="."

# --- 构建区 ---
echo -e "\u231B 正在构建 $APP_NAME..."
# 清理旧的构建产物
if [ -f "${APP_NAME}" ]; then
    echo -e "\u26A0 清理旧的构建产物..."
    rm "${APP_NAME}"
fi

# 构建新的可执行文件
go build -o "${APP_NAME}" "${SOURCE_PATH}"

# 检查构建结果
if [ -f "${APP_NAME}" ]; then
    echo "✅ ${APP_NAME} 构建成功！"
    chmod +x "${APP_NAME}"
else
    echo "❌ ${APP_NAME} 构建失败！"
    exit 1
fi