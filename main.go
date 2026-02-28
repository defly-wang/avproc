package main

import (
	"avproc/ffmpeg"
	"avproc/ui"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	if err := ffmpeg.CheckFFmpeg(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		fmt.Println("Please install FFmpeg first.")
		os.Exit(1)
	}

	a := app.NewWithID("com.avproc.app")
	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("AVProc - 音视频处理工具")
	w.Resize(fyne.NewSize(900, 650))

	content := ui.NewMainUI(w)
	w.SetContent(content)

	w.ShowAndRun()
}
