package auth

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

func Authenticate(
	client *http.Client,
	resp *http.Response,
	url, username, password, cnonce string,
) (*http.Request, error) {
	authHeader := resp.Header.Get("WWW-Authenticate")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Digest" {
		return nil, fmt.Errorf("unsupported authentication mechanism")
	}
	authParams := parseParams(parts[1])

	realm := authParams["realm"]
	nonce := authParams["nonce"]
	uri := url
	qop := authParams["qop"]
	nc := "00000001"

	response := calculateResponse(username, password, realm, nonce, uri, qop, cnonce, nc)

	authHeader = fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", qop=%s, nc=%s, cnonce="%s", response="%s"`,
		username,
		realm,
		nonce,
		uri,
		qop,
		nc,
		cnonce,
		response,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	return req, nil
}

func parseParams(s string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(s, ", ")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.Trim(kv[0], `"`)
			value := strings.Trim(kv[1], `"`)
			params[key] = value
		}
	}
	return params
}

func calculateResponse(username, password, realm, nonce, uri, qop, cnonce, nc string) string {
	HA1 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password))))
	HA2 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("GET:%s", uri))))
	response := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", HA1, nonce, nc, cnonce, qop, HA2))))
	return response
}
