package main

import (
	"fmt"
	"io"
	"os/exec"
)

type FFMPEG struct {
	cmd *exec.Cmd
	done chan struct{}
	broadcast io.Writer
}

func NewFFMPEG(input string, broadcast io.Writer) *FFMPEG {
	args := []string{
		"-re",
		"-y",
		"-loglevel", "error",
		"-i", input,
		"-ac", "2",
		"-b:a", fmt.Sprintf("%d", bitrate),
		"-ar", "48000",
		"-vn",
		"-f", "mp3",
		"pipe:1",
	}

	return &FFMPEG{
		cmd: exec.Command("ffmpeg", args...),
		done: make(chan struct{}),
		broadcast: broadcast,
	}
}

func (f *FFMPEG) Start() error {
	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	if err := f.cmd.Start(); err != nil {
		return fmt.Errorf("error starting ffmpeg; %w", err)
	}

	go func() {
		io.Copy(f.broadcast, stdout)
		close(f.done)
	}()

	return nil
}

func (f *FFMPEG) Done() <-chan struct{} {
	return f.done
}