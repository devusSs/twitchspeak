package httplib

import (
	"encoding/json"
	"net/http"
	"time"
)

// Gets response from url by method (with token) as json
// and unmarshals it to v
func AuthorizedGet(method string, url string, token string, v interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

var (
	client = &http.Client{Timeout: 10 * time.Second}
)
