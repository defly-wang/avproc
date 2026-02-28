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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
		OpenPlayerWindow(currentPath, info)
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
			DisplayInfo(info, infoLabel)
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
