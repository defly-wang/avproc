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

func NewCropTab(window fyne.Window) fyne.Widget {
	var progress float64
	var inputPath string
	var outputPath string
	var duration float64

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.Value = 0

	statusLabel := widget.NewLabel("")
	pathLabel := widget.NewLabel("未选择文件")

	previewImageStart := canvas.NewImageFromResource(nil)
	previewImageStart.FillMode = canvas.ImageFillContain
	previewImageStart.SetMinSize(fyne.NewSize(240, 135))

	previewImageEnd := canvas.NewImageFromResource(nil)
	previewImageEnd.FillMode = canvas.ImageFillContain
	previewImageEnd.SetMinSize(fyne.NewSize(240, 135))

	loadingLabel := widget.NewLabel("")

	minSlider := widget.NewSlider(0, 100)
	minSlider.Step = 1
	maxSlider := widget.NewSlider(0, 100)
	maxSlider.Step = 1
	maxSlider.Value = 100

	timeLabel := widget.NewLabel("开始: 00:00  结束: 00:00")

	extractFrame := func(path string, timestamp float64, imgView *canvas.Image) {
		go func() {
			data, err := ffmpeg.ExtractFrame(path, timestamp)
			if err != nil {
				return
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				return
			}
			rgba := image.NewRGBA(img.Bounds())
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					rgba.Set(x, y, img.At(x, y))
				}
			}
			fyne.Do(func() {
				imgView.Image = rgba
				imgView.Refresh()
			})
		}()
	}

	minSlider.OnChanged = func(value float64) {
		if value > maxSlider.Value {
			maxSlider.SetValue(value)
		}
		UpdateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 {
			extractFrame(inputPath, value, previewImageStart)
		}
	}

	maxSlider.OnChanged = func(value float64) {
		if value < minSlider.Value {
			minSlider.SetValue(value)
		}
		UpdateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 {
			extractFrame(inputPath, value, previewImageEnd)
		}
	}

	var cropBtn *widget.Button
	var previewBtn *widget.Button

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
			cropBtn.Enable()

			minSlider.Max = duration
			maxSlider.Max = duration
			minSlider.SetValue(0)
			maxSlider.SetValue(duration)
			UpdateTimeLabel(0, duration, duration, timeLabel)

			loadingLabel.SetText("正在生成预览...")
			extractFrame(inputPath, 0.5, previewImageStart)
			extractFrame(inputPath, duration-1.0, previewImageEnd)

			fyne.Do(func() {
				loadingLabel.SetText("")
			})
		}, window)
		fd.SetFilter(filter)
		fd.Show()
	})

	cropBtn = widget.NewButtonWithIcon("剪裁", theme.MediaRecordIcon(), func() {
		if inputPath == "" {
			statusLabel.SetText("请先选择输入文件")
			return
		}

		startTime := FormatTimeSec(minSlider.Value)
		endTime := FormatTimeSec(maxSlider.Value)

		dialog.ShowFileSave(func(closer fyne.URIWriteCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			outputPath = closer.URI().Path()

			ext := ".mp4"
			os.Remove(outputPath)
			if len(outputPath) >= 4 && outputPath[len(outputPath)-4:] == ".mp4" {
				ext = ""
			} else if len(outputPath) >= 4 && outputPath[len(outputPath)-4:] != ".mp4" {
				ext = ".mp4"
			}
			if ext != "" {
				outputPath = outputPath + ext
			}

			statusLabel.SetText("剪裁中...")
			cropBtn.Disable()

			go func() {
				err := ffmpeg.Crop(inputPath, outputPath, startTime, endTime, func(p ffmpeg.Progress) {
					progress = p.Percent
					fyne.DoAndWait(func() {
						progressBar.SetValue(progress)
						statusLabel.SetText(fmt.Sprintf("剪裁中... %.1f%%", progress))
					})
				})

				fyne.Do(func() {
					if err != nil {
						statusLabel.SetText(fmt.Sprintf("剪裁失败: %v", err))
					} else {
						statusLabel.SetText("剪裁完成!")
						progressBar.SetValue(100)
						previewBtn.Enable()
					}
					cropBtn.Enable()
				})
			}()
		}, window)
	})
	cropBtn.Disable()

	previewBtn = widget.NewButtonWithIcon("预览", theme.MediaPlayIcon(), func() {
		if outputPath != "" {
			OpenPlayerWindow(outputPath, nil)
		}
	})
	previewBtn.Disable()

	toolbar := container.NewHBox(
		openInputBtn,
		cropBtn,
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
			timeLabel,
			minSlider,
			maxSlider,
			widget.NewSeparator(),
			container.NewHBox(
				container.NewVBox(
					widget.NewLabel("开始"),
					previewImageStart,
				),
				layout.NewSpacer(),
				container.NewVBox(
					widget.NewLabel("结束"),
					previewImageEnd,
				),
			),
			loadingLabel,
			progressBar,
			statusLabel,
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}
