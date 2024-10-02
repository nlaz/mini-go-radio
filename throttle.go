package main

import (
	"time"
)

type Throttle struct {
	bytesPerSecond int
	chunkSize int
	input chan []byte
	output chan []byte
}

func NewThrottle(bytesPerSecond int) *Throttle {
	t := &Throttle{
		bytesPerSecond: bytesPerSecond,
		chunkSize: bytesPerSecond / 10,
		input: make(chan []byte, 100),
		output: make(chan []byte, 100),
	}
	go t.run()
	return t
}

func (t *Throttle) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var buffer []byte
	start := time.Now()
	totalBytes := 0

	for {
		select {
		case chunk, ok := <-t.input:
			if !ok {
				close(t.output)
				return
			}
			buffer = append(buffer, chunk...)
		case <-ticker.C:
			if len(buffer) > 0 {
				bytesToSend := t.chunkSize
				if bytesToSend > len(buffer) {
					bytesToSend = len(buffer)
				}

				t.output <- buffer[:bytesToSend]
				buffer = buffer[bytesToSend:]
				totalBytes += bytesToSend

				elapsed := time.Since(start).Seconds()
				expectedBytes := int(elapsed * float64(t.bytesPerSecond))

				if totalBytes > expectedBytes {
					time.Sleep(time.Duration(float64(totalBytes-expectedBytes)/float64(t.bytesPerSecond)*100) * time.Millisecond)
				}
			}
		}
	}
}

func (t *Throttle) Write(p []byte) (n int, err error) {
	t.input <- p
	return len(p), nil
}

func (t *Throttle) Output() <-chan []byte {
	return t.output
}