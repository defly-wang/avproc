# 项目开发日志

## 项目概述

- **项目名称**：AVProc - 音视频处理工具
- **项目类型**：桌面应用程序 (GUI)
- **技术栈**：Go + Fyne + FFmpeg + mpv
- **开发方式**：由 OpenCode AI 编程助手自动生成

---

## 开发记录

### 2026-03-01

#### Windows 兼容性和打包
- 添加 Windows 平台支持
- 创建 `ffmpeg/player_windows.go` 处理 Windows 信号兼容性问题
- 添加 `build.bat` 打包脚本，自动包含 FFmpeg 二进制文件

#### 用户体验优化
- 为所有文件打开对话框添加文件过滤器
- 过滤器显示：媒体文件、视频文件、音频文件
- 支持格式：mp4, avi, mkv, mov, flv, wmv, webm, mp3, wav, aac, flac, ogg, m4a

### 2026-02-28

#### UI 框架搭建
- 创建主窗口和标签页结构
- 实现预览、转换、剪裁、拼接四个功能标签

#### 预览功能
- 添加文件选择对话框
- 显示媒体文件信息（格式、时长、大小、比特率、视频/音频轨道）
- 实现视频第一帧预览图提取和显示
- 调用 mpv 进行播放

#### 转换功能
- 添加文件打开和保存对话框
- 实现格式选择（mp4, avi, mkv, mov, wmv, flv, webm, mp3, wav, aac, ogg, flac, m4a, wma）
- 实现质量选择（高、中、低）
- 添加转换进度条显示
- 转换完成后启用预览按钮

#### 剪裁功能
- 添加时间滑块选择开始和结束时间
- 显示开始帧和结束帧预览图
- 滑块拖动时实时更新预览图
- 添加剪裁进度显示

#### 拼接功能
- 添加文件列表管理（添加/移除）
- 实现视频文件拼接功能
- 显示拼接进度

### 2026-02-28 (后期优化)

#### 代码重构
- 将 `ui/ui.go` 拆分为多个文件：
  - `main_ui.go` - 主界面
  - `preview.go` - 预览功能
  - `convert.go` - 转换功能
  - `crop.go` - 剪裁功能
  - `merge.go` - 拼接功能
  - `player.go` - 播放功能
  - `utils.go` - 工具函数

- 将 `ffmpeg/ffmpeg.go` 拆分为多个文件：
  - `types.go` - 类型定义
  - `check.go` - FFmpeg 检查
  - `info.go` - 媒体信息获取
  - `convert.go` - 转换功能
  - `crop.go` - 剪裁功能
  - `merge.go` - 拼接功能
  - `frame.go` - 帧提取
  - `player.go` - 播放器

#### 功能改进
- 修复 Fyne 应用需要唯一 ID 的问题
- 添加 fyne.Do/fyne.DoAndWait 确保 UI 更新在主线程执行
- 修复进度条不更新的问题
- 修复无扩展名文件多生成的问题

#### 拼接功能增强
- 文件列表显示视频缩略图作为图标
- 增大缩略图尺寸 (80x45 → 120x68)
- 增大文件列表区域高度

#### 文档
- 编写 README.md
- 添加由 OpenCode 自动生成说明

---

## 问题与解决方案

1. **nil pointer dereference 错误**
   - 原因：fyne-streamer 库初始化问题
   - 解决：回退使用 mpv 播放

2. **Preferences API requires a unique ID**
   - 原因：Fyne 应用需要唯一 ID
   - 解决：使用 `app.NewWithID("com.avproc.app")`

3. **Error in Fyne call thread**
   - 原因：在 goroutine 中直接更新 UI
   - 解决：使用 `fyne.Do()` 和 `fyne.DoAndWait()`

4. **转换时多生成无扩展名文件**
   - 原因：保存对话框返回的路径未处理
   - 解决：在转换前删除可能存在的文件

---

## 后续优化方向

- [ ] 支持更多音视频编码选项
- [ ] 添加批量处理功能
- [ ] 实现视频预览缩放
- [ ] 添加快捷键支持
- [ ] 国际化支持
