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

	execDir := ""
	if len(os.Args) > 0 {
		execDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}

	if execDir != "" {
		localFFmpeg := filepath.Join(execDir, "ffmpeg"+ext)
		if info, err := os.Stat(localFFmpeg); err == nil && !info.IsDir() {
			return localFFmpeg
		}

		localFFmpeg = filepath.Join(execDir, "ffmpeg")
		if info, err := os.Stat(localFFmpeg); err == nil && !info.IsDir() {
			return localFFmpeg
		}
	}

	if path, err := exec.LookPath("ffmpeg" + ext); err == nil {
		return path
	}

	if path, err := exec.LookPath("ffmpeg"); err == nil {
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

	ext := ""
	if os.PathSeparator == '\\' {
		ext = ".exe"
	}
	ffprobePath := filepath.Join(filepath.Dir(ffmpegPath), "ffprobe"+ext)
	if _, err := os.Stat(ffprobePath); err != nil {
		ffprobePath = "ffprobe"
		if path, err := exec.LookPath("ffprobe"); err == nil {
			ffprobePath = path
		} else {
			return fmt.Errorf("ffprobe not found. Please install ffprobe")
		}
	}

	cmd = exec.Command(ffprobePath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install ffprobe")
	}
	return nil
}
