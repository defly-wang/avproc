package ui

import (
	"avproc/ffmpeg"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewScreenshotTab(window fyne.Window) fyne.Widget {
	var inputPath string
	var duration float64
	var outputDir string

	pathLabel := widget.NewLabel("未选择文件")
	statusLabel := widget.NewLabel("")
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.Value = 0

	countEntry := widget.NewEntry()
	countEntry.SetText("10")

	countLabel := widget.NewLabel("截图数量:")
	countLabel.Alignment = fyne.TextAlignCenter

	var screenshotBtn *widget.Button

	screenshotBtn = widget.NewButtonWithIcon("开始截图", theme.MediaRecordIcon(), func() {
		if inputPath == "" {
			statusLabel.SetText("请先选择输入文件")
			return
		}

		count := 10
		if countEntry.Text != "" {
			fmt.Sscanf(countEntry.Text, "%d", &count)
		}
		if count <= 0 || count > 100 {
			statusLabel.SetText("截图数量必须在 1-100 之间")
			return
		}

		dialog.ShowFolderOpen(func(closer fyne.ListableURI, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			outputDir = closer.Path()

			statusLabel.SetText("正在截图...")
			screenshotBtn.Disable()
			progressBar.SetValue(0)

			go func() {
				results, err := ffmpeg.ExtractScreenshots(inputPath, outputDir, count, func(current, total int) {
					percent := float64(current) / float64(total) * 100
					fyne.Do(func() {
						progressBar.SetValue(percent)
						statusLabel.SetText(fmt.Sprintf("正在截图... %d/%d", current, total))
					})
				})

				fyne.Do(func() {
					if err != nil {
						statusLabel.SetText(fmt.Sprintf("截图失败: %v", err))
					} else {
						statusLabel.SetText(fmt.Sprintf("截图完成! 共 %d 张图片", len(results)))
						progressBar.SetValue(100)
					}
					screenshotBtn.Enable()
				})
			}()
		}, window)
	})

	openInputBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		filter := storage.NewExtensionFileFilter([]string{".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv", ".webm"})
		fd := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			inputPath = closer.URI().Path()

			info, err := ffmpeg.GetMediaInfo(inputPath)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}

			duration = info.DurationSec
			pathLabel.SetText(inputPath)

			count := int(duration / 10)
			if count < 3 {
				count = 3
			}
			if count > 50 {
				count = 50
			}
			countEntry.SetText(fmt.Sprintf("%d", count))

			statusLabel.SetText(fmt.Sprintf("视频时长: %s", FormatDuration(duration)))
			screenshotBtn.Enable()
		}, window)
		fd.SetFilter(filter)
		fd.Show()
	})

	toolbar := container.NewHBox(
		openInputBtn,
		screenshotBtn,
	)

	countRow := container.NewHBox(
		countLabel,
		countEntry,
		widget.NewLabel("张"),
	)

	content := container.NewVBox(
		toolbar,
		widget.NewSeparator(),
		pathLabel,
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			countRow,
			layout.NewSpacer(),
		),
		widget.NewSeparator(),
		progressBar,
		statusLabel,
		layout.NewSpacer(),
	)

	screenshotBtn.Disable()

	return container.NewScroll(content)
}

func FormatDuration(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func GetFilenameWithoutExt(path string) string {
	filename := filepath.Base(path)
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}
	return filename
}
