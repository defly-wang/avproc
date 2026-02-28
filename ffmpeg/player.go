package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
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
