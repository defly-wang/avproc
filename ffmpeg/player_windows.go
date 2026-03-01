//go:build windows

package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

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
				time.Sleep(100 * time.Millisecond)
				continue
			}

			onTime(current)

			p.mu.Lock()
			p.current += frameInterval
			if p.current >= p.duration {
				p.current = p.duration
			}
			p.mu.Unlock()

			time.Sleep(time.Duration(frameInterval * float64(time.Second)))
		}
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

func (p *Player) Stop() {
	close(p.stopped)
}

func (p *Player) Seek(pos float64) {
	p.seekPos <- pos
}

func (p *Player) startAudio(pos float64, paused bool) {
	if p.audioCmd != nil && p.audioCmd.Process != nil {
		p.audioCmd.Process.Kill()
	}

	args := []string{"-i", p.path, "-ss", fmt.Sprintf("%.2f", pos), "-f", "wav", "-"}
	if paused {
		args = append([]string{"-i", p.path, "-ss", fmt.Sprintf("%.2f", pos), "-f", "wav", "-"}, "-nostdin")
	}

	p.audioCmd = exec.Command("ffplay", args...)
	p.audioCmd.Stdout = os.Stdout
	p.audioCmd.Stderr = os.Stderr

	if err := p.audioCmd.Start(); err != nil {
		fmt.Printf("Error starting audio: %v\n", err)
		return
	}
}
