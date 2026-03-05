package ui

import (
	"avproc/ffmpeg"
	"bytes"
	"fmt"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewConvertTab(window fyne.Window) fyne.Widget {
	var quality string = "medium"
	var resolution string = "original"
	var progress float64
	var inputPath string
	var outputPath string

	formatSelect := widget.NewSelect([]string{
		"mp4", "avi", "mkv", "mov", "wmv", "flv", "webm",
		"mp3", "wav", "aac", "ogg", "flac", "m4a", "wma",
	}, func(s string) {})
	formatSelect.SetSelected("mp4")

	qualitySelect := widget.NewSelect([]string{"high", "medium", "low"}, func(s string) {
		quality = s
	})
	qualitySelect.SetSelected("medium")

	resolutionSelect := widget.NewSelect([]string{
		"original", "1920:1080", "1280:720", "854:480", "640:360",
	}, func(s string) {
		resolution = s
	})
	resolutionSelect.SetSelected("original")

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.Value = 0

	statusLabel := widget.NewLabel("")
	pathLabel := widget.NewLabel("未选择文件")
	infoLabel := widget.NewLabel("")
	//infoLabel.Wrapping = fyne.TextWrapWord
	loadingLabel := widget.NewLabel("")

	previewImage := canvas.NewImageFromResource(nil)
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(320, 180))

	previewWithLabel := container.NewVBox(
		previewImage,
		loadingLabel,
	)

	infoContainer := container.NewVBox(
		widget.NewLabel("文件信息"),
		infoLabel,
	)

	var convertBtn *widget.Button
	var previewBtn *widget.Button

	openInputBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		filter := storage.NewExtensionFileFilter([]string{".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv", ".webm", ".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a"})
		fd := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			inputPath = closer.URI().Path()
			pathLabel.SetText(inputPath)

			info, err := ffmpeg.GetMediaInfo(inputPath)
			if err == nil {
				infoText := fmt.Sprintf("时长: %.2f秒  大小: %s", info.DurationSec, info.Size)
				if len(info.VideoTracks) > 0 {
					vt := info.VideoTracks[0]
					infoText += fmt.Sprintf("\n视频: %s  分辨率: %dx%d\n帧率: %s  比特率: %s",
						vt.Codec, vt.Width, vt.Height, vt.FrameRate, vt.Bitrate)
				}
				if len(info.AudioTracks) > 0 {
					at := info.AudioTracks[0]
					infoText += fmt.Sprintf("\n音频: %s  采样率: %s  声道: %d",
						at.Codec, at.SampleRate, at.Channels)
				}
				infoLabel.SetText(infoText)
			}

			convertBtn.Enable()

			loadingLabel.SetText("正在生成预览...")
			go func() {
				data, err := ffmpeg.ExtractFrame(inputPath, 1.0)
				if err != nil {
					loadingLabel.SetText("")
					return
				}
				img, _, err := image.Decode(bytes.NewReader(data))
				if err != nil {
					loadingLabel.SetText("")
					return
				}
				rgba := image.NewRGBA(img.Bounds())
				for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
					for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
						rgba.Set(x, y, img.At(x, y))
					}
				}
				previewImage.Image = rgba
				previewImage.Refresh()
				loadingLabel.SetText("")
			}()
		}, window)
		fd.SetFilter(filter)
		fd.Show()
	})

	convertBtn = widget.NewButtonWithIcon("转换", theme.MediaRecordIcon(), func() {
		if inputPath == "" {
			statusLabel.SetText("请先选择输入文件")
			return
		}

		dialog.ShowFileSave(func(closer fyne.URIWriteCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			outputPath = closer.URI().Path()

			format := formatSelect.Selected
			if format != "" && outputPath != "" {
				ext := "." + format
				os.Remove(outputPath)
				if len(outputPath) < len(ext) || outputPath[len(outputPath)-len(ext):] != ext {
					outputPath = outputPath + ext
				}
			}

			statusLabel.SetText("转换中...")
			convertBtn.Disable()

			go func() {
				err := ffmpeg.Convert(inputPath, outputPath, quality, resolution, func(p ffmpeg.Progress) {
					progress = p.Percent
					fyne.DoAndWait(func() {
						progressBar.SetValue(progress)
						statusLabel.SetText(fmt.Sprintf("转换中... %.1f%%", progress))
					})
				})

				fyne.Do(func() {
					if err != nil {
						statusLabel.SetText(fmt.Sprintf("转换失败: %v", err))
					} else {
						statusLabel.SetText("转换完成!")
						progressBar.SetValue(100)
						previewBtn.Enable()
					}
					convertBtn.Enable()
				})
			}()
		}, window)
	})
	convertBtn.Disable()

	previewBtn = widget.NewButtonWithIcon("预览", theme.MediaPlayIcon(), func() {
		if outputPath != "" {
			OpenPlayerWindow(outputPath, nil)
		}
	})
	previewBtn.Disable()

	toolbar := container.NewHBox(
		openInputBtn,
		formatSelect,
		qualitySelect,
		resolutionSelect,
		convertBtn,
		previewBtn,
	)

	content := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewVBox(
			pathLabel,
			widget.NewSeparator(),
			container.NewHBox(
				previewWithLabel,
				infoContainer,
			),
			progressBar,
			statusLabel,
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}
