#!/bin/bash

# PDF合并工具Docker构建脚本

set -e

echo "🐳 PDF合并工具 - Docker构建"
echo "==========================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
IMAGE_NAME="pdf-merger"
VERSION=${VERSION:-"latest"}
REGISTRY=${REGISTRY:-""}
DOCKERFILE="Dockerfile"
BUILD_CONTEXT="."

# 检查Docker环境
check_docker() {
    echo "🔍 检查Docker环境..."
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker未安装${NC}"
        echo "请访问 https://docs.docker.com/get-docker/ 安装Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo -e "${RED}❌ Docker服务未运行${NC}"
        echo "请启动Docker服务"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker环境正常${NC}"
    echo "Docker版本: $(docker --version)"
}

# 创建Dockerfile
create_dockerfile() {
    if [ ! -f "$DOCKERFILE" ]; then
        echo "📝 创建Dockerfile..."
        
        cat > "$DOCKERFILE" << 'EOF'
# 多阶段构建Dockerfile for PDF合并工具

# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pdf-merger ./cmd/pdfmerger

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/pdf-merger .

# 创建必要的目录
RUN mkdir -p /app/temp /app/output && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口（如果有Web界面）
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./pdf-merger --health-check || exit 1

# 启动命令
CMD ["./pdf-merger"]
EOF
        
        echo -e "${GREEN}✅ Dockerfile已创建${NC}"
    else
        echo "📄 使用现有Dockerfile"
    fi
}

# 创建.dockerignore文件
create_dockerignore() {
    if [ ! -f ".dockerignore" ]; then
        echo "📝 创建.dockerignore..."
        
        cat > ".dockerignore" << 'EOF'
# 构建文件
build/
dist/
tmp/
*.exe
pdf-merger
pdf-merger-*

# 测试文件
coverage.out
coverage.html
*.test

# 开发文件
.git/
.gitignore
.air.toml
.vscode/
.idea/

# 文档
docs/
README.md
LICENSE
*.md

# 日志文件
*.log

# 临时文件
.DS_Store
Thumbs.db
*~
.#*
EOF
        
        echo -e "${GREEN}✅ .dockerignore已创建${NC}"
    fi
}

# 构建Docker镜像
build_image() {
    local tag="$1"
    
    echo -e "${BLUE}🔨 构建Docker镜像: $tag${NC}"
    
    # 构建参数
    local build_args=(
        --build-arg "VERSION=$VERSION"
        --build-arg "BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        --tag "$tag"
        --file "$DOCKERFILE"
    )
    
    # 如果有registry，添加额外标签
    if [ -n "$REGISTRY" ]; then
        build_args+=(--tag "$REGISTRY/$tag")
    fi
    
    # 执行构建
    if docker build "${build_args[@]}" "$BUILD_CONTEXT"; then
        echo -e "${GREEN}✅ 镜像构建成功: $tag${NC}"
        
        # 显示镜像信息
        echo "镜像信息:"
        docker images "$IMAGE_NAME:$VERSION" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
        
        return 0
    else
        echo -e "${RED}❌ 镜像构建失败${NC}"
        return 1
    fi
}

# 测试Docker镜像
test_image() {
    local tag="$1"
    
    echo -e "${BLUE}🧪 测试Docker镜像: $tag${NC}"
    
    # 运行容器进行测试
    local container_name="pdf-merger-test-$$"
    
    echo "启动测试容器..."
    if docker run --name "$container_name" --rm -d "$tag" --version; then
        sleep 2
        
        # 检查容器状态
        if docker ps -q -f name="$container_name" | grep -q .; then
            echo -e "${GREEN}✅ 容器启动成功${NC}"
            
            # 停止容器
            docker stop "$container_name" &> /dev/null || true
            
            return 0
        else
            echo -e "${RED}❌ 容器启动失败${NC}"
            docker logs "$container_name" 2>/dev/null || true
            return 1
        fi
    else
        echo -e "${RED}❌ 容器运行失败${NC}"
        return 1
    fi
}

# 推送镜像到registry
push_image() {
    local tag="$1"
    
    if [ -z "$REGISTRY" ]; then
        echo -e "${YELLOW}⚠️  未设置REGISTRY，跳过推送${NC}"
        return 0
    fi
    
    echo -e "${BLUE}📤 推送镜像到registry: $REGISTRY/$tag${NC}"
    
    # 登录registry（如果需要）
    if [ -n "$REGISTRY_USERNAME" ] && [ -n "$REGISTRY_PASSWORD" ]; then
        echo "登录registry..."
        echo "$REGISTRY_PASSWORD" | docker login "$REGISTRY" -u "$REGISTRY_USERNAME" --password-stdin
    fi
    
    # 推送镜像
    if docker push "$REGISTRY/$tag"; then
        echo -e "${GREEN}✅ 镜像推送成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 镜像推送失败${NC}"
        return 1
    fi
}

# 清理构建缓存
cleanup_build_cache() {
    echo "🧹 清理Docker构建缓存..."
    
    # 清理悬空镜像
    docker image prune -f
    
    # 清理构建缓存
    docker builder prune -f
    
    echo -e "${GREEN}✅ 构建缓存已清理${NC}"
}

# 生成docker-compose.yml
generate_docker_compose() {
    if [ ! -f "docker-compose.yml" ]; then
        echo "📝 生成docker-compose.yml..."
        
        cat > "docker-compose.yml" << EOF
version: '3.8'

services:
  pdf-merger:
    image: ${REGISTRY:+$REGISTRY/}$IMAGE_NAME:$VERSION
    container_name: pdf-merger
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./input:/app/input:ro
      - ./output:/app/output
      - ./temp:/app/temp
    environment:
      - LOG_LEVEL=info
      - MAX_FILE_SIZE=100MB
      - TEMP_DIR=/app/temp
    healthcheck:
      test: ["CMD", "./pdf-merger", "--health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=100m

  # 可选：添加监控服务
  # prometheus:
  #   image: prom/prometheus:latest
  #   ports:
  #     - "9090:9090"
  #   volumes:
  #     - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro

networks:
  default:
    name: pdf-merger-network
EOF
        
        echo -e "${GREEN}✅ docker-compose.yml已生成${NC}"
    fi
}

# 显示使用说明
show_usage() {
    echo ""
    echo "📚 Docker使用说明"
    echo "================="
    echo ""
    echo "构建镜像:"
    echo "  docker build -t $IMAGE_NAME:$VERSION ."
    echo ""
    echo "运行容器:"
    echo "  docker run -d --name pdf-merger -p 8080:8080 $IMAGE_NAME:$VERSION"
    echo ""
    echo "使用docker-compose:"
    echo "  docker-compose up -d"
    echo ""
    echo "查看日志:"
    echo "  docker logs pdf-merger"
    echo ""
    echo "进入容器:"
    echo "  docker exec -it pdf-merger sh"
    echo ""
    echo "停止容器:"
    echo "  docker stop pdf-merger"
    echo ""
    echo "删除容器:"
    echo "  docker rm pdf-merger"
}

# 主函数
main() {
    echo "开始Docker构建流程..."
    echo ""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --registry)
                REGISTRY="$2"
                shift 2
                ;;
            --push)
                PUSH_IMAGE=true
                shift
                ;;
            --no-test)
                SKIP_TEST=true
                shift
                ;;
            --cleanup)
                cleanup_build_cache
                exit 0
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --version <版本>     设置镜像版本"
                echo "  --registry <地址>    设置registry地址"
                echo "  --push              构建后推送到registry"
                echo "  --no-test           跳过镜像测试"
                echo "  --cleanup           清理构建缓存"
                echo "  --help              显示帮助信息"
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 执行构建流程
    check_docker
    create_dockerfile
    create_dockerignore
    
    local image_tag="$IMAGE_NAME:$VERSION"
    
    if build_image "$image_tag"; then
        if [ "$SKIP_TEST" != true ]; then
            test_image "$image_tag"
        fi
        
        if [ "$PUSH_IMAGE" = true ]; then
            push_image "$image_tag"
        fi
        
        generate_docker_compose
        show_usage
        
        echo ""
        echo -e "${GREEN}🎊 Docker构建完成！${NC}"
        echo "镜像: $image_tag"
    else
        echo -e "${RED}❌ Docker构建失败${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
