package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Crop(input, output, startTime, endTime string, onProgress func(Progress)) error {
	args := []string{
		"-ss", startTime,
		"-i", input,
		"-t", calculateDuration(startTime, endTime),
		"-c", "copy",
		"-y",
		output,
	}

	cmd := exec.Command(FindFFmpeg(), args...)
	cmd.Stderr = os.Stderr

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout = pw

	info, err := GetMediaInfo(input)
	if err != nil {
		return err
	}
	duration := info.DurationSec

	go func() {
		defer pr.Close()
		buf := make([]byte, 1024)
		for {
			n, err := pr.Read(buf)
			if err != nil {
				break
			}
			out := string(buf[:n])
			lines := strings.Split(out, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "out_time_ms=") {
					timeStr := strings.TrimPrefix(line, "out_time_ms=")
					timeMs, _ := strconv.ParseFloat(timeStr, 64)
					percent := (timeMs / 1000000) / duration * 100
					onProgress(Progress{
						Percent: percent,
						Time:    timeMs / 1000000,
					})
				}
			}
		}
	}()

	return cmd.Run()
}

func calculateDuration(start, end string) string {
	startSec := parseTimeToSeconds(start)
	endSec := parseTimeToSeconds(end)
	duration := endSec - startSec
	if duration < 0 {
		duration = 0
	}
	return fmt.Sprintf("%.0f", duration)
}

func parseTimeToSeconds(t string) float64 {
	parts := strings.Split(t, ":")
	var sec float64
	if len(parts) == 3 {
		h, _ := strconv.ParseFloat(parts[0], 64)
		m, _ := strconv.ParseFloat(parts[1], 64)
		s, _ := strconv.ParseFloat(parts[2], 64)
		sec = h*3600 + m*60 + s
	} else if len(parts) == 2 {
		m, _ := strconv.ParseFloat(parts[0], 64)
		s, _ := strconv.ParseFloat(parts[1], 64)
		sec = m*60 + s
	}
	return sec
}
