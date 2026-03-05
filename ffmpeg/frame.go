package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ExtractFrame(path string, timestamp float64) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "frame_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "frame.jpg")

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", path,
		"-frames:v", "1",
		"-vf", "scale=240:-1",
		"-f", "image2",
		"-update", "1",
		tmpPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted frame: %w", err)
	}
	return data, nil
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
