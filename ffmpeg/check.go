package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func FindFFmpeg() string {
	ext := ""
	if os.PathSeparator == '\\' {
		ext = ".exe"
	}

	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	localFFmpeg := filepath.Join(execDir, "ffmpeg"+ext)
	if _, err := os.Stat(localFFmpeg); err == nil {
		return localFFmpeg
	}

	path, err := exec.LookPath("ffmpeg" + ext)
	if err == nil {
		return path
	}

	return ""
}

func CheckFFmpeg() error {
	ffmpegPath := FindFFmpeg()
	if ffmpegPath == "" {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg")
	}

	cmd := exec.Command(ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg")
	}

	ffprobePath := filepath.Join(filepath.Dir(ffmpegPath), "ffprobe.exe")
	if _, err := os.Stat(ffprobePath); err != nil {
		return fmt.Errorf("ffprobe not found. Please install ffprobe")
	}

	cmd = exec.Command(ffprobePath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install ffprobe")
	}
	return nil
}
