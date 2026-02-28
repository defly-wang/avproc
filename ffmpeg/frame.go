package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
)

func ExtractFrame(path string, timestamp float64) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "frame_*")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpDir + "/frame.jpg"

	args := []string{
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", path,
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		tmpPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	data, err := os.ReadFile(tmpPath)
	os.RemoveAll(tmpDir)
	return data, err
}

func Play(path string) error {
	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func PlayWithWindow(path string) error {
	cmd := exec.Command("ffplay", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
