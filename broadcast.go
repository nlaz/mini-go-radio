package main

import (
	"sync"
)

type Broadcast struct {
	subscribers map[int]chan []byte
	mu sync.RWMutex
	throttle *Throttle
}

func NewBroadcast() *Broadcast {
	b := &Broadcast{
		subscribers: make(map[int]chan []byte),
		throttle: NewThrottle(bitrate / 8),
	}
	go b.run()
	return b
}

func (b *Broadcast) run() {
	for chunk := range b.throttle.Output() {
		b.broadcast(chunk)
	}
}

func (b *Broadcast) Subscribe(id int) <-chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[id] = ch
	return ch
}

func (b *Broadcast) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.subscribers[id]; ok {
		close(ch)
		delete(b.subscribers, id)
	}
}

func (b *Broadcast) broadcast(chunk []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- chunk:
		default:
		}
	}
}

func (b *Broadcast) Write(p []byte) (n int, err error) {
	return b.throttle.Write(p)
}