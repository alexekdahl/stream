package stream

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/alexekdahl/stream/auth"
)

type Config struct {
	URL         string
	Username    string
	Password    string
	ClientNonce string
}

type Streamer struct {
	reader *multipart.Reader
	stream io.ReadCloser
}

func (s *Streamer) NextPart() (io.Reader, error) {
	part, err := s.reader.NextPart()
	if err == io.EOF {
		return nil, fmt.Errorf("no more parts")
	}

	if err != nil {
		return nil, err
	}

	return part, nil
}

func newStreamer(stream io.ReadCloser, boundary string) *Streamer {
	return &Streamer{
		reader: multipart.NewReader(stream, boundary),
		stream: stream,
	}
}

func (s *Streamer) Close() error {
	return s.stream.Close()
}

func FetchVideoStream(client *http.Client, config Config) (*Streamer, error) {
	req, err := http.NewRequest("GET", config.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		authedReq, err := auth.Authenticate(
			client,
			resp,
			config.URL,
			config.Username,
			config.Password,
			config.ClientNonce,
		)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}

		resp, err = client.Do(authedReq)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/x-mixed-replace") {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	parts := strings.Split(contentType, "boundary=")
	if len(parts) != 2 {
		resp.Body.Close()
		return nil, fmt.Errorf("invalid content type")
	}

	return newStreamer(resp.Body, parts[1]), nil
}
