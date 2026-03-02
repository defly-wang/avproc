# AVProc - 音视频处理工具

> 本项目由 **OpenCode** AI 编程助手自动生成

一款基于 FFmpeg 的桌面音视频处理工具，支持预览、转换、剪裁、拼接等功能。

## 功能特性

### 1. 预览
- 打开音视频文件，查看详细信息（格式、时长、大小、比特率、视频/音频轨道信息）
- 显示视频第一帧预览图
- 调用 mpv 播放音视频

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

## 环境依赖

### Linux
- FFmpeg
- mpv（用于播放）
- GTK3 库

### Windows

- FFmpeg

> 注意：Windows 版本已内置播放器，不再依赖 mpv

## 编译方法

### Linux

```bash
# 安装依赖（Ubuntu/Debian）
sudo apt-get install libglib2.0-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev

# 克隆项目
git clone <repository-url>
cd avproc

# 编译
go build ./...

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
4. 点击"播放"按钮使用 mpv 播放

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

- **GUI 框架**：Fyne (Go)
- **音视频处理**：FFmpeg
- **播放器**：内置播放器 (Windows) / mpv (Linux)

## 许可证

MIT License
