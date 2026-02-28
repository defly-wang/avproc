package ffmpeg

import (
	"fmt"
	"os/exec"
)

func CheckFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg")
	}
	cmd = exec.Command("ffprobe", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install ffprobe")
	}
	return nil
}
