package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type MediaInfo struct {
	Filename    string
	Format      string
	Duration    string
	DurationSec float64
	Size        string
	Bitrate     string
	VideoTracks []VideoTrack
	AudioTracks []AudioTrack
}

type VideoTrack struct {
	Codec     string
	Width     int
	Height    int
	FrameRate string
	Bitrate   string
}

type AudioTrack struct {
	Codec      string
	SampleRate string
	Channels   int
	Bitrate    string
}

type Progress struct {
	Percent    float64
	Time       float64
	Speed      string
	OutputSize string
}

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

func Convert(input, output, quality string, onProgress func(Progress)) error {
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
			output := string(buf[:n])
			lines := strings.Split(output, "\n")
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

func Crop(input, output, startTime, endTime string, onProgress func(Progress)) error {
	args := []string{
		"-ss", startTime,
		"-i", input,
		"-t", calculateDuration(startTime, endTime),
		"-c", "copy",
		"-y",
		output,
	}

	cmd := exec.Command("ffmpeg", args...)
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
			output := string(buf[:n])
			lines := strings.Split(output, "\n")
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
			output := string(buf[:n])
			lines := strings.Split(output, "\n")
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

type Player struct {
	path     string
	duration float64
	cmd      *exec.Cmd
	audioCmd *exec.Cmd
	stopped  chan struct{}
	paused   bool
	mu       sync.Mutex
	current  float64
	seekPos  chan float64
}

func NewPlayer(path string, duration float64) *Player {
	return &Player{
		path:     path,
		duration: duration,
		stopped:  make(chan struct{}),
		seekPos:  make(chan float64, 1),
	}
}

func (p *Player) Play(onTime func(float64)) {
	p.stopped = make(chan struct{})
	p.paused = false
	p.current = 0

	p.startAudio(p.current, p.paused)

	go func() {
		frameInterval := 1.0 / 5.0

		for {
			select {
			case <-p.stopped:
				if p.audioCmd != nil && p.audioCmd.Process != nil {
					p.audioCmd.Process.Kill()
				}
				return
			case newPos := <-p.seekPos:
				p.mu.Lock()
				wasPaused := p.paused
				p.current = newPos
				p.mu.Unlock()

				p.startAudio(newPos, wasPaused)
				onTime(newPos)
				continue
			default:
			}

			p.mu.Lock()
			paused := p.paused
			current := p.current
			p.mu.Unlock()

			if paused {
				if p.audioCmd != nil && p.audioCmd.Process != nil {
					p.audioCmd.Process.Signal(syscall.SIGSTOP)
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if p.audioCmd != nil && p.audioCmd.Process != nil {
				p.audioCmd.Process.Signal(syscall.SIGCONT)
			}

			onTime(current)

			p.mu.Lock()
			p.current += frameInterval
			if p.current >= p.duration {
				p.current = p.duration
				onTime(p.current)
				p.mu.Unlock()
				if p.audioCmd != nil && p.audioCmd.Process != nil {
					p.audioCmd.Process.Kill()
				}
				return
			}
			p.mu.Unlock()

			time.Sleep(time.Duration(frameInterval * float64(time.Second)))
		}
	}()
}

func (p *Player) startAudio(startPos float64, wasPaused bool) {
	if p.audioCmd != nil && p.audioCmd.Process != nil {
		p.audioCmd.Process.Kill()
		p.audioCmd = nil
	}

	if wasPaused {
		return
	}

	p.audioCmd = exec.Command("ffplay", "-volume", "100", "-ss", fmt.Sprintf("%.3f", startPos), "-autoexit", p.path)
	p.audioCmd.Stderr = os.Stderr
	if err := p.audioCmd.Start(); err != nil {
		fmt.Printf("Failed to start player: %v\n", err)
		return
	}

	go func() {
		p.audioCmd.Wait()
	}()
}

func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
}

func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = false
}

func (p *Player) Seek(pos float64) {
	select {
	case p.seekPos <- pos:
	default:
	}
}

func (p *Player) Stop() {
	close(p.stopped)
	if p.audioCmd != nil && p.audioCmd.Process != nil {
		p.audioCmd.Process.Kill()
	}
}
