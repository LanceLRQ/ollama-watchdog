#!/bin/bash

# GitHub 仓库信息
REPO_OWNER="LanceLRQ"
REPO_NAME="ollama-watchdog"
RELEASE_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"

# 默认安装路径
INSTALL_DIR="/opt/ollama-watchdog"
SERVICE_NAME="ollama-watchdog.service"
SERVICE_PATH="/etc/systemd/system/$SERVICE_NAME"
BIN_NAME="ollama-watchdog"
BIN_PATH="$INSTALL_DIR/$BIN_NAME"
SYMLINK_PATH="/usr/bin/$BIN_NAME"

# 检查系统是否为 Linux 和 amd64 架构
OS=$(uname -s)
ARCH=$(uname -m)
if [ "$OS" != "Linux" ] || [ "$ARCH" != "x86_64" ]; then
    echo "错误：此脚本仅支持 Linux amd64 环境。"
    exit 1
fi

# 检查是否已经安装
if [ -d "$INSTALL_DIR" ]; then
    echo "检测到 Ollama Watchdog 已经安装在 $INSTALL_DIR。"
    read -p "请选择操作：[1] 覆盖安装 [2] 卸载 [3] 退出: " choice

    case $choice in
        1)
            echo "开始覆盖安装..."
            # 停止并禁用服务
            sudo systemctl stop $SERVICE_NAME > /dev/null 2>&1
            sudo systemctl disable $SERVICE_NAME > /dev/null 2>&1
            # 删除旧文件和软链接
            sudo rm -rf $INSTALL_DIR
            sudo rm -f $SYMLINK_PATH
            ;;
        2)
            echo "开始卸载..."
            # 停止并禁用服务
            sudo systemctl stop $SERVICE_NAME > /dev/null 2>&1
            sudo systemctl disable $SERVICE_NAME > /dev/null 2>&1
            # 删除安装目录、服务文件和软链接
            sudo rm -rf $INSTALL_DIR
            sudo rm -f $SERVICE_PATH
            sudo rm -f $SYMLINK_PATH
            echo "卸载完成。"
            exit 0
            ;;
        3)
            echo "退出安装。"
            exit 0
            ;;
        *)
            echo "无效的选择，退出安装。"
            exit 1
            ;;
    esac
fi

# 提示用户修改安装路径
read -p "请输入安装路径（默认: $INSTALL_DIR）: " user_dir
if [ -n "$user_dir" ]; then
    INSTALL_DIR=$user_dir
fi

# 创建安装目录
echo "正在安装到 $INSTALL_DIR..."
sudo mkdir -p $INSTALL_DIR
if [ $? -ne 0 ]; then
    echo "创建安装目录失败，请检查权限。"
    exit 1
fi

# 从 GitHub Release 下载最新版本
echo "正在从 GitHub Release 下载最新版本..."
DOWNLOAD_URL=$(curl -s $RELEASE_URL | grep -oP '"browser_download_url": "\K[^"]+' | grep "linux-amd64")
if [ -z "$DOWNLOAD_URL" ]; then
    echo "错误：未找到适用于当前系统的发布文件。"
    exit 1
fi

# 下载文件
sudo curl -L -o $BIN_PATH $DOWNLOAD_URL
if [ $? -ne 0 ]; then
    echo "下载失败，请检查网络连接。"
    exit 1
fi

# 赋予可执行权限
sudo chmod +x $BIN_PATH
if [ $? -ne 0 ]; then
    echo "赋予可执行权限失败。"
    exit 1
fi

# 创建软链接到 /usr/bin
echo "正在创建软链接到 /usr/bin..."
sudo ln -sf $BIN_PATH $SYMLINK_PATH
if [ $? -ne 0 ]; then
    echo "创建软链接失败。"
    exit 1
fi

# 创建服务配置文件
echo "正在注册系统服务..."
cat <<EOF | sudo tee $SERVICE_PATH > /dev/null
[Unit]
Description=Ollama Watchdog Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_PATH
Restart=on-failure
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=ollama-watchdog

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd 配置
sudo systemctl daemon-reload
if [ $? -ne 0 ]; then
    echo "重新加载 systemd 配置失败。"
    exit 1
fi

# 启用并启动服务
sudo systemctl enable $SERVICE_NAME
sudo systemctl start $SERVICE_NAME
if [ $? -ne 0 ]; then
    echo "启动服务失败，请检查日志。"
    exit 1
fi

echo "安装完成！"
echo "服务已启动并设置为开机自启。"
echo "你可以通过以下命令管理服务："
echo "  sudo systemctl start $SERVICE_NAME    # 启动服务"
echo "  sudo systemctl stop $SERVICE_NAME     # 停止服务"
echo "  sudo systemctl status $SERVICE_NAME   # 查看服务状态"
echo "  sudo journalctl -u $SERVICE_NAME      # 查看服务日志"
echo "你可以直接运行 '$BIN_NAME' 启动程序。"
