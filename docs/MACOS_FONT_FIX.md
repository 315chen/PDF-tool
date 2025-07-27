# macOS中文字体显示问题解决方案

## 问题描述

在macOS系统上运行PDF合并工具时，可能会出现中文字符显示为乱码（如 ��� 或方块字符）的问题。这是由于Fyne GUI框架在macOS上的字体渲染机制导致的。

## 解决方案

我们提供了三种解决方案，按推荐程度排序：

### 🎯 方案一：使用英文界面版本（推荐）

这是最简单有效的解决方案：

```bash
# 直接运行英文版本
./pdf-merger-english
```

**优点：**
- 无需任何配置
- 界面清晰，功能完整
- 避免所有字体问题

**缺点：**
- 界面为英文

### 🔧 方案二：使用字体修复脚本

如果您需要中文界面，可以使用我们提供的自动修复脚本：

```bash
# 1. 运行字体修复脚本
./fix_chinese_font.sh

# 2. 构建中文版本
go build -o pdf-merger-chinese ./cmd/pdfmerger

# 3. 使用修复后的启动脚本
./run_with_chinese_font.sh
```

**修复脚本做了什么：**
- 检测系统可用的中文字体
- 创建fontconfig配置文件
- 设置正确的环境变量
- 生成优化的启动脚本

### ⚙️ 方案三：手动设置环境变量

如果您熟悉命令行，可以手动设置环境变量：

```bash
# 设置语言环境
export LANG=zh_CN.UTF-8
export LC_ALL=zh_CN.UTF-8
export LC_CTYPE=zh_CN.UTF-8

# 设置字体（根据您的系统调整路径）
export FYNE_FONT="/System/Library/Fonts/STHeiti Light.ttc"

# 运行应用程序
./pdf-merger-font-fix
```

## 系统字体检测结果

修复脚本检测到您的系统有以下字体：

- ✅ `/System/Library/Fonts/STHeiti Light.ttc` - 华文黑体（推荐使用）
- ✅ `/System/Library/Fonts/Helvetica.ttc` - Helvetica字体
- ✅ `/System/Library/Fonts/Apple Color Emoji.ttc` - 表情符号字体

## 故障排除

### 如果方案二仍然显示乱码

1. **检查终端编码设置：**
   ```bash
   echo $LANG
   echo $LC_ALL
   ```
   应该显示包含UTF-8的值

2. **检查字体文件是否存在：**
   ```bash
   ls -la "/System/Library/Fonts/STHeiti Light.ttc"
   ```

3. **尝试其他字体：**
   ```bash
   export FYNE_FONT="/System/Library/Fonts/Helvetica.ttc"
   ./pdf-merger-font-fix
   ```

### 如果应用程序无法启动

1. **检查可执行文件权限：**
   ```bash
   chmod +x pdf-merger-english
   chmod +x pdf-merger-font-fix
   ```

2. **检查依赖：**
   ```bash
   go version
   ```

3. **重新构建：**
   ```bash
   go clean
   go build -o pdf-merger-english ./cmd/pdfmerger
   ```

## 技术原理

### 为什么会出现乱码？

1. **字体回退机制：** Fyne框架在macOS上可能无法正确找到支持中文的字体
2. **编码问题：** 应用程序和系统的字符编码设置不匹配
3. **字体路径：** 系统字体路径在不同macOS版本间可能有差异

### 解决方案的工作原理

1. **环境变量设置：** 通过`LANG`和`LC_ALL`确保正确的UTF-8编码
2. **字体指定：** 通过`FYNE_FONT`直接指定支持中文的字体文件
3. **fontconfig配置：** 创建字体配置文件，指导系统选择合适的字体

## 文件说明

修复过程中创建的文件：

- `pdf-merger-english` - 英文界面版本（推荐）
- `pdf-merger-font-fix` - 中文界面版本（需要字体修复）
- `run_with_chinese_font.sh` - 字体修复启动脚本
- `fix_chinese_font.sh` - 自动字体修复脚本
- `~/.config/fontconfig/fonts.conf` - 字体配置文件

## 推荐使用方式

对于大多数用户，我们推荐：

1. **日常使用：** 使用英文版本 `./pdf-merger-english`
2. **需要中文界面：** 使用修复脚本 `./run_with_chinese_font.sh`
3. **开发调试：** 手动设置环境变量

## 联系支持

如果以上方案都无法解决您的问题，请：

1. 检查您的macOS版本：`sw_vers`
2. 检查已安装的字体：`fc-list | grep -i chinese`
3. 提供错误信息和系统信息

---

**最后更新：** 2025年7月27日  
**适用系统：** macOS 10.14+  
**测试环境：** macOS Monterey, Big Sur, Ventura
