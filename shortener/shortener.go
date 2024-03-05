package shortener

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	shortenerTimeout  = 10 * time.Second
	shortenerBase     = "https://frame.vote/"
	shortenerEndpoint = shortenerBase + "add/%s"
)

// ShortURL returns a shortened version of the provided URL. It uses the
// vocdoni shortener service to shorten the URL. It returns the shortened URL or
// an error if something went wrong. It uses a timeout of 10 seconds for the
// request.
func ShortURL(ctx context.Context, electionURL string) (string, error) {
	// remove the protocol from the URL provided
	parsedURL, err := url.Parse(electionURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	urlToShort := parsedURL.Hostname()
	if port := parsedURL.Port(); port != "" {
		urlToShort += ":" + port
	}
	urlToShort += parsedURL.RequestURI()

	internalCtx, cancel := context.WithTimeout(ctx, shortenerTimeout)
	defer cancel()
	endpointURL := fmt.Sprintf(shortenerEndpoint, urlToShort)
	req, err := http.NewRequestWithContext(internalCtx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading json: %s", res.Status)
	}
	shortenerResponse := &struct {
		Link string `json:"link"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(shortenerResponse); err != nil {
		return "", fmt.Errorf("failed to decode json: %w", err)
	}
	return shortenerBase + shortenerResponse.Link, nil
}
