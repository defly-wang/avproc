package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GetMediaInfo(path string) (*MediaInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to probe file: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	info := &MediaInfo{
		Filename:    filepath.Base(path),
		VideoTracks: []VideoTrack{},
		AudioTracks: []AudioTrack{},
	}

	if format, ok := result["format"].(map[string]interface{}); ok {
		if v, ok := format["format_name"].(string); ok {
			info.Format = v
		}
		if v, ok := format["duration"].(string); ok {
			info.Duration = formatDuration(v)
			info.DurationSec, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := format["size"].(string); ok {
			info.Size = formatFileSize(v)
		}
		if v, ok := format["bit_rate"].(string); ok {
			info.Bitrate = formatBitrate(v)
		}
	}

	if streams, ok := result["streams"].([]interface{}); ok {
		for _, s := range streams {
			if stream, ok := s.(map[string]interface{}); ok {
				codecType, _ := stream["codec_type"].(string)
				if codecType == "video" {
					vt := VideoTrack{
						Codec: getString(stream["codec_name"]),
					}
					if w, ok := stream["width"].(float64); ok {
						vt.Width = int(w)
					}
					if h, ok := stream["height"].(float64); ok {
						vt.Height = int(h)
					}
					if r, ok := stream["r_frame_rate"].(string); ok {
						vt.FrameRate = parseFrameRate(r)
					}
					if b, ok := stream["bit_rate"].(string); ok {
						vt.Bitrate = formatBitrate(b)
					}
					info.VideoTracks = append(info.VideoTracks, vt)
				} else if codecType == "audio" {
					at := AudioTrack{
						Codec:      getString(stream["codec_name"]),
						SampleRate: getString(stream["sample_rate"]),
					}
					if c, ok := stream["channels"].(float64); ok {
						at.Channels = int(c)
					}
					if b, ok := stream["bit_rate"].(string); ok {
						at.Bitrate = formatBitrate(b)
					}
					info.AudioTracks = append(info.AudioTracks, at)
				}
			}
		}
	}

	return info, nil
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func formatDuration(d string) string {
	sec, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return d
	}
	duration := time.Duration(sec * float64(time.Second))
	h := duration / time.Hour
	m := (duration % time.Hour) / time.Minute
	s := (duration % time.Minute) / time.Second
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatFileSize(s string) string {
	size, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	if size < 1024 {
		return fmt.Sprintf("%.0f B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", size/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", size/1024/1024)
	}
	return fmt.Sprintf("%.2f GB", size/1024/1024/1024)
}

func formatBitrate(b string) string {
	rate, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return b
	}
	if rate < 1000 {
		return fmt.Sprintf("%.0f bps", rate)
	}
	if rate < 1000000 {
		return fmt.Sprintf("%.0f Kbps", rate/1000)
	}
	return fmt.Sprintf("%.1f Mbps", rate/1000000)
}

func parseFrameRate(r string) string {
	parts := strings.Split(r, "/")
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		if den > 0 {
			return fmt.Sprintf("%.2f fps", num/den)
		}
	}
	return r
}
