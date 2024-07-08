package stream

import (
	"net/http"

	"github.com/alexekdahl/stream/auth"
)

type Config struct {
	URL         string
	Username    string
	Password    string
	ClientNonce string
}

func FetchStream(client *http.Client, config Config) (*http.Response, error) {
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
			return nil, err
		}

		resp, err = client.Do(authedReq)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
