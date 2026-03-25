package utils

import (
	"fmt"
	"sync"
	"time"
)

var spinChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Progress struct {
	total     int
	completed int
	current   string
	mu        sync.Mutex
	done      chan struct{}
	stopped   bool
}

func NewProgress(total int) *Progress {
	return &Progress{
		total: total,
		done:  make(chan struct{}),
	}
}

func (p *Progress) Start() {
	go p.spin()
}

func (p *Progress) Update(completed int, current string) {
	p.mu.Lock()
	p.completed = completed
	p.current = current
	p.mu.Unlock()
}

func (p *Progress) Increment(current string) {
	p.mu.Lock()
	p.completed++
	p.current = current
	p.mu.Unlock()
}

func (p *Progress) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.stopped = true
	close(p.done)
	fmt.Printf("\r\033[K")
}

func (p *Progress) spin() {
	i := 0
	for {
		select {
		case <-p.done:
			return
		case <-time.After(100 * time.Millisecond):
			p.mu.Lock()
			completed := p.completed
			total := p.total
			current := p.current
			p.mu.Unlock()

			spin := spinChars[i%len(spinChars)]
			i++

			if total > 0 {
				pct := completed * 100 / total
				name := TruncateString(current, 40)
				fmt.Printf("\r  %s %d/%d (%d%%) %s", spin, completed, total, pct, Dim(name))
			} else {
				name := TruncateString(current, 50)
				fmt.Printf("\r  %s %s", spin, Dim(name))
			}
		}
	}
}
