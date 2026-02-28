package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Merge(files []string, output string, onProgress func(Progress)) error {
	tmpFile, err := os.CreateTemp("", "concat_*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	for _, f := range files {
		fmt.Fprintf(tmpFile, "file '%s'\n", f)
	}
	tmpFile.Close()

	info, err := GetMediaInfo(files[0])
	if err != nil {
		return err
	}
	totalDuration := info.DurationSec * float64(len(files))

	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", tmpFile.Name(),
		"-c", "copy",
		"-y",
		output,
	)
	cmd.Stderr = os.Stderr

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout = pw

	processed := 0.0

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
					currentTime := timeMs/1000000 + float64(processed)
					percent := currentTime / totalDuration * 100
					onProgress(Progress{
						Percent: percent,
						Time:    currentTime,
					})
				}
			}
		}
	}()

	err = cmd.Run()
	if err != nil {
		return err
	}

	processed += info.DurationSec
	for i := 1; i < len(files); i++ {
		info, err := GetMediaInfo(files[i])
		if err != nil {
			continue
		}
		processed += info.DurationSec
		percent := processed / totalDuration * 100
		onProgress(Progress{
			Percent: percent,
			Time:    processed,
		})
	}

	return nil
}
