package warpcast

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.vocdoni.io/dvote/util"
)

const (
	maxRetries            = 12              // Maximum number of retries
	baseDelay             = 1 * time.Second // Initial delay, increases exponentially
	WarpcastApi           = "https://api.warpcast.com/v2"
	warpcastDirectMessage = WarpcastApi + "/ext-send-direct-cast"
)

type WarpcastAPI struct {
	userFID      uint64
	apiKey       string
	reqSemaphore chan struct{} // Semaphore to limit concurrent requests
}

func (w *WarpcastAPI) SetFarcasterUser(fid uint64, apiKey string) error {
	w.userFID = fid
	w.apiKey = apiKey
	return nil
}

// DirectMessage method sends a direct message to the user with the given fid.
// If something goes wrong, it returns an error.
func (w *WarpcastAPI) DirectMessage(ctx context.Context, content string, to uint64) error {
	var reqBody = map[string]interface{}{
		"recipientFid":   to,
		"message":        content,
		"idempotencyKey": uuid.New().String(),
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %w", err)
	}
	req, err := w.prepareRequest(ctx, http.MethodPut, warpcastDirectMessage, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	if _, err := w.sendRequest(req); err != nil {
		return fmt.Errorf("error sending direct message: %w", err)
	}
	return nil
}

// prepareRequest method creates a new request with the given method, url and
// body. It sets the Authorization header with the client apiKey and the
// Content-Type header to the request and returns the request and an error if
// something goes wrong.
func (w *WarpcastAPI) prepareRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// sendRequest method sends the given request and returns the response body and
// an error if something goes wrong. It retries the request if it fails.
func (w *WarpcastAPI) sendRequest(req *http.Request) ([]byte, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		// We need to avoid too much concurrent requests and penalization from the API
		w.reqSemaphore <- struct{}{}
		res, err := http.DefaultClient.Do(req)
		<-w.reqSemaphore
		if err != nil {
			return nil, fmt.Errorf("error downloading json: %w", err)
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Duration(attempt+1)*baseDelay + time.Duration(util.RandomInt(0, 2000))*time.Millisecond)
		} else if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error downloading json: %s", res.Status)
		} else {
			respBody, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			return respBody, nil // Success
		}
	}
	return nil, fmt.Errorf("error downloading json: exceeded retry limit")
}
