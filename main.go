package main

import (
	"avproc/ffmpeg"
	"avproc/ui"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
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

	iconPath := "icon.png"
	if _, err := os.Stat(iconPath); err == nil {
		icon, err := fyne.LoadResourceFromPath(iconPath)
		if err == nil {
			a.SetIcon(icon)
		}
	} else {
		a.SetIcon(theme.MediaVideoIcon())
	}

	w := a.NewWindow("AVProc - 音视频处理工具")
	w.Resize(fyne.NewSize(900, 650))

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	content := ui.NewMainUI(w)
	w.SetContent(content)

	if desk, ok := a.(desktop.App); ok {
		fyne.DoAndWait(func() {
			desk.SetSystemTrayWindow(w)
		})

		menu := fyne.NewMenu("AVProc",
			fyne.NewMenuItem("显示", func() {
				fyne.DoAndWait(func() {
					w.Show()
					w.RequestFocus()
				})
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("退出", func() {
				a.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)

		if _, err := os.Stat(iconPath); err == nil {
			icon, err := fyne.LoadResourceFromPath(iconPath)
			if err == nil {
				desk.SetSystemTrayIcon(icon)
			}
		} else {
			desk.SetSystemTrayIcon(theme.MediaVideoIcon())
		}
	}

	log.Println("应用已启动，点击关闭按钮最小化到系统托盘")
	w.ShowAndRun()
}
