#!/bin/bash

# PDF合并工具部署脚本

set -e

echo "🚀 PDF合并工具 - 部署脚本"
echo "========================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
APP_NAME="pdf-merger"
VERSION=${VERSION:-"latest"}
DEPLOY_ENV=${DEPLOY_ENV:-"production"}
DEPLOY_USER=${DEPLOY_USER:-"app"}
DEPLOY_PATH=${DEPLOY_PATH:-"/opt/pdf-merger"}
SERVICE_NAME="pdf-merger"
BACKUP_DIR="/opt/backups/pdf-merger"

# 部署配置
HEALTH_CHECK_URL=${HEALTH_CHECK_URL:-"http://localhost:8080/health"}
HEALTH_CHECK_TIMEOUT=${HEALTH_CHECK_TIMEOUT:-30}
ROLLBACK_ON_FAILURE=${ROLLBACK_ON_FAILURE:-true}

# 检查部署环境
check_deploy_environment() {
    echo "🔍 检查部署环境..."
    
    # 检查用户权限
    if [ "$EUID" -eq 0 ] && [ "$DEPLOY_USER" != "root" ]; then
        echo -e "${YELLOW}⚠️  以root用户运行，将切换到$DEPLOY_USER用户${NC}"
    fi
    
    # 检查部署目录
    if [ ! -d "$DEPLOY_PATH" ]; then
        echo "创建部署目录: $DEPLOY_PATH"
        sudo mkdir -p "$DEPLOY_PATH"
        sudo chown "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_PATH"
    fi
    
    # 检查备份目录
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "创建备份目录: $BACKUP_DIR"
        sudo mkdir -p "$BACKUP_DIR"
        sudo chown "$DEPLOY_USER:$DEPLOY_USER" "$BACKUP_DIR"
    fi
    
    # 检查systemd
    if ! command -v systemctl &> /dev/null; then
        echo -e "${YELLOW}⚠️  systemd不可用，将使用传统方式管理服务${NC}"
        USE_SYSTEMD=false
    else
        USE_SYSTEMD=true
    fi
    
    echo -e "${GREEN}✅ 部署环境检查完成${NC}"
    echo "部署环境: $DEPLOY_ENV"
    echo "部署路径: $DEPLOY_PATH"
    echo "部署用户: $DEPLOY_USER"
}

# 备份当前版本
backup_current_version() {
    echo "💾 备份当前版本..."
    
    if [ -f "$DEPLOY_PATH/$APP_NAME" ]; then
        local backup_name="$APP_NAME-$(date +%Y%m%d-%H%M%S)"
        local backup_path="$BACKUP_DIR/$backup_name"
        
        echo "创建备份: $backup_path"
        sudo -u "$DEPLOY_USER" cp -r "$DEPLOY_PATH" "$backup_path"
        
        # 保留最近5个备份
        sudo -u "$DEPLOY_USER" find "$BACKUP_DIR" -maxdepth 1 -type d -name "$APP_NAME-*" | \
            sort -r | tail -n +6 | xargs -r rm -rf
        
        echo -e "${GREEN}✅ 备份完成: $backup_path${NC}"
        echo "$backup_path" > /tmp/pdf-merger-backup-path
    else
        echo "未找到现有版本，跳过备份"
    fi
}

# 停止服务
stop_service() {
    echo "⏹️  停止服务..."
    
    if [ "$USE_SYSTEMD" = true ]; then
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            echo "停止systemd服务: $SERVICE_NAME"
            sudo systemctl stop "$SERVICE_NAME"
        fi
    else
        # 查找并停止进程
        local pid=$(pgrep -f "$APP_NAME" || true)
        if [ -n "$pid" ]; then
            echo "停止进程: $pid"
            sudo kill -TERM "$pid"
            
            # 等待进程优雅退出
            local count=0
            while [ $count -lt 10 ] && kill -0 "$pid" 2>/dev/null; do
                sleep 1
                count=$((count + 1))
            done
            
            # 强制杀死进程
            if kill -0 "$pid" 2>/dev/null; then
                echo "强制停止进程"
                sudo kill -KILL "$pid"
            fi
        fi
    fi
    
    echo -e "${GREEN}✅ 服务已停止${NC}"
}

# 部署新版本
deploy_new_version() {
    echo "📦 部署新版本..."
    
    local binary_path="$1"
    
    if [ ! -f "$binary_path" ]; then
        echo -e "${RED}❌ 二进制文件不存在: $binary_path${NC}"
        return 1
    fi
    
    # 复制二进制文件
    echo "复制二进制文件到: $DEPLOY_PATH"
    sudo -u "$DEPLOY_USER" cp "$binary_path" "$DEPLOY_PATH/$APP_NAME"
    sudo chmod +x "$DEPLOY_PATH/$APP_NAME"
    
    # 复制配置文件（如果存在）
    if [ -f "config.yml" ]; then
        echo "复制配置文件"
        sudo -u "$DEPLOY_USER" cp config.yml "$DEPLOY_PATH/"
    fi
    
    # 创建日志目录
    sudo -u "$DEPLOY_USER" mkdir -p "$DEPLOY_PATH/logs"
    
    # 设置权限
    sudo chown -R "$DEPLOY_USER:$DEPLOY_USER" "$DEPLOY_PATH"
    
    echo -e "${GREEN}✅ 新版本部署完成${NC}"
}

# 创建systemd服务文件
create_systemd_service() {
    if [ "$USE_SYSTEMD" != true ]; then
        return 0
    fi
    
    echo "📝 创建systemd服务文件..."
    
    local service_file="/etc/systemd/system/$SERVICE_NAME.service"
    
    sudo tee "$service_file" > /dev/null << EOF
[Unit]
Description=PDF合并工具
After=network.target
Wants=network.target

[Service]
Type=simple
User=$DEPLOY_USER
Group=$DEPLOY_USER
WorkingDirectory=$DEPLOY_PATH
ExecStart=$DEPLOY_PATH/$APP_NAME
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=always
RestartSec=5
StartLimitInterval=0

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DEPLOY_PATH

# 环境变量
Environment=LOG_LEVEL=info
Environment=PORT=8080
Environment=ENV=$DEPLOY_ENV

# 日志设置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载systemd配置
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    
    echo -e "${GREEN}✅ systemd服务文件已创建${NC}"
}

# 启动服务
start_service() {
    echo "▶️  启动服务..."
    
    if [ "$USE_SYSTEMD" = true ]; then
        echo "启动systemd服务: $SERVICE_NAME"
        sudo systemctl start "$SERVICE_NAME"
        
        # 检查服务状态
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            echo -e "${GREEN}✅ 服务启动成功${NC}"
        else
            echo -e "${RED}❌ 服务启动失败${NC}"
            sudo systemctl status "$SERVICE_NAME" --no-pager
            return 1
        fi
    else
        echo "以后台方式启动服务"
        sudo -u "$DEPLOY_USER" nohup "$DEPLOY_PATH/$APP_NAME" > "$DEPLOY_PATH/logs/app.log" 2>&1 &
        
        # 等待服务启动
        sleep 3
        
        if pgrep -f "$APP_NAME" > /dev/null; then
            echo -e "${GREEN}✅ 服务启动成功${NC}"
        else
            echo -e "${RED}❌ 服务启动失败${NC}"
            return 1
        fi
    fi
}

# 健康检查
health_check() {
    echo "🏥 执行健康检查..."
    
    local count=0
    local max_attempts=$((HEALTH_CHECK_TIMEOUT / 5))
    
    while [ $count -lt $max_attempts ]; do
        echo "健康检查尝试 $((count + 1))/$max_attempts..."
        
        if curl -f -s "$HEALTH_CHECK_URL" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ 健康检查通过${NC}"
            return 0
        fi
        
        sleep 5
        count=$((count + 1))
    done
    
    echo -e "${RED}❌ 健康检查失败${NC}"
    return 1
}

# 回滚到上一版本
rollback() {
    echo "🔄 回滚到上一版本..."
    
    local backup_path_file="/tmp/pdf-merger-backup-path"
    
    if [ ! -f "$backup_path_file" ]; then
        echo -e "${RED}❌ 未找到备份路径信息${NC}"
        return 1
    fi
    
    local backup_path=$(cat "$backup_path_file")
    
    if [ ! -d "$backup_path" ]; then
        echo -e "${RED}❌ 备份目录不存在: $backup_path${NC}"
        return 1
    fi
    
    # 停止当前服务
    stop_service
    
    # 恢复备份
    echo "恢复备份: $backup_path"
    sudo -u "$DEPLOY_USER" cp -r "$backup_path"/* "$DEPLOY_PATH/"
    
    # 启动服务
    start_service
    
    # 健康检查
    if health_check; then
        echo -e "${GREEN}✅ 回滚成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 回滚后健康检查失败${NC}"
        return 1
    fi
}

# 显示部署状态
show_deployment_status() {
    echo ""
    echo "📊 部署状态"
    echo "==========="
    
    # 服务状态
    if [ "$USE_SYSTEMD" = true ]; then
        echo "服务状态:"
        sudo systemctl status "$SERVICE_NAME" --no-pager -l
    else
        echo "进程状态:"
        pgrep -f "$APP_NAME" | while read pid; do
            ps -p "$pid" -o pid,ppid,cmd --no-headers
        done
    fi
    
    echo ""
    echo "部署信息:"
    echo "  版本: $VERSION"
    echo "  环境: $DEPLOY_ENV"
    echo "  路径: $DEPLOY_PATH"
    echo "  用户: $DEPLOY_USER"
    
    # 显示最近的日志
    echo ""
    echo "最近日志:"
    if [ "$USE_SYSTEMD" = true ]; then
        sudo journalctl -u "$SERVICE_NAME" --no-pager -n 10
    else
        tail -n 10 "$DEPLOY_PATH/logs/app.log" 2>/dev/null || echo "无日志文件"
    fi
}

# 主函数
main() {
    echo "开始部署流程..."
    echo ""
    
    local binary_path=""
    local skip_backup=false
    local skip_health_check=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --binary)
                binary_path="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --env)
                DEPLOY_ENV="$2"
                shift 2
                ;;
            --skip-backup)
                skip_backup=true
                shift
                ;;
            --skip-health-check)
                skip_health_check=true
                shift
                ;;
            --rollback)
                rollback
                exit $?
                ;;
            --status)
                show_deployment_status
                exit 0
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --binary <路径>        指定二进制文件路径"
                echo "  --version <版本>       设置版本号"
                echo "  --env <环境>           设置部署环境"
                echo "  --skip-backup          跳过备份"
                echo "  --skip-health-check    跳过健康检查"
                echo "  --rollback             回滚到上一版本"
                echo "  --status               显示部署状态"
                echo "  --help                 显示帮助信息"
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 检查必需参数
    if [ -z "$binary_path" ]; then
        echo -e "${RED}❌ 请指定二进制文件路径 --binary${NC}"
        exit 1
    fi
    
    # 执行部署流程
    check_deploy_environment
    
    if [ "$skip_backup" != true ]; then
        backup_current_version
    fi
    
    stop_service
    deploy_new_version "$binary_path"
    create_systemd_service
    start_service
    
    if [ "$skip_health_check" != true ]; then
        if ! health_check; then
            if [ "$ROLLBACK_ON_FAILURE" = true ]; then
                echo "健康检查失败，开始回滚..."
                rollback
                exit 1
            else
                echo -e "${RED}❌ 部署失败，请手动检查${NC}"
                exit 1
            fi
        fi
    fi
    
    show_deployment_status
    
    echo ""
    echo -e "${GREEN}🎉 部署完成！${NC}"
    echo "版本: $VERSION"
    echo "环境: $DEPLOY_ENV"
}

# 运行主函数
main "$@"
