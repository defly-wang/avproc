package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
