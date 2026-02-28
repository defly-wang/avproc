package ui

import (
	"avproc/ffmpeg"
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainUI struct {
	window fyne.Window
	tabs   *container.AppTabs
}

func NewMainUI(window fyne.Window) fyne.CanvasObject {
	ui := &MainUI{window: window}

	preview := NewPreviewTab(window)
	convert := NewConvertTab(window)
	crop := NewCropTab(window)
	merge := NewMergeTab(window)

	ui.tabs = container.NewAppTabs(
		container.NewTabItem("预览", preview),
		container.NewTabItem("转换", convert),
		container.NewTabItem("剪裁", crop),
		container.NewTabItem("拼接", merge),
	)

	ui.tabs.SetTabLocation(container.TabLocationTop)

	return ui.tabs
}

func NewPreviewTab(window fyne.Window) fyne.Widget {
	var info *ffmpeg.MediaInfo
	var currentPath string

	pathLabel := widget.NewLabel("未选择文件")
	infoLabel := widget.NewLabel("")
	infoLabel.Wrapping = fyne.TextWrapWord

	previewImage := canvas.NewImageFromResource(nil)
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(320, 180))

	loadingLabel := widget.NewLabel("")

	playBtn := widget.NewButtonWithIcon("播放", theme.MediaPlayIcon(), func() {
		if currentPath == "" {
			infoLabel.SetText("请先选择文件")
			return
		}
		openPlayerWindow(currentPath, info)
	})
	playBtn.Disable()

	selectBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				infoLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			path := closer.URI().Path()
			i, err := ffmpeg.GetMediaInfo(path)
			if err != nil {
				infoLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			info = i
			currentPath = path
			pathLabel.SetText(path)
			displayInfo(info, infoLabel)
			if len(info.VideoTracks) > 0 || len(info.AudioTracks) > 0 {
				playBtn.Enable()
			}

			if len(info.VideoTracks) > 0 {
				loadingLabel.SetText("正在生成预览...")
				go func() {
					data, err := ffmpeg.ExtractFrame(path, 1.0)
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
			}
		}, window)
	})

	toolbar := container.NewHBox(
		selectBtn,
		playBtn,
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
				previewImage,
				container.NewVBox(
					loadingLabel,
					layout.NewSpacer(),
				),
			),
			infoLabel,
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}

func displayInfo(info *ffmpeg.MediaInfo, label *widget.Label) {
	text := fmt.Sprintf(`文件: %s
格式: %s
时长: %s
大小: %s
比特率: %s
`, info.Filename, info.Format, info.Duration, info.Size, info.Bitrate)

	for i, v := range info.VideoTracks {
		text += fmt.Sprintf(`
视频轨道 %d:
  编解码器: %s
  分辨率: %dx%d
  帧率: %s
  比特率: %s
`, i+1, v.Codec, v.Width, v.Height, v.FrameRate, v.Bitrate)
	}

	for i, a := range info.AudioTracks {
		text += fmt.Sprintf(`
音频轨道 %d:
  编解码器: %s
  采样率: %s Hz
  声道数: %d
  比特率: %s
`, i+1, a.Codec, a.SampleRate, a.Channels, a.Bitrate)
	}

	label.SetText(text)
}

func NewConvertTab(window fyne.Window) fyne.Widget {
	var quality string = "medium"
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

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.Value = 0

	statusLabel := widget.NewLabel("")
	pathLabel := widget.NewLabel("未选择文件")

	previewImage := canvas.NewImageFromResource(nil)
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(320, 180))

	loadingLabel := widget.NewLabel("")

	var convertBtn *widget.Button

	openInputBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			inputPath = closer.URI().Path()
			pathLabel.SetText(inputPath)
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
				err := ffmpeg.Convert(inputPath, outputPath, quality, func(p ffmpeg.Progress) {
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
					}
					convertBtn.Enable()
				})
			}()
		}, window)
	})
	convertBtn.Disable()

	toolbar := container.NewHBox(
		openInputBtn,
		formatSelect,
		qualitySelect,
		convertBtn,
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
				previewImage,
				container.NewVBox(
					loadingLabel,
					layout.NewSpacer(),
				),
			),
			progressBar,
			statusLabel,
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}

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
		updateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 {
			extractFrame(inputPath, value, previewImageStart)
		}
	}

	maxSlider.OnChanged = func(value float64) {
		if value < minSlider.Value {
			minSlider.SetValue(value)
		}
		updateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 {
			extractFrame(inputPath, value, previewImageEnd)
		}
	}

	var cropBtn *widget.Button

	openInputBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
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
			updateTimeLabel(0, duration, duration, timeLabel)

			loadingLabel.SetText("正在生成预览...")
			extractFrame(inputPath, 0.5, previewImageStart)
			extractFrame(inputPath, duration-1.0, previewImageEnd)

			fyne.Do(func() {
				loadingLabel.SetText("")
			})
		}, window)
	})

	cropBtn = widget.NewButtonWithIcon("剪裁", theme.MediaRecordIcon(), func() {
		if inputPath == "" {
			statusLabel.SetText("请先选择输入文件")
			return
		}

		startTime := formatTimeSec(minSlider.Value)
		endTime := formatTimeSec(maxSlider.Value)

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
					}
					cropBtn.Enable()
				})
			}()
		}, window)
	})
	cropBtn.Disable()

	toolbar := container.NewHBox(
		openInputBtn,
		cropBtn,
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

func updateTimeLabel(start, end, duration float64, label *widget.Label) {
	startStr := formatTimeSec(start)
	endStr := formatTimeSec(end)
	label.SetText(fmt.Sprintf("开始: %s  结束: %s", startStr, endStr))
}

func formatTimeSec(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", m/60, m%60, s)
}

func NewMergeTab(window fyne.Window) fyne.Widget {
	var files []string
	var selectedIndex int = -1

	list := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(files[id])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("输出文件路径...")

	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("")

	addBtn := widget.NewButton("添加文件", func() {
		dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			files = append(files, closer.URI().Path())
			list.Refresh()
		}, window)
	})

	removeBtn := widget.NewButton("移除选中", func() {
		if selectedIndex >= 0 && selectedIndex < len(files) {
			files = append(files[:selectedIndex], files[selectedIndex+1:]...)
			list.Refresh()
			selectedIndex = -1
		}
	})

	var mergeBtn *widget.Button
	mergeBtn = widget.NewButton("开始拼接", func() {
		if len(files) < 2 {
			statusLabel.SetText("请至少添加2个文件")
			return
		}
		outputPath := pathEntry.Text
		if outputPath == "" {
			statusLabel.SetText("请填写输出路径")
			return
		}

		statusLabel.SetText("拼接中...")
		mergeBtn.Disable()

		go func() {
			err := ffmpeg.Merge(files, outputPath, func(p ffmpeg.Progress) {
				progressBar.SetValue(p.Percent)
				statusLabel.SetText(fmt.Sprintf("拼接中... %.1f%%", p.Percent))
			})

			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("拼接失败: %v", err))
				} else {
					statusLabel.SetText("拼接完成!")
					progressBar.SetValue(100)
				}
				mergeBtn.Enable()
			})
		}()
	})

	buttons := container.NewHBox(addBtn, removeBtn)

	content := container.NewVBox(
		widget.NewLabel("文件列表:"),
		container.NewMax(list),
		buttons,
		widget.NewLabel("输出文件:"),
		pathEntry,
		layout.NewSpacer(),
		mergeBtn,
		progressBar,
		statusLabel,
	)

	return container.NewScroll(content)
}

func openPlayerWindow(path string, info *ffmpeg.MediaInfo) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("播放出错: %v\n", r)
			}
		}()

		cmd := exec.Command("mpv", "--volume=100", "--osd-level=2", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("播放错误: %v\n", err)
		}
	}()
}
