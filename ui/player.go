package ui

import (
	"avproc/ffmpeg"
	"fmt"
	"os"
	"os/exec"
)

func OpenPlayerWindow(path string, info *ffmpeg.MediaInfo) {
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
