package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/alexekdahl/stream/imageproc"
	"github.com/alexekdahl/stream/stream"
)

const (
	// streamURL   = "http://192.168.0.6/axis-cgi/mjpg/video.cgi"
	streamURL   = "http://172.29.166.189/axis-cgi/mjpg/video.cgi?camera=2"
	username    = "root"
	password    = "pass"
	clientNonce = "abcdef"
)

const (
	MoveCursorOrigin   = "\x1b[1;1H"
	EraseEntireDisplay = "\x1b[2J"
	HideCursor         = "\x1b[?25l"
)

func main() {
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

	resp, err := stream.FetchStream(client, config)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/x-mixed-replace") {
		log.Fatal(fmt.Errorf("unexpected content type: %s", contentType))
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		os.Stdout.Write([]byte("\x1b[?25h\x1b[2J"))
		os.Exit(0)
	}()

	fmt.Print(HideCursor)
	fmt.Print(EraseEntireDisplay)
	if err := imageproc.ProcessStream(resp.Body, contentType); err != nil {
		log.Fatal(err)
	}
}
