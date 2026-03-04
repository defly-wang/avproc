package ffmpeg

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Convert(input, output, quality, resolution string, onProgress func(Progress)) error {
	info, err := GetMediaInfo(input)
	if err != nil {
		return err
	}
	duration := info.DurationSec

	args := []string{"-i", input, "-y"}

	ext := ""
	if len(output) >= 4 {
		ext = output[len(output)-4:]
		if ext[0] == '.' {
			ext = ext[1:]
		} else if len(output) >= 5 && output[len(output)-5] == '.' {
			ext = output[len(output)-4:]
		}
	}

	audioFormats := map[string]bool{
		"mp3": true, "wav": true, "aac": true, "ogg": true,
		"flac": true, "m4a": true, "wma": true,
	}

	if audioFormats[ext] {
		switch ext {
		case "mp3":
			args = append(args, "-vn", "-acodec", "libmp3lame", "-ab", "192k")
		case "wav":
			args = append(args, "-vn", "-acodec", "pcm_s16le")
		case "aac":
			args = append(args, "-vn", "-acodec", "aac", "-ab", "192k")
		case "ogg":
			args = append(args, "-vn", "-acodec", "libvorbis", "-ab", "192k")
		case "flac":
			args = append(args, "-vn", "-acodec", "flac")
		case "m4a":
			args = append(args, "-vn", "-acodec", "aac", "-ab", "192k")
		case "wma":
			args = append(args, "-vn", "-acodec", "wmav2", "-ab", "192k")
		}
	} else {
		if resolution != "" && resolution != "original" {
			args = append(args, "-vf", "scale="+resolution)
		}

		switch quality {
		case "high":
			args = append(args, "-crf", "18")
		case "medium":
			args = append(args, "-crf", "23")
		case "low":
			args = append(args, "-crf", "28")
		}
	}

	args = append(args, "-progress", "pipe:1", output)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd.Stdout = pw

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
					currentTime := timeMs / 1000000
					percent := 0.0
					if duration > 0 {
						percent = currentTime / duration * 100
					}
					onProgress(Progress{
						Percent: percent,
						Time:    currentTime,
					})
				}
			}
		}
	}()

	return cmd.Run()
}
