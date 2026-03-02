package ui

import (
	"avproc/ffmpeg"
	"bytes"
	"fmt"
	"image"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewMergeTab(window fyne.Window) fyne.Widget {
	var files []string
	var selectedIndex int = -1
	var outputPath string
	var progressBar *widget.ProgressBar
	var statusLabel *widget.Label

	progressBar = widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.Value = 0

	statusLabel = widget.NewLabel("")

	thumbnails := make(map[int]*image.RGBA)
	var mu sync.Mutex

	list := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject {
			img := canvas.NewImageFromResource(nil)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(120, 68))
			return container.NewHBox(
				img,
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			img := container.Objects[0].(*canvas.Image)
			label := container.Objects[1].(*widget.Label)

			label.SetText(files[id])

			mu.Lock()
			if thumb, ok := thumbnails[id]; ok {
				img.Image = thumb
				img.Refresh()
			}
			mu.Unlock()
		},
	)

	listScroll := container.NewScroll(list)
	listScroll.SetMinSize(fyne.NewSize(400, 250))

	var mergeBtn *widget.Button
	var previewBtn *widget.Button

	addFile := func(path string) {
		files = append(files, path)
		list.Refresh()
		mergeBtn.Enable()

		go func() {
			data, err := ffmpeg.ExtractFrame(path, 0.5)
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
			mu.Lock()
			thumbnails[len(files)-1] = rgba
			mu.Unlock()
			fyne.Do(func() {
				list.Refresh()
			})
		}()
	}

	addBtn := widget.NewButtonWithIcon("添加", theme.FolderOpenIcon(), func() {
		filter := storage.NewExtensionFileFilter([]string{".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv", ".webm"})
		fd := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("错误: %v", err))
				return
			}
			if closer == nil {
				return
			}
			addFile(closer.URI().Path())
		}, window)
		fd.SetFilter(filter)
		fd.Show()
	})

	removeBtn := widget.NewButtonWithIcon("移除", theme.DeleteIcon(), func() {
		if selectedIndex >= 0 && selectedIndex < len(files) {
			files = append(files[:selectedIndex], files[selectedIndex+1:]...)
			list.Refresh()
			selectedIndex = -1
		}
	})

	mergeBtn = widget.NewButtonWithIcon("拼接", theme.MediaRecordIcon(), func() {
		if len(files) < 2 {
			statusLabel.SetText("请至少添加2个文件")
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

			statusLabel.SetText("拼接中...")
			mergeBtn.Disable()

			go func() {
				err := ffmpeg.Merge(files, outputPath, func(p ffmpeg.Progress) {
					progress := p.Percent
					fyne.DoAndWait(func() {
						progressBar.SetValue(progress)
						statusLabel.SetText(fmt.Sprintf("拼接中... %.1f%%", progress))
					})
				})

				fyne.Do(func() {
					if err != nil {
						statusLabel.SetText(fmt.Sprintf("拼接失败: %v", err))
					} else {
						statusLabel.SetText("拼接完成!")
						progressBar.SetValue(100)
						previewBtn.Enable()
					}
					mergeBtn.Enable()
				})
			}()
		}, window)
	})
	mergeBtn.Disable()

	previewBtn = widget.NewButtonWithIcon("预览", theme.MediaPlayIcon(), func() {
		if outputPath != "" {
			OpenPlayerWindow(outputPath, nil)
		}
	})
	previewBtn.Disable()

	toolbar := container.NewHBox(
		addBtn,
		removeBtn,
		mergeBtn,
		previewBtn,
	)

	content := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewVBox(
			widget.NewLabel("文件列表:"),
			listScroll,
			widget.NewSeparator(),
			progressBar,
			statusLabel,
			layout.NewSpacer(),
		),
	)

	return container.NewScroll(content)
}
