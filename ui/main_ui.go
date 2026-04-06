package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func loadIcon(name string) fyne.Resource {
	switch name {
	case "preview":
		return theme.MediaPlayIcon()
	case "convert":
		return theme.ViewRefreshIcon()
	case "crop":
		return theme.ZoomFitIcon()
	case "merge":
		return theme.FileIcon()
	case "screenshot":
		return theme.DocumentIcon()
	default:
		return theme.DocumentIcon()
	}
}

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
	screenshot := NewScreenshotTab(window)

	ui.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("预览", loadIcon("preview"), preview),
		container.NewTabItemWithIcon("转换", loadIcon("convert"), convert),
		container.NewTabItemWithIcon("剪裁", loadIcon("crop"), crop),
		container.NewTabItemWithIcon("截图", loadIcon("screenshot"), screenshot),
		container.NewTabItemWithIcon("拼接", loadIcon("merge"), merge),
	)

	ui.tabs.SetTabLocation(container.TabLocationTop)

	return ui.tabs
}
