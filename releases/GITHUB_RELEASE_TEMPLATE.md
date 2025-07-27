# GitHub Release 创建模板

## 发布步骤

### 1. 准备发布文件

确保已运行构建脚本：
```bash
./scripts/build_releases.sh
```

### 2. 创建GitHub Release

1. 访问GitHub仓库页面
2. 点击 "Releases" 标签
3. 点击 "Create a new release"
4. 填写以下信息：

**Tag version**: `v1.0.0`
**Release title**: `PDF合并工具 v1.0.0`
**Description**: 使用下面的模板

### 3. Release Description 模板

```markdown
# PDF合并工具 v1.0.0 🚀

一个功能强大、易于使用的PDF文件合并工具，支持跨平台使用。

## 📦 下载

选择适合您操作系统的版本：

### macOS
- **Intel Mac**: [pdf-merger-macos-intel](https://github.com/YOUR_USERNAME/pdf-merger/releases/download/v1.0.0/pdf-merger-macos-intel)
- **Apple Silicon (M1/M2)**: 请使用Intel版本（通过Rosetta运行）

### Windows
- **64位系统**: [pdf-merger-windows-64bit.exe](https://github.com/YOUR_USERNAME/pdf-merger/releases/download/v1.0.0/pdf-merger-windows-64bit.exe)

### Linux
- **64位系统**: [pdf-merger-linux-64bit](https://github.com/YOUR_USERNAME/pdf-merger/releases/download/v1.0.0/pdf-merger-linux-64bit)

## 🚀 快速开始

### macOS
```bash
# 下载后添加执行权限
chmod +x pdf-merger-macos-intel
./pdf-merger-macos-intel
```

### Windows
直接双击 `.exe` 文件运行

### Linux
```bash
# 下载后添加执行权限
chmod +x pdf-merger-linux-64bit
./pdf-merger-linux-64bit
```

## ✨ 主要功能

- 📄 **PDF文件合并** - 支持多个PDF文件合并为单个文件
- 🔐 **加密文件处理** - 自动处理密码保护的PDF文件
- 🎨 **现代化界面** - 基于Fyne的跨平台GUI
- 📊 **实时进度** - 详细的合并进度和状态显示
- 🔄 **拖拽支持** - 支持文件拖拽添加和排序
- ⚡ **高性能** - 流式处理，支持大文件合并
- 🛡️ **错误恢复** - 完善的错误处理和恢复机制

## 🔧 系统要求

- **内存**: 至少 512MB 可用内存
- **磁盘空间**: 至少 100MB 可用空间
- **操作系统**:
  - macOS 10.14 或更高版本
  - Windows 10 或更高版本
  - Ubuntu 18.04 或更高版本（Linux）

## 📚 文档

- [用户使用指南](https://github.com/YOUR_USERNAME/pdf-merger/blob/main/docs/USER_GUIDE.md)
- [技术开发指南](https://github.com/YOUR_USERNAME/pdf-merger/blob/main/docs/TECHNICAL_GUIDE.md)
- [macOS字体修复指南](https://github.com/YOUR_USERNAME/pdf-merger/blob/main/docs/MACOS_FONT_FIX.md)

## 🐛 已知问题

- **macOS中文字体**: 如果遇到中文显示问题，请参考 [macOS字体修复指南](https://github.com/YOUR_USERNAME/pdf-merger/blob/main/docs/MACOS_FONT_FIX.md)

## 🔒 文件校验

下载后可以验证文件完整性：

```bash
# 下载校验和文件
curl -L -o checksums.sha256 https://github.com/YOUR_USERNAME/pdf-merger/releases/download/v1.0.0/checksums.sha256

# 验证文件
sha256sum -c checksums.sha256
```

## 📞 技术支持

如有问题，请：
1. 查看 [用户使用指南](https://github.com/YOUR_USERNAME/pdf-merger/blob/main/docs/USER_GUIDE.md)
2. 搜索已有的 [Issues](https://github.com/YOUR_USERNAME/pdf-merger/issues)
3. 创建新的 [Issue](https://github.com/YOUR_USERNAME/pdf-merger/issues/new)

---

**构建信息**
- 版本: v1.0.0
- 构建时间: 2025-07-27T02:06:14Z
- 测试状态: ✅ 通过
```

### 4. 上传文件

将以下文件拖拽到Release页面：

- `releases/v1.0.0/pdf-merger-macos-intel`
- `releases/v1.0.0/checksums.sha256`
- `releases/v1.0.0/RELEASE_NOTES.md`

### 5. 发布设置

- ✅ Set as the latest release
- ✅ Create a discussion for this release (可选)

### 6. 发布后更新

1. 更新README.md中的下载链接
2. 更新文档中的版本信息
3. 通知用户新版本发布

## 注意事项

1. **替换占位符**: 将 `YOUR_USERNAME` 替换为实际的GitHub用户名
2. **测试下载**: 发布后测试下载链接是否正常工作
3. **文档更新**: 确保所有文档链接指向正确的版本
4. **多平台构建**: 如需其他平台版本，需要在对应系统上构建

## 自动化发布

可以考虑使用GitHub Actions自动化发布流程：

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Build
        run: ./scripts/build_releases.sh
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: releases/v*/pdf-merger-*
```
