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

type FrameCache struct {
	mu         sync.Mutex
	frames     map[int][]byte
	loading    bool
	path       string
	duration   float64
	preloadInt int
	step       int
}

func NewFrameCache() *FrameCache {
	return &FrameCache{
		frames: make(map[int][]byte),
	}
}

func (fc *FrameCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	for k := range fc.frames {
		delete(fc.frames, k)
	}
}

func (fc *FrameCache) Get(second int) []byte {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if data, ok := fc.frames[second]; ok {
		return data
	}
	if fc.step <= 1 {
		return nil
	}
	step := fc.step
	closest := ((second + step/2) / step) * step
	return fc.frames[closest]
}

func (fc *FrameCache) Set(second int, data []byte) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.frames[second] = data
}

func (fc *FrameCache) GetOrLoad(second int, loader func(int) []byte) []byte {
	fc.mu.Lock()
	if data, ok := fc.frames[second]; ok {
		fc.mu.Unlock()
		return data
	}
	if fc.step > 1 {
		step := fc.step
		closest := ((second + step/2) / step) * step
		if data, ok := fc.frames[closest]; ok {
			fc.mu.Unlock()
			return data
		}
	}
	fc.mu.Unlock()

	data := loader(second)
	if data != nil {
		fc.mu.Lock()
		fc.frames[second] = data
		fc.mu.Unlock()
	}
	return data
}

func (fc *FrameCache) Init(path string, dur float64, stepVal int) {
	fc.mu.Lock()
	fc.path = path
	fc.duration = dur
	fc.step = stepVal
	for k := range fc.frames {
		delete(fc.frames, k)
	}
	if dur > 300 {
		fc.preloadInt = 10
	} else if dur > 60 {
		fc.preloadInt = 5
	} else {
		fc.preloadInt = 2
	}
	fc.mu.Unlock()
}

func (fc *FrameCache) Preload(loader func(int) []byte, onDone func()) {
	fc.mu.Lock()
	if fc.loading || fc.path == "" {
		fc.mu.Unlock()
		return
	}
	fc.loading = true
	dur := fc.duration
	interval := fc.step
	if interval < 2 {
		interval = 2
	}
	fc.mu.Unlock()

	var wg sync.WaitGroup
	concurrency := 3

	secChan := make(chan int, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sec := range secChan {
				if fc.Get(sec) == nil {
					if data := loader(sec); data != nil {
						fc.Set(sec, data)
					}
				}
			}
		}()
	}

	go func() {
		for sec := 0; sec <= int(dur); sec += interval {
			secChan <- sec
		}
		close(secChan)
		wg.Wait()
		fc.mu.Lock()
		fc.loading = false
		fc.mu.Unlock()
		if onDone != nil {
			onDone()
		}
	}()
}

func NewCropTab(window fyne.Window) fyne.Widget {
	var progress float64
	var inputPath string
	var outputPath string
	var duration float64
	var frameCache *FrameCache

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

	displayImage := func(data []byte, imgView *canvas.Image) {
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
	}

	loadFrame := func(second int, imgView *canvas.Image) {
		currentPath := inputPath
		loader := func(sec int) []byte {
			if currentPath != inputPath {
				return nil
			}
			data, err := ffmpeg.ExtractFrame(currentPath, float64(sec))
			if err != nil {
				return nil
			}
			return data
		}

		data := frameCache.GetOrLoad(second, loader)
		if data != nil {
			displayImage(data, imgView)
		}
	}

	preloadFrames := func() {
		if frameCache == nil {
			return
		}

		currentPath := inputPath
		loader := func(sec int) []byte {
			if currentPath != inputPath {
				return nil
			}
			data, err := ffmpeg.ExtractFrame(currentPath, float64(sec))
			if err != nil {
				return nil
			}
			return data
		}

		loadingLabel.SetText("正在缓存预览图...")
		frameCache.Preload(loader, func() {
			fyne.Do(func() {
				loadingLabel.SetText("")
			})
		})
	}

	minSlider.OnChanged = func(value float64) {
		if value > maxSlider.Value {
			maxSlider.SetValue(value)
		}
		UpdateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 && frameCache != nil {
			second := int(value)
			if second >= 0 && second <= int(duration) {
				loadFrame(second, previewImageStart)
			}
		}
	}

	maxSlider.OnChanged = func(value float64) {
		if value < minSlider.Value {
			minSlider.SetValue(value)
		}
		UpdateTimeLabel(minSlider.Value, maxSlider.Value, duration, timeLabel)
		if inputPath != "" && duration > 0 && frameCache != nil {
			second := int(value)
			if second >= 0 && second <= int(duration) {
				loadFrame(second, previewImageEnd)
			}
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

			stepVal := 1
			if duration > 300 {
				stepVal = 10
			} else if duration > 60 {
				stepVal = 5
			}

			frameCache = NewFrameCache()
			frameCache.Init(inputPath, duration, stepVal)

			minSlider.Max = duration
			maxSlider.Max = duration
			if duration > 300 {
				minSlider.Step = 10
				maxSlider.Step = 10
			} else if duration > 60 {
				minSlider.Step = 5
				maxSlider.Step = 5
			} else {
				minSlider.Step = 1
				maxSlider.Step = 1
			}
			minSlider.SetValue(0)
			maxSlider.SetValue(duration)
			UpdateTimeLabel(0, duration, duration, timeLabel)

			preloadFrames()

			if duration > 1 {
				loadFrame(0, previewImageStart)
				loadFrame(int(duration)-1, previewImageEnd)
			}
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
