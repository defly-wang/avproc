package ui

import (
	"avproc/ffmpeg"
	"fmt"

	"fyne.io/fyne/v2/widget"
)

func DisplayInfo(info *ffmpeg.MediaInfo, label *widget.Label) {
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

func UpdateTimeLabel(start, end, duration float64, label *widget.Label) {
	startStr := FormatTimeSec(start)
	endStr := FormatTimeSec(end)
	label.SetText(fmt.Sprintf("开始: %s  结束: %s", startStr, endStr))
}

func FormatTimeSec(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", m/60, m%60, s)
}
