package ui

import (
	"avproc/ffmpeg"
	"bytes"
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewPreviewTab(window fyne.Window) fyne.Widget {
	var info *ffmpeg.MediaInfo
	var currentPath string

	pathLabel := widget.NewLabel("未选择文件")
	infoLabel := widget.NewLabel("")
	//infoLabel.Wrapping = fyne.TextWrapWord

	previewImage := canvas.NewImageFromResource(nil)
	previewImage.FillMode = canvas.ImageFillContain
	previewImage.SetMinSize(fyne.NewSize(320, 180))

	loadingLabel := widget.NewLabel("")

	playBtn := widget.NewButtonWithIcon("播放", theme.MediaPlayIcon(), func() {
		if currentPath == "" {
			infoLabel.SetText("请先选择文件")
			return
		}
		OpenPlayerWindow(currentPath, info)
	})
	playBtn.Disable()

	selectBtn := widget.NewButtonWithIcon("打开", theme.FolderOpenIcon(), func() {
		filter := storage.NewExtensionFileFilter([]string{".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv", ".webm", ".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a"})
		fd := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
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
			DisplayInfo(info, infoLabel)
			if len(info.VideoTracks) > 0 || len(info.AudioTracks) > 0 {
				playBtn.Enable()
			}

			loadingLabel.SetText("正在加载预览...")
			go func() {
				data, err := ffmpeg.ExtractFrame(path, 1.0)
				if err != nil {
					fyne.Do(func() {
						loadingLabel.SetText("")
					})
					return
				}
				img, _, err := image.Decode(bytes.NewReader(data))
				if err != nil {
					fyne.Do(func() {
						loadingLabel.SetText("")
					})
					return
				}
				rgba := image.NewRGBA(img.Bounds())
				for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
					for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
						rgba.Set(x, y, img.At(x, y))
					}
				}
				fyne.Do(func() {
					previewImage.Image = rgba
					previewImage.Refresh()
					loadingLabel.SetText("")
				})
			}()
		}, window)
		fd.SetFilter(filter)
		fd.Show()
	})

	toolbar := container.NewHBox(
		selectBtn,
		playBtn,
	)

	previewContainer := container.NewVBox(
		previewImage,
		loadingLabel,
	)

	infoContainer := container.NewVBox(
		widget.NewLabel("文件信息"),
		infoLabel,
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
				previewContainer,
				layout.NewSpacer(),
				infoContainer,
				layout.NewSpacer(),
			),
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}
