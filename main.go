package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/alexekdahl/stream/imageproc"
	"github.com/alexekdahl/stream/stream"
)

const (
	// streamURL   = "http://192.168.0.6/axis-cgi/mjpg/video.cgi"
	streamURL = "http://172.29.166.189/axis-cgi/mjpg/video.cgi?camera=2"
	// streamURL   = "http://172.29.166.199/axis-cgi/mjpg/video.cgi?camera=2"
	username    = "root"
	password    = "pass"
	clientNonce = "abcdef"
)

const (
	MoveCursorOrigin   = "\x1b[1;1H"
	EraseEntireDisplay = "\x1b[2J"
	hideCursor         = "\x1b[?25l"
	showCursor         = "\x1b[?25h\x1b[2J"
)

var showTerminalcursor = []byte{0x1b, 0x5b, 0x3f, 0x32, 0x35, 0, 0x68}

func main() {
	go handleInterrupt()

	config := stream.Config{
		URL:         streamURL,
		Username:    username,
		Password:    password,
		ClientNonce: clientNonce,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	streamer, err := stream.FetchVideoStream(client, config)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	os.Stdout.Write([]byte(hideCursor))
	os.Stdout.Write([]byte(EraseEntireDisplay))

	if err := imageproc.ProcessVideoStream(streamer); err != nil {
		log.Fatal(err)
	}
}

func handleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	os.Stdout.Write(showTerminalcursor)
	os.Exit(0)
}
