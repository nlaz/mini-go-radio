package main

import (
	"time"
)

type Throttle struct {
	bytesPerSecond int
	input          chan []byte
	output         chan []byte
}

func NewThrottle(bytesPerSecond int) *Throttle {
	t := &Throttle{
		bytesPerSecond: bytesPerSecond,
		input:          make(chan []byte, 100),
		output:         make(chan []byte, 100),
	}
	go t.run()
	return t
}

func (t *Throttle) run() {
	var buffer []byte
	start := time.Now()
	totalBytes := 0

	for {
		var chunk []byte
		var ok bool

		select {
		case chunk, ok = <-t.input:
			if !ok {
				close(t.output)
				return
			}
			buffer = append(buffer, chunk...)
		default:
		}

		if len(buffer) > 0 {
			elapsed := time.Since(start).Seconds()
			expectedBytes := int(elapsed * float64(t.bytesPerSecond))
			availableBytes := expectedBytes - totalBytes

			if availableBytes > 0 {
				bytesToSend := availableBytes
				if bytesToSend > len(buffer) {
					bytesToSend = len(buffer)
				}

				t.output <- buffer[:bytesToSend]
				buffer = buffer[bytesToSend:]
				totalBytes += bytesToSend
			}

			if availableBytes <= 0 {
				sleepDuration := time.Duration((float64(-availableBytes) / float64(t.bytesPerSecond)) * float64(time.Second))
				time.Sleep(sleepDuration)
			}
		} else {
			time.Sleep(time.Millisecond)
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