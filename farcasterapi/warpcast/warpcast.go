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
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

const (
	maxRetries            = 12              // Maximum number of retries
	baseDelay             = 1 * time.Second // Initial delay, increases exponentially
	WarpcastApi           = "https://api.warpcast.com/v2"
	warpcastDirectMessage = WarpcastApi + "/ext-send-direct-cast"

	unauthorizedMessage = "Must provide Authorization header with API key"
)

type dmRequest struct {
	RecipientFID   uint64 `json:"recipientFid"`
	Message        string `json:"message"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type errResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// WarpcastAPI struct represents the Warpcast API client. It contains the user
// FID and the API key to authenticate the requests. It also contains a
// semaphore to limit the number of concurrent requests.
type WarpcastAPI struct {
	userFID      uint64
	apiKey       string
	reqSemaphore chan struct{} // Semaphore to limit concurrent requests
}

// NewWarpcastAPI method creates a new WarpcastAPI instance and returns it.
func NewWarpcastAPI() *WarpcastAPI {
	return &WarpcastAPI{
		reqSemaphore: make(chan struct{}, 10),
	}
}

func (w *WarpcastAPI) SetFarcasterUser(fid uint64, apiKey string) error {
	w.userFID = fid
	w.apiKey = apiKey
	// check if the provided API key is valid
	valid, err := w.validApiKey()
	if err != nil {
		return fmt.Errorf("error validating API key: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid API key")
	}
	return nil
}

// DirectMessage method sends a direct message to the user with the given fid.
// If something goes wrong, it returns an error.
func (w *WarpcastAPI) DirectMessage(ctx context.Context, content string, to uint64) error {
	if w.apiKey == "" {
		return fmt.Errorf("missing API key")
	}
	body, err := json.Marshal(dmRequest{
		RecipientFID:   to,
		Message:        content,
		IdempotencyKey: uuid.New().String(),
	})
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

// validApiKey method checks if the API key is valid by sending a request to the
// warpcastDirectMessage endpoint. If the API key is invalid, it returns false
// and an error if something goes wrong. The API key is valid if the response
// status code is 400 and the error message is different from the
// unauthorizedMessage. Since the Warpcast API does not provide an endpoint to
// check the validity of the API key, we need to use a different endpoint to
// check it with an expected error.
func (w *WarpcastAPI) validApiKey() (bool, error) {
	if w.apiKey == "" {
		return false, fmt.Errorf("missing API key")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// create a request to send a direct message but with empty body to raise an
	// expected error
	body, err := json.Marshal(dmRequest{RecipientFID: 1})
	if err != nil {
		return false, fmt.Errorf("error marshalling request body: %w", err)
	}
	req, err := w.prepareRequest(ctx, http.MethodPut, warpcastDirectMessage, bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	// send the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error downloading json: %w", err)
	}
	defer res.Body.Close()
	// to check if the API key is valid, we need to check the status code of the
	// response, it must be 400 and the error message must be different from the
	// unauthorizedMessage
	switch res.StatusCode {
	case http.StatusOK:
		// if the status code is 200, something is wrong because the endpoint
		// should return a 400 status code in any case
		return false, fmt.Errorf("unexpected 200 status code")
	case http.StatusBadRequest:
		// read the response body
		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			return false, fmt.Errorf("error reading response body: %w", err)
		}
		// unmarshal the error response body
		var errResp errResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return false, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		if len(errResp.Errors) == 0 {
			return false, fmt.Errorf("unexpected error response")
		}
		// check if the error is due to an invalid API key, if not, the API key
		// is valid
		if errResp.Errors[0].Message != unauthorizedMessage {
			return true, nil
		}
		return false, nil
	}
	// if the status error code is different from 400, return the error
	return false, fmt.Errorf("error downloading json: %s", res.Status)
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
	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}
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
			log.Info("Response status code: ", res.StatusCode)
			respBody, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			log.Info("Response body: ", string(respBody))
			return respBody, nil // Success
		}
	}
	return nil, fmt.Errorf("error downloading json: exceeded retry limit")
}
