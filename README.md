# AVProc - 音视频处理工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux-blue.svg)](https://github.com/defly-wang/avproc)

> 本项目由 **OpenCode** AI 编程助手自动生成

一款基于 FFmpeg 的桌面音视频处理工具，支持预览、转换、剪裁、拼接等功能。采用纯 Go 开发，跨平台支持 Windows 和 Linux。

## 功能特性

### 1. 预览
- 打开音视频文件，查看详细信息（格式、时长、大小、比特率、视频/音频轨道信息）
- 显示视频第一帧预览图
- 内置播放器（Windows）/ mpv（Linux）

### 2. 转换
- 支持多种视频格式：mp4, avi, mkv, mov, wmv, flv, webm
- 支持多种音频格式：mp3, wav, aac, ogg, flac, m4a, wma
- 三种质量选项：高、中、低
- 显示输入视频预览图
- 转换完成后可直接预览

### 3. 剪裁
- 通过滑块选择开始和结束时间
- 实时显示开始/结束帧预览图
- 剪裁完成后可直接预览

### 4. 拼接
- 添加多个视频文件
- 列表中显示视频缩略图
- 拼接完成后可直接预览

## 系统要求

### Windows
- Windows 10 或更高版本
- 已发布版本包含 FFmpeg，无需额外安装

### Linux
- FFmpeg
- mpv（用于播放）
- GTK3 库

## 下载使用

### 从 Release 下载（推荐）
访问 [ Releases ](https://github.com/defly-wang/avproc/releases) 页面下载最新版本。

Windows 用户下载 `dist` 目录下的文件即可运行：
- `avproc.exe` - 主程序
- `ffmpeg.exe` - FFmpeg
- `ffplay.exe` - FFplay
- `ffprobe.exe` - FFprobe

## 编译方法

### Linux

```bash
# 安装依赖（Ubuntu/Debian）
sudo apt-get install libglib2.0-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev

# 克隆项目
git clone https://github.com/defly-wang/avproc.git
cd avproc

# 编译
go build -o avproc .

# 运行
./avproc
```

### Windows

```bash
# 安装 MinGW (用于 CGO 编译)
# 下载地址: https://github.com/mstorsjo/llvm-mingw/releases

# 设置环境变量
set PATH=C:\path\to\llvm-mingw\bin;%PATH%
set CGO_ENABLED=1

# 编译
go build -o avproc.exe .

# 或者使用打包脚本
build.bat
```

打包脚本会自动将 FFmpeg 二进制文件复制到 dist 目录。

## 使用方法

### 启动应用

```bash
# Linux
./avproc

# Windows
avproc.exe
```

### 功能操作

#### 预览功能
1. 点击"打开"按钮选择音视频文件
2. 查看文件信息（格式、时长、大小等）
3. 查看视频预览图
4. 点击"播放"按钮播放

#### 转换功能
1. 点击"打开"按钮选择输入文件
2. 选择输出格式和质量
3. 点击"转换"按钮
4. 选择保存路径
5. 等待转换完成
6. 点击"预览"播放转换后的文件

#### 剪裁功能
1. 点击"打开"按钮选择视频文件
2. 拖动滑块设置开始和结束时间
3. 观察预览图确认剪裁范围
4. 点击"剪裁"按钮
5. 选择保存路径
6. 等待剪裁完成
7. 点击"预览"播放剪裁后的文件

#### 拼接功能
1. 点击"添加"按钮添加多个视频文件
2. 可以在列表中查看每个文件的缩略图
3. 选择要移除的文件，点击"移除"
4. 点击"拼接"按钮
5. 选择保存路径
6. 等待拼接完成
7. 点击"预览"播放拼接后的文件

## 技术栈

- **GUI 框架**：[Fyne](https://fyne.io/) (Go)
- **音视频处理**：[FFmpeg](https://ffmpeg.org/)
- **播放器**：内置播放器 (Windows) / [mpv](https://mpv.io/) (Linux)

## 项目结构

```
avproc/
├── ffmpeg/          # FFmpeg 封装库
│   ├── check.go     # FFmpeg 检查
│   ├── convert.go   # 格式转换
│   ├── crop.go     # 视频剪裁
│   ├── merge.go    # 视频拼接
│   ├── frame.go    # 帧提取
│   ├── info.go     # 媒体信息
│   ├── player.go   # Linux 播放器
│   └── player_windows.go  # Windows 播放器
├── ui/              # 界面代码
│   ├── main_ui.go  # 主界面
│   ├── preview.go  # 预览功能
│   ├── convert.go  # 转换功能
│   ├── crop.go    # 剪裁功能
│   ├── merge.go    # 拼接功能
│   └── player.go   # 播放功能
├── main.go          # 程序入口
├── build.bat        # Windows 打包脚本
└── README.md       # 说明文档
```

## 许可证

MIT License - 查看 [LICENSE](LICENSE) 了解详情
