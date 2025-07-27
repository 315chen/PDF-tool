# PDF合并工具 - 下载页面

## 📦 最新版本：v1.0.0

### 🚀 快速下载

| 平台 | 文件 | 大小 | 状态 |
|------|------|------|------|
| **macOS (Intel)** | [pdf-merger-macos-intel](releases/v1.0.0/pdf-merger-macos-intel) | 19MB | ✅ 可用 |
| **Windows 64位** | pdf-merger-windows-64bit.exe | - | 🔄 即将发布 |
| **Linux 64位** | pdf-merger-linux-64bit | - | 🔄 即将发布 |

### 📋 文件校验

下载 [checksums.sha256](releases/v1.0.0/checksums.sha256) 验证文件完整性：

```bash
# macOS/Linux
shasum -a 256 -c checksums.sha256

# Windows (PowerShell)
Get-FileHash pdf-merger-*.exe -Algorithm SHA256
```

## 🚀 安装和使用

### macOS
```bash
# 1. 下载文件
curl -L -o pdf-merger-macos-intel https://github.com/YOUR_USERNAME/pdf-merger/raw/main/releases/v1.0.0/pdf-merger-macos-intel

# 2. 添加执行权限
chmod +x pdf-merger-macos-intel

# 3. 运行应用程序
./pdf-merger-macos-intel
```

**注意：** 首次运行时，macOS可能会显示安全警告。请按以下步骤操作：
1. 右键点击文件，选择"打开"
2. 在弹出的对话框中点击"打开"
3. 或者在"系统偏好设置" > "安全性与隐私"中允许运行

### Windows
```cmd
# 1. 下载 pdf-merger-windows-64bit.exe
# 2. 双击运行
```

### Linux
```bash
# 1. 下载文件
wget https://github.com/YOUR_USERNAME/pdf-merger/raw/main/releases/v1.0.0/pdf-merger-linux-64bit

# 2. 添加执行权限
chmod +x pdf-merger-linux-64bit

# 3. 运行应用程序
./pdf-merger-linux-64bit
```

## 🔧 系统要求

### 最低要求
- **内存**: 512MB 可用内存
- **磁盘空间**: 100MB 可用空间
- **网络**: 无需网络连接（离线使用）

### 支持的操作系统
- **macOS**: 10.14 (Mojave) 或更高版本
- **Windows**: Windows 10 或更高版本
- **Linux**: Ubuntu 18.04 或同等版本

## 🎯 功能特性

- ✅ **PDF文件合并** - 将多个PDF文件合并为一个
- ✅ **加密文件支持** - 自动处理密码保护的PDF
- ✅ **拖拽操作** - 支持文件拖拽添加
- ✅ **实时进度** - 显示合并进度和状态
- ✅ **跨平台** - 支持Windows、macOS、Linux
- ✅ **无依赖** - 单文件可执行程序，无需安装

## 🐛 故障排除

### macOS问题

**问题：显示"无法打开，因为无法验证开发者"**
```bash
# 解决方案1：右键打开
右键点击文件 → 选择"打开" → 点击"打开"

# 解决方案2：命令行移除隔离属性
xattr -d com.apple.quarantine pdf-merger-macos-intel
```

**问题：中文字符显示为乱码**
```bash
# 使用字体修复脚本
./fix_chinese_font.sh
./run_with_chinese_font.sh
```

### Windows问题

**问题：Windows Defender报告威胁**
- 这是误报，可以添加到排除列表
- 或者从源码自行编译

**问题：缺少运行时库**
- 下载并安装 Microsoft Visual C++ Redistributable

### Linux问题

**问题：权限被拒绝**
```bash
chmod +x pdf-merger-linux-64bit
```

**问题：缺少GUI库**
```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-glx libxrandr2 libxss1 libxcursor1 libxcomposite1 libasound2 libxi6 libxtst6

# CentOS/RHEL
sudo yum install mesa-libGL libXrandr libXss libXcursor libXcomposite alsa-lib libXi libXtst
```

## 📚 相关文档

- [用户使用指南](docs/USER_GUIDE.md) - 详细使用说明
- [技术开发指南](docs/TECHNICAL_GUIDE.md) - 开发者文档
- [macOS字体修复](docs/MACOS_FONT_FIX.md) - macOS字体问题解决
- [快速开始指南](QUICK_START_MACOS.md) - macOS快速开始

## 🔄 版本历史

### v1.0.0 (2025-07-27)
- 🎉 初始发布版本
- ✅ 基本PDF合并功能
- ✅ 加密文件处理
- ✅ 现代化GUI界面
- ✅ 跨平台支持

## 📞 获取帮助

### 问题报告
如果遇到问题，请：
1. 查看 [故障排除](#-故障排除) 部分
2. 搜索已有的 [Issues](https://github.com/YOUR_USERNAME/pdf-merger/issues)
3. 创建新的 [Issue](https://github.com/YOUR_USERNAME/pdf-merger/issues/new)

### 功能建议
欢迎提出功能建议：
- 创建 [Feature Request](https://github.com/YOUR_USERNAME/pdf-merger/issues/new?template=feature_request.md)
- 参与 [Discussions](https://github.com/YOUR_USERNAME/pdf-merger/discussions)

### 联系方式
- **GitHub**: [项目主页](https://github.com/YOUR_USERNAME/pdf-merger)
- **Issues**: [问题跟踪](https://github.com/YOUR_USERNAME/pdf-merger/issues)
- **Discussions**: [讨论区](https://github.com/YOUR_USERNAME/pdf-merger/discussions)

---

**最后更新**: 2025-07-27  
**当前版本**: v1.0.0  
**下载统计**: [GitHub Releases](https://github.com/YOUR_USERNAME/pdf-merger/releases)
