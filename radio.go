package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	port = 3000
	bitrate = 196 * 1024
)

var (
	idCounter int
	idMutex sync.Mutex
)

func generateID() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return idCounter
}

type Radio struct {
	broadcast *Broadcast
}

func NewRadio() *Radio {
	return &Radio{
		broadcast: NewBroadcast(),
	}
}

func (r *Radio) Run() {
	for {
		currentTrack := r.selectRandomTrack()
		input := filepath.Join("library", currentTrack)
		fmt.Printf("Now playing: %s\n", currentTrack)

		ffmpeg := NewFFMPEG(input, r.broadcast)
		if err := ffmpeg.Start(); err != nil {
			log.Printf("Error starting ffmpeg: %v", err)
			continue
		}

		<-ffmpeg.Done()
	}
}

func (f *Radio) selectRandomTrack() string {
	files, err := os.ReadDir("library")
	if err != nil {
		log.Fatal(err)
	}
	return files[rand.Intn(len(files))].Name()
}

func main() {
	radio := NewRadio()
	go radio.Run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s, Accept: %s", r.Method, r.Header.Get("Accept"))

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, HEAD")
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Connection", "close")
		w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")

		id := generateID()
		stream := radio.broadcast.Subscribe(id)
		defer radio.broadcast.Unsubscribe(id)

		for chunk := range stream {
			_, err := w.Write(chunk)
			if err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	})

	log.Printf("Server running on http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}