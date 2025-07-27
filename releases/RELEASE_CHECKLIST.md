# 发布检查清单

## 📋 发布前检查

### ✅ 代码质量检查
- [ ] 所有测试通过 (`go test ./...`)
- [ ] 代码构建成功 (`go build ./cmd/pdfmerger`)
- [ ] 无编译警告或错误
- [ ] 代码格式化 (`go fmt ./...`)
- [ ] 静态分析通过 (`go vet ./...`)

### ✅ 功能测试
- [ ] 基本PDF合并功能正常
- [ ] 加密PDF处理正常
- [ ] 用户界面响应正常
- [ ] 文件拖拽功能正常
- [ ] 进度显示正常
- [ ] 错误处理正常

### ✅ 文档检查
- [ ] README.md 更新完整
- [ ] 用户指南完整
- [ ] 技术文档完整
- [ ] 版本号一致
- [ ] 更新日志完整

### ✅ 构建检查
- [ ] 运行构建脚本 (`./scripts/build_releases.sh`)
- [ ] 生成的可执行文件正常运行
- [ ] 文件大小合理（< 50MB）
- [ ] 校验和文件生成
- [ ] 发布说明生成

## 📦 发布流程

### 1. 准备发布
```bash
# 1. 确保代码最新
git pull origin main

# 2. 运行完整测试
go test ./...

# 3. 构建发布版本
./scripts/build_releases.sh

# 4. 测试构建的可执行文件
./releases/v1.0.0/pdf-merger-macos-intel
```

### 2. 创建GitHub Release

#### 2.1 访问GitHub
- 访问: https://github.com/YOUR_USERNAME/pdf-merger
- 点击 "Releases" 标签
- 点击 "Create a new release"

#### 2.2 填写发布信息
- **Tag version**: `v1.0.0`
- **Release title**: `PDF合并工具 v1.0.0`
- **Description**: 使用 `releases/GITHUB_RELEASE_TEMPLATE.md` 中的模板

#### 2.3 上传文件
拖拽以下文件到Release页面：
- [ ] `releases/v1.0.0/pdf-merger-macos-intel`
- [ ] `releases/v1.0.0/checksums.sha256`
- [ ] `releases/v1.0.0/RELEASE_NOTES.md`

#### 2.4 发布设置
- [ ] ✅ Set as the latest release
- [ ] ✅ Create a discussion for this release (可选)

### 3. 发布后验证

#### 3.1 下载测试
- [ ] 测试下载链接正常工作
- [ ] 验证文件完整性 (`sha256sum -c checksums.sha256`)
- [ ] 测试下载的文件能正常运行

#### 3.2 文档更新
- [ ] 更新README.md中的下载链接
- [ ] 更新版本号引用
- [ ] 检查所有文档链接

#### 3.3 通知用户
- [ ] 在项目主页添加发布公告
- [ ] 更新项目描述
- [ ] 考虑在相关社区分享

## 🔄 多平台发布

### 当前状态
- ✅ macOS (Intel) - 已构建
- ⏳ macOS (Apple Silicon) - 需要M1/M2 Mac构建
- ⏳ Windows 64位 - 需要Windows系统构建
- ⏳ Windows 32位 - 需要Windows系统构建
- ⏳ Linux 64位 - 需要Linux系统构建
- ⏳ Linux ARM64 - 需要ARM64 Linux构建

### 多平台构建计划

#### 方案一：手动构建
在不同系统上运行构建脚本：
```bash
# 在对应系统上运行
./scripts/build_releases.sh
```

#### 方案二：GitHub Actions自动构建
创建 `.github/workflows/release.yml`：
```yaml
name: Release
on:
  push:
    tags: ['v*']
jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Build
        run: ./scripts/build_releases.sh
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries-${{ matrix.os }}
          path: releases/v*/pdf-merger-*
```

## 📊 发布后监控

### 下载统计
- [ ] 监控GitHub Release下载数量
- [ ] 收集用户反馈
- [ ] 记录常见问题

### 问题跟踪
- [ ] 监控新的Issues
- [ ] 及时回复用户问题
- [ ] 收集改进建议

### 版本规划
- [ ] 规划下一个版本功能
- [ ] 评估用户需求
- [ ] 制定开发计划

## 🚨 紧急修复流程

如果发现严重问题需要紧急修复：

1. **立即行动**
   - 在Release页面添加警告说明
   - 暂时隐藏有问题的版本

2. **快速修复**
   - 创建hotfix分支
   - 修复问题
   - 快速测试

3. **紧急发布**
   - 构建新版本 (如 v1.0.1)
   - 创建新的Release
   - 更新文档说明

## 📝 发布记录

### v1.0.0 (2025-07-27)
- [x] 初始发布版本
- [x] 基本PDF合并功能
- [x] macOS Intel版本构建完成
- [ ] 其他平台版本待构建

---

**最后更新**: 2025-07-27  
**负责人**: 开发团队  
**状态**: 准备发布
